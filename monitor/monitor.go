package monitor

import (
	"fmt"
	"github.com/fastdevops/monitorx/logger"
	"github.com/fastdevops/monitorx/tg_notify"
	"github.com/fastdevops/monitorx/utils"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// 通过 SSH 执行远程 jps 命令
func CheckServices(node utils.Node) ([]string, error) {
	key, err := os.ReadFile(node.SSHKeyFile) // 读取私钥文件
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	SshConfig := &ssh.ClientConfig{
		User: node.SSH_USER,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer), // 使用公钥认证
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// 建立 SSH 连接
	conn, err := ssh.Dial("tcp", node.SSH_HOST+":"+node.SSH_PORT, SshConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 执行 jps 命令
	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	output, err := session.CombinedOutput("jps")
	if err != nil {
		logger.Logger.Error("检测代码执行出错，请检查代码是否正常")
		return nil, err
	}
	logger.Logger.Info("集群检测成功")

	runningServices := string(output)
	var missingServices []string
	for _, service := range node.Services {
		if !strings.Contains(runningServices, service) {
			missingServices = append(missingServices, service)
		}
	}
	// error service list
	return missingServices, nil
}

// CheckNodeResources 通过 SSH 获取节点的 CPU、内存和磁盘使用情况
func CheckNodeResources(node utils.Node) (string, string, string, error) {
	key, err := os.ReadFile(node.SSHKeyFile) // 读取私钥文件
	if err != nil {
		return "", "", "", err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", "", "", err
	}

	SshConfig := &ssh.ClientConfig{
		User: node.SSH_USER,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer), // 使用公钥认证
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// 建立 SSH 连接
	conn, err := ssh.Dial("tcp", node.SSH_HOST+":"+node.SSH_PORT, SshConfig)
	if err != nil {
		return "", "", "", err
	}
	defer conn.Close()

	// 获取 CPU 使用情况
	cpuUsage, err := executeCommandOverSSH(conn, "top -bn1 | grep \"Cpu(s)\" | awk '{print $2 + $4}'")
	if err != nil {
		return "", "", "", fmt.Errorf("获取 CPU 使用情况失败: %v", err)
	}

	// 获取内存使用情况
	memUsage, err := executeCommandOverSSH(conn, "free -m | awk 'NR==2{printf \"%.2f\", $3*100/$2 }'")
	if err != nil {
		return "", "", "", fmt.Errorf("获取内存使用情况失败: %v", err)
	}

	// 获取磁盘使用情况
	diskUsage, err := executeCommandOverSSH(conn, "df -h --total | grep 'total' | awk '{print $5}'")
	if err != nil {
		return "", "", "", fmt.Errorf("获取磁盘使用情况失败: %v", err)
	}

	// 返回 CPU、内存、磁盘使用情况
	return cpuUsage, memUsage, diskUsage, err
}

// executeCommandOverSSH 在远程机器上执行指定命令
func executeCommandOverSSH(conn *ssh.Client, cmd string) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// RunCheck Start check task
func RunCheck(cf *utils.Config, checks int, HadoopStatus *utils.HadoopCluster) {
	for {
		StartCheck(cf, HadoopStatus)
		time.Sleep(time.Duration(checks) * time.Second)
	}
}

// Check
func StartCheck(cf *utils.Config, hadoopStatus *utils.HadoopCluster) {
	for _, node := range cf.Nodes {
		data := utils.Node{
			Name:       node.Name,
			Services:   node.Services,
			SSH_HOST:   node.SSH_HOST,
			SSH_USER:   node.SSH_USER,
			SSHKeyFile: node.SSHKeyFile,
			SSH_PORT:   node.SSH_PORT,
		}
		setStatus := ""
		missingServices, err := CheckServices(data)
		if err != nil {
			logger.Logger.Error("参数验证错误，请检查监控服务配置文件是否正确", zap.Error(err))
			fmt.Println("注意！ 大数据集群服务检查失败")
			return
		}

		// check error
		if len(missingServices) > 0 {
			hadoopStatus.SetStatus(node.Name, missingServices, "error")
			logger.Logger.Info("巡检完成，发现异常服务，开始发送告警信息")
			msgs := tg_notify.CreateAlertMessage(node.Name, missingServices)
			tg_notify.SendTGMessage(cf.Telegram.Token, cf.Telegram.ChatID, msgs, hadoopStatus, node.Name)
			return
		} else {
			hadoopStatus.SetStatus(node.Name, missingServices, "running")
			// check the project is Recovery
			if ret, ok := hadoopStatus.GetStatus(node.Name); !ok {
				logger.Logger.Error("获取节点状态信息失败，节点信息未保存")
				return
			} else if ret.RecoveryTime != "" {
				logger.Logger.Info("巡检完成，发现异常服务后恢复，开始发送恢复告警信息")
				msgs := tg_notify.CreateAlertMessage(node.Name, missingServices)
				tg_notify.SendTGMessage(cf.Telegram.Token, cf.Telegram.ChatID, msgs, hadoopStatus, node.Name)
				setStatus = "recovery"
			}
			hadoopStatus.SetStatus(node.Name, missingServices, setStatus)
			time.Sleep(time.Duration(3) * time.Second)
		}
	}
}
