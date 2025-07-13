package formatter

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"grunichat-onebot-adapter/internal/config"
)

// 消息格式化器
type MessageFormatter struct {
	config *config.Config
	logger *logrus.Logger
}

// 创建消息格式化器
func NewMessageFormatter(cfg *config.Config, logger *logrus.Logger) *MessageFormatter {
	return &MessageFormatter{
		config: cfg,
		logger: logger,
	}
}

// 格式化OneBot群消息
func (mf *MessageFormatter) FormatOneBotGroupMessage(message string) string {
	format := mf.config.Format.GroupMessageFormat
	format = strings.ReplaceAll(format, "{message}", message)
	return format
}

// 格式化发送到OneBot的聊天消息
func (mf *MessageFormatter) FormatChatMessageForOneBot(from, sender, message string) string {
	return fmt.Sprintf("<[%s] %s> %s", from, sender, message)
}

// 格式化发送到OneBot的事件消息
func (mf *MessageFormatter) FormatEventMessageForOneBot(from, eventDetail string) string {
	return fmt.Sprintf("<[%s]> %s", from, eventDetail)
}

// 格式化确认消息
func (mf *MessageFormatter) FormatConfirmationMessage(senderName, command string) string {
	return fmt.Sprintf("@%s 您要执行命令：%s\n请回复 '确认' 或 '取消'", senderName, command)
}
