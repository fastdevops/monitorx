package utils

import (
	"strings"
	"time"
)

// check status
type HadoopCluster struct {
	Nodes map[string]ResponseStatus // 使用字典以便更方便地管理节点状态
}

type ResponseStatus struct {
	Status       string `json:"status"`
	LastStatus   string `json:"last_status"`
	ActionTime   string `json:"active_time"`
	RecoveryTime string `json:"recovery_time"`
	LasterServer string `json:"laster_server"`
}

// SetStatus 设置节点状态
func (hc *HadoopCluster) SetStatus(nodeName string, server []string, status string) {
	if hc.Nodes == nil {
		hc.Nodes = make(map[string]ResponseStatus)
	}

	currentStatus, exists := hc.Nodes[nodeName]
	if !exists {
		currentStatus = ResponseStatus{LastStatus: "unknown"}
	}

	currentStatus.LasterServer = ServerDataJoinSet(server)
	currentStatus.LastStatus = currentStatus.Status // 更新上次状态
	currentStatus.Status = status

	if status == "running" && currentStatus.LastStatus == "error" {
		currentStatus.RecoveryTime = time.Now().Format("2006-01-02 15:04:05")
	}

	if currentStatus.Status == "error" && currentStatus.LastStatus != "error" {
		currentStatus.ActionTime = time.Now().Format("2006-01-02 15:04:05")
	}

	// 处理告警恢复的时间数据
	if status == "recovery" {
		currentStatus.Status = "running"
		currentStatus.LastStatus = currentStatus.Status
		currentStatus.RecoveryTime = ""
	}

	hc.Nodes[nodeName] = currentStatus
}

// GetStatus 获取节点状态
func (hc *HadoopCluster) GetStatus(nodeName string) (ResponseStatus, bool) {
	status, exists := hc.Nodes[nodeName]
	return status, exists
}

func ServerDataJoinSet(ServerInfo []string) string {
	return strings.Join(ServerInfo, ",")
}

// 配置对接
type Config struct {
	Nodes    []Node        `mapstructure:"nodes"`
	Check    CheckSecond   `mapstructure:"check"`
	Telegram Telegram      `mapstructure:"telegram"`
	Logger   LoggerConfig  `mapstructure:"logger"`
	Session  SessionConfig `mapstructure:"session"`
	Auth     AuthConfig    `mapstructure:"auth"`
}

type Node struct {
	Name       string   `mapstructure:"name"`
	SSH_HOST   string   `mapstructure:"ssh_host"`
	SSH_USER   string   `mapstructure:"ssh_user"`
	SSH_PORT   string   `mapstructure:"ssh_port"`
	SSHKeyFile string   `mapstructure:"ssh_keyfile"`
	Services   []string `mapstructure:"services"`
}

type CheckSecond struct {
	MAX_SECOND string `mapstructure:"max_second"`
}

type Telegram struct {
	Token  string `mapstructure:"token"`
	ChatID string `mapstructure:"chat_id"`
}

type LoggerConfig struct {
	LogLevel      string `mapstructure:"log_level"`
	LogPath       string `mapstructure:"log_path"`
	LogName       string `mapstructure:"log_name"`
	LogMaxSize    int    `mapstructure:"log_max_size"`    // MB
	LogMaxBackups int    `mapstructure:"log_max_backups"` // backup sum
	LogMaxAge     int    `mapstructure:"log_max_age"`     // day
}

type SessionConfig struct {
	Secret string `mapstructure:"secret"`
}

type AuthConfig struct {
	Username string `mapstructure:"user"`
	Password string `mapstructure:"passwd"`
}

type ViewsStatus struct {
	HadoopStatus bool
	HdfsStatus   bool
	YarnStatus   bool
}

type HadoopStatusInfo struct {
	CPU string
	RAM string
	ROM string
}
