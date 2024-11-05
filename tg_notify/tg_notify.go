package tg_notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fastdevops/monitorx/logger"
	"github.com/fastdevops/monitorx/utils"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// Telegram 消息格式
type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

// Generate msg not running
func CreateAlertMessage(nodeName string, missingServices []string) string {
	fmt.Println("服务信息：", missingServices)
	return fmt.Sprintf("节点 %s 缺少服务: %s", nodeName, strings.Join(missingServices, ", "))
}

// send tg alert
func SendTGMessage(token, chatID, message string, AlertStatus *utils.HadoopCluster, nodeName string) (error, bool) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?parse_mode=Markdown", token)

	// status icon
	IconStatus := ""
	if HadoopServerStatusInfo, ok := AlertStatus.GetStatus(nodeName); !ok {
		logger.Logger.Error("节点信息不存在，无法进行告警通知")
		return nil, false
	} else {
		if HadoopServerStatusInfo.Status != "error" {
			IconStatus = "✅"
		} else {
			IconStatus = "🚨"
		}

		// alert msg payload
		messageText := fmt.Sprintf(
			"生产环境\n*核心问题*: %s\n*目标状态*: %s\n*异常对象*: %s\n*触发时间*: %s\n*恢复时间*: %s\n",
			message, HadoopServerStatusInfo.Status, HadoopServerStatusInfo.LasterServer, HadoopServerStatusInfo.ActionTime, HadoopServerStatusInfo.RecoveryTime)

		telegramMessage := TelegramMessage{
			ChatID: chatID, // 替换为你的频道用户名或聊天 ID
			Text:   fmt.Sprintf("%s %s", IconStatus, messageText),
		}

		messageBytes, err := json.Marshal(telegramMessage)
		if err != nil {
			logger.Logger.Error("告警消息格式化错误，请检查数据拼接与配置是否正常", zap.Error(err))
			return err, false
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(messageBytes))
		if err != nil {
			logger.Logger.Error("告警接口请求失败，请查看日志： ", zap.Error(err))
			return err, false
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var respBody bytes.Buffer
			respBody.ReadFrom(resp.Body)
			fmt.Printf("Failed to send alert to Telegram, status code: %d, response: %s\n", resp.StatusCode, respBody.String())
			return fmt.Errorf("failed to send alert to Telegram, status code: %d", resp.StatusCode), false
		}
		return nil, true
	}
}
