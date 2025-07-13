package sender

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"grunichat-onebot-adapter/internal/websocket"
)

// 统一的消息发送器接口（仅支持群聊）
type IMessageSender interface {
	SendGroupMessage(groupID int64, message string)
}

// OneBot消息发送器
type OneBotMessageSender struct {
	wsManager websocket.IWebSocketManager
	logger    *logrus.Logger
}

// 创建OneBot发送器
func NewOneBotMessageSender(wsManager websocket.IWebSocketManager, logger *logrus.Logger) *OneBotMessageSender {
	return &OneBotMessageSender{
		wsManager: wsManager,
		logger:    logger,
	}
}

// 发送群消息
func (s *OneBotMessageSender) SendGroupMessage(groupID int64, message string) {
	if !s.wsManager.IsConnected() {
		s.logger.Warn("OneBot WebSocket not connected, cannot send message")
		return
	}

	onebotMsg := map[string]interface{}{
		"action": "send_group_msg",
		"params": map[string]interface{}{
			"group_id": groupID,
			"message":  message,
		},
		"echo": "adapter_" + uuid.New().String(),
	}

	if err := s.wsManager.SendMessage(onebotMsg); err != nil {
		s.logger.Errorf("Failed to send group message: %v", err)
	} else {
		s.logger.Debugf("Sent group message to %d: %s", groupID, message)
	}
}
