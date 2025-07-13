package confirmation

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"grunichat-onebot-adapter/internal/formatter"
	"grunichat-onebot-adapter/internal/sender"
	"grunichat-onebot-adapter/internal/types"
	"grunichat-onebot-adapter/internal/websocket"
)

// 统一的确认管理器接口
type IConfirmationManager interface {
	HandleConfirmationCommand(onebot *types.OneBotMessage, senderName, command, originalMsg string)
	HandleConfirmationReply(onebot *types.OneBotMessage, message string) bool
	CleanupExpiredCommands()
	GetPendingCount() int
}

// 命令确认管理器
type CommandConfirmationManager struct {
	pendingCommands map[string]*types.PendingCommand // key: userID_groupID
	formatter       *formatter.MessageFormatter
	sender          sender.IMessageSender
	grunichatSender websocket.IWebSocketManager // 用于向GRUniChat发送广播消息
	logger          *logrus.Logger
}

// 创建命令确认管理器
func NewCommandConfirmationManager(
	fmt *formatter.MessageFormatter,
	sender sender.IMessageSender,
	grunichatSender websocket.IWebSocketManager,
	logger *logrus.Logger,
) *CommandConfirmationManager {
	return &CommandConfirmationManager{
		pendingCommands: make(map[string]*types.PendingCommand),
		formatter:       fmt,
		sender:          sender,
		grunichatSender: grunichatSender,
		logger:          logger,
	}
}

// 处理需要确认的命令
func (ccm *CommandConfirmationManager) HandleConfirmationCommand(onebot *types.OneBotMessage, senderName, command, originalMsg string) {
	// 生成确认键
	confirmKey := fmt.Sprintf("%d_%d", onebot.UserID, onebot.GroupID)

	// 存储待确认的命令
	ccm.pendingCommands[confirmKey] = &types.PendingCommand{
		UserID:      onebot.UserID,
		GroupID:     onebot.GroupID,
		Command:     command,
		OriginalMsg: originalMsg,
		Timestamp:   time.Now().Unix(),
	}

	// 发送确认消息到群里
	confirmationMsg := ccm.formatter.FormatConfirmationMessage(senderName, command)
	ccm.sender.SendGroupMessage(onebot.GroupID, confirmationMsg)

	ccm.logger.Debugf("Command pending confirmation from user %d in group %d: %s", onebot.UserID, onebot.GroupID, command)
}

// 处理确认回复
func (ccm *CommandConfirmationManager) HandleConfirmationReply(onebot *types.OneBotMessage, message string) bool {
	// 检查是否为确认回复
	message = strings.TrimSpace(strings.ToLower(message))
	if message != "yes" && message != "y" && message != "确认" && message != "是" {
		// 检查取消命令
		if message == "cancel" || message == "no" || message == "取消" {
			confirmKey := fmt.Sprintf("%d_%d", onebot.UserID, onebot.GroupID)
			if _, exists := ccm.pendingCommands[confirmKey]; exists {
				delete(ccm.pendingCommands, confirmKey)
				ccm.sender.SendGroupMessage(onebot.GroupID, "命令已取消")
				return true
			}
		}
		return false
	}

	// 生成确认键
	confirmKey := fmt.Sprintf("%d_%d", onebot.UserID, onebot.GroupID)

	// 查找待确认的命令
	pending, exists := ccm.pendingCommands[confirmKey]
	if !exists {
		return false
	}

	// 检查是否超时（5分钟）
	if time.Now().Unix()-pending.Timestamp > 300 {
		delete(ccm.pendingCommands, confirmKey)
		ccm.logger.Debugf("Confirmation expired for user %d in group %d", onebot.UserID, onebot.GroupID)
		ccm.sender.SendGroupMessage(onebot.GroupID, "命令确认已超时，请重新发送命令")
		return false
	}

	// 执行确认的命令，发送到所有客户端
	ccm.logger.Debugf("Command confirmed by user %d in group %d: %s", onebot.UserID, onebot.GroupID, pending.Command)

	// 直接向GRUniChat发送广播消息（不带executeAt）
	ccm.executeConfirmedCommandToGRUniChat(pending)

	// 发送确认消息到QQ群
	ccm.sender.SendGroupMessage(onebot.GroupID, "命令已确认，正在广播到所有客户端")

	// 清理待确认命令
	delete(ccm.pendingCommands, confirmKey)

	return true
}

// 执行已确认的命令，直接发送到GRUniChat广播
func (ccm *CommandConfirmationManager) executeConfirmedCommandToGRUniChat(pending *types.PendingCommand) {
	// 构建要广播的GRUniChat消息（不带executeAt字段）
	gruniMsg := &types.GRUniChatMessage{
		From:        "QQ", // 或者从配置中获取
		TotalID:     uuid.New().String(),
		CurrentTime: time.Now().Format("2006-01-02 15:04:05"),
		Type:        "command",
		Body: types.GRUniChatBody{
			Command: pending.Command,
			Sender:  "QQ用户确认执行",
			// 注意：不设置ExecuteAt，这样会广播到所有客户端
		},
	}

	// 发送到GRUniChat
	if err := ccm.grunichatSender.SendMessage(gruniMsg); err != nil {
		ccm.logger.Errorf("Failed to send confirmed command to GRUniChat: %v", err)
	} else {
		ccm.logger.Infof("Successfully broadcasted confirmed command to all clients: %s", pending.Command)
	}
}

// 清理过期的待确认命令
func (ccm *CommandConfirmationManager) CleanupExpiredCommands() {
	now := time.Now().Unix()
	for key, pending := range ccm.pendingCommands {
		if now-pending.Timestamp > 300 { // 5分钟超时
			delete(ccm.pendingCommands, key)
			ccm.logger.Debugf("Cleaned up expired command from user %d in group %d", pending.UserID, pending.GroupID)
			ccm.sender.SendGroupMessage(pending.GroupID, "命令确认已超时，请重新发送命令")
		}
	}
}

// 获取待确认命令数量
func (ccm *CommandConfirmationManager) GetPendingCount() int {
	return len(ccm.pendingCommands)
}
