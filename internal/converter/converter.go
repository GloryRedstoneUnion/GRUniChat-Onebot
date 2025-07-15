package converter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"grunichat-onebot-adapter/internal/config"
	"grunichat-onebot-adapter/internal/confirmation"
	"grunichat-onebot-adapter/internal/formatter"
	"grunichat-onebot-adapter/internal/sender"
	"grunichat-onebot-adapter/internal/types"
)

// 消息转换器接口
type IMessageConverter interface {
	OneBotToGRUniChat(onebot *types.OneBotMessage) *types.GRUniChatMessage
	GRUniChatToOneBot(gruni *types.GRUniChatMessage) *types.OneBotMessage
}

// 消息过滤器
type MessageFilter struct {
	config         *config.Config
	logger         *logrus.Logger
	serviceGroups  map[int64]bool // 提供服务的群聊
	blacklistUsers map[int64]bool
}

// 创建消息过滤器
func NewMessageFilter(cfg *config.Config, logger *logrus.Logger) *MessageFilter {
	// 构建服务群聊映射
	serviceGroups := make(map[int64]bool)
	for _, groupID := range cfg.Filter.ServiceGroups {
		serviceGroups[groupID] = true
	}

	// 构建黑名单用户映射
	blacklistUsers := make(map[int64]bool)
	for _, userID := range cfg.Filter.BlacklistUsers {
		blacklistUsers[userID] = true
	}

	return &MessageFilter{
		config:         cfg,
		logger:         logger,
		serviceGroups:  serviceGroups,
		blacklistUsers: blacklistUsers,
	}
}

// 检查消息是否应该被过滤
func (mf *MessageFilter) ShouldFilter(onebot *types.OneBotMessage) bool {
	// 只处理群聊消息，过滤掉私聊消息
	if onebot.MessageType != "group" {
		mf.logger.Debugf("Message type %s not supported, only group messages are allowed", onebot.MessageType)
		return true
	}

	// 检查用户黑名单
	if onebot.UserID != 0 && mf.blacklistUsers[onebot.UserID] {
		mf.logger.Debugf("Message from blacklisted user %d, filtering", onebot.UserID)
		return true
	}

	// 检查是否为服务群聊（如果设置了服务群聊列表，只处理列表中的群聊）
	if len(mf.serviceGroups) > 0 && onebot.GroupID != 0 && !mf.serviceGroups[onebot.GroupID] {
		mf.logger.Debugf("Message from non-service group %d, filtering", onebot.GroupID)
		return true
	}

	return false
}

// 消息转换器
type MessageConverter struct {
	config              *config.Config
	logger              *logrus.Logger
	formatter           *formatter.MessageFormatter
	confirmationManager confirmation.IConfirmationManager
	onebotSender        sender.IMessageSender
	filter              *MessageFilter
}

// 创建消息转换器
func NewMessageConverter(
	cfg *config.Config,
	logger *logrus.Logger,
	fmt *formatter.MessageFormatter,
	confirmationManager confirmation.IConfirmationManager,
	onebotSender sender.IMessageSender,
) *MessageConverter {
	return &MessageConverter{
		config:              cfg,
		logger:              logger,
		formatter:           fmt,
		confirmationManager: confirmationManager,
		onebotSender:        onebotSender,
		filter:              NewMessageFilter(cfg, logger),
	}
}

// 将OneBot消息转换为GRUniChat消息
func (mc *MessageConverter) OneBotToGRUniChat(onebot *types.OneBotMessage) *types.GRUniChatMessage {
	if onebot.PostType != "message" {
		return nil // 暂时只处理消息类型
	}

	// 过滤消息
	if mc.filter.ShouldFilter(onebot) {
		return nil
	}

	// 构建发送者名称
	senderName := onebot.Sender.Nickname
	if onebot.Sender.Card != "" {
		senderName = onebot.Sender.Card
	}

	// 解析消息内容
	rawMessage := mc.extractMessageText(onebot.Message)

	// 检查是否为确认回复
	if mc.confirmationManager.HandleConfirmationReply(onebot, rawMessage) {
		return nil // 确认回复已处理，不需要转发
	}

	// 构建基础消息结构
	gruniMsg := &types.GRUniChatMessage{
		From:        mc.config.GRUniChat.ClientID, // 使用配置中的client_id
		TotalID:     uuid.New().String(),
		CurrentTime: time.Now().Format("2006-01-02 15:04:05"), // 使用正确的时间格式
		Body: types.GRUniChatBody{
			Sender: senderName, // 发送者昵称
		},
	}

	// 检查是否为命令 (!!command 格式)
	if strings.HasPrefix(rawMessage, "!!command ") {
		return mc.handleCommand(onebot, senderName, rawMessage, gruniMsg)
	}

	// 普通聊天消息
	gruniMsg.Type = "chat"
	gruniMsg.Body.ChatMessage = mc.formatter.FormatOneBotGroupMessage(rawMessage)

	// 设置群组路由信息（如果是群消息）
	if onebot.MessageType == "group" && onebot.GroupID != 0 {
		gruniMsg.Body.ExecuteAt = fmt.Sprintf("group_%d", onebot.GroupID)
	}

	return gruniMsg
}

// 处理命令消息
func (mc *MessageConverter) handleCommand(onebot *types.OneBotMessage, senderName, rawMessage string, gruniMsg *types.GRUniChatMessage) *types.GRUniChatMessage {
	// 检查用户权限
	if !mc.config.HasCommandPermission(onebot.UserID) {
		mc.logger.Warnf("User %d attempted to execute command without permission: %s", onebot.UserID, rawMessage)

		// 发送权限不足的回复消息
		mc.sendPermissionDeniedReply(onebot)
		return nil // 不转发命令
	}

	// 解析命令格式: !!command executeAt command_content
	parts := strings.SplitN(rawMessage, " ", 3)
	if len(parts) >= 3 {
		executeAt := parts[1]
		command := parts[2]

		// 检查特殊确认值
		if executeAt == "i_confirm_all_client" {
			mc.confirmationManager.HandleConfirmationCommand(onebot, senderName, command, rawMessage)
			return nil // 等待确认，不转发
		}

		gruniMsg.Type = "command"
		gruniMsg.Body.Command = command
		gruniMsg.Body.ExecuteAt = executeAt
	} else if len(parts) == 2 {
		executeAt := parts[1]

		// 检查特殊确认值
		if executeAt == "i_confirm_all_client" {
			mc.confirmationManager.HandleConfirmationCommand(onebot, senderName, "", rawMessage)
			return nil // 等待确认，不转发
		}

		// 只有 !!command executeAt 的情况
		gruniMsg.Type = "command"
		gruniMsg.Body.Command = ""
		gruniMsg.Body.ExecuteAt = executeAt
	} else {
		// 格式不正确，当作普通消息处理
		gruniMsg.Type = "chat"
		gruniMsg.Body.ChatMessage = mc.formatter.FormatOneBotGroupMessage(rawMessage)
	}

	return gruniMsg
}

// 将GRUniChat消息转换为OneBot消息（用于发送到OneBot）
func (mc *MessageConverter) GRUniChatToOneBot(gruni *types.GRUniChatMessage) *types.OneBotMessage {
	mc.logger.Debugf("Converting GRUniChat message to OneBot: %s from %s", gruni.Body.ChatMessage, gruni.Body.Sender)

	// 处理聊天类型和事件类型的消息
	if gruni.Type != "chat" && gruni.Type != "event" {
		mc.logger.Debugf("Ignoring message type: %s", gruni.Type)
		return nil
	}

	// 检查是否有ExecuteAt路由信息
	if gruni.Body.ExecuteAt == "" {
		mc.logger.Debug("No executeAt specified, broadcasting to all service groups")
		mc.broadcastToServiceGroups(gruni)
		return nil
	}

	// 解析ExecuteAt格式：group_123456
	if strings.HasPrefix(gruni.Body.ExecuteAt, "group_") {
		groupIDStr := strings.TrimPrefix(gruni.Body.ExecuteAt, "group_")
		if groupID, err := strconv.ParseInt(groupIDStr, 10, 64); err == nil {
			mc.sendToSpecificGroup(gruni, groupID)
		} else {
			mc.logger.Errorf("Invalid group ID in executeAt: %s", gruni.Body.ExecuteAt)
		}
	} else {
		mc.logger.Debugf("Unknown executeAt format: %s", gruni.Body.ExecuteAt)
	}

	return nil
}

// 广播消息到所有服务群组
func (mc *MessageConverter) broadcastToServiceGroups(gruni *types.GRUniChatMessage) {
	for groupID := range mc.getServiceGroups() {
		mc.sendToSpecificGroup(gruni, groupID)
	}
}

// 发送消息到指定群组
func (mc *MessageConverter) sendToSpecificGroup(gruni *types.GRUniChatMessage, groupID int64) {
	// 检查是否需要过滤命令执行结果消息
	if mc.config.Filter.FilterCommandExecutions && mc.isCommandExecutionMessage(gruni) {
		mc.logger.Debugf("Filtered command execution message from %s: %s", gruni.From, gruni.Body.EventDetail)
		return
	}

	var message string

	// 根据消息类型格式化内容
	if gruni.Type == "event" {
		// 事件消息格式：<[客户端]> 事件详情
		message = mc.formatter.FormatEventMessageForOneBot(gruni.From, gruni.Body.EventDetail)
	} else {
		// 聊天消息格式：<[客户端] 用户名> 消息内容
		message = mc.formatter.FormatChatMessageForOneBot(gruni.From, gruni.Body.Sender, gruni.Body.ChatMessage)
	}

	// 发送消息
	mc.onebotSender.SendGroupMessage(groupID, message)
	mc.logger.Debugf("Sent message to group %d: %s", groupID, message)
}

// 获取服务群组列表
func (mc *MessageConverter) getServiceGroups() map[int64]bool {
	serviceGroups := make(map[int64]bool)
	for _, groupID := range mc.config.Filter.ServiceGroups {
		serviceGroups[groupID] = true
	}
	return serviceGroups
}

// 从OneBot消息中提取文本内容（处理string和数组两种格式）
func (mc *MessageConverter) extractMessageText(message interface{}) string {
	switch msg := message.(type) {
	case string:
		// 如果是字符串直接返回
		return msg
	case []interface{}:
		// 如果是数组，遍历提取text类型的消息段
		var textParts []string
		for _, segment := range msg {
			if segmentMap, ok := segment.(map[string]interface{}); ok {
				if msgType, exists := segmentMap["type"]; exists && msgType == "text" {
					if data, dataExists := segmentMap["data"]; dataExists {
						if dataMap, dataOk := data.(map[string]interface{}); dataOk {
							if text, textExists := dataMap["text"]; textExists {
								if textStr, textOk := text.(string); textOk {
									textParts = append(textParts, textStr)
								}
							}
						}
					}
				}
			}
		}
		return strings.Join(textParts, "")
	default:
		// 未知格式，记录日志并返回空字符串
		mc.logger.Warnf("Unknown message format: %T", message)
		return ""
	}
}

// 发送权限不足的回复消息
func (mc *MessageConverter) sendPermissionDeniedReply(onebot *types.OneBotMessage) {
	// 只在群聊中回复权限不足消息
	if onebot.MessageType == "group" {
		mc.onebotSender.SendGroupMessage(onebot.GroupID, mc.config.Command.PermissionDeniedMsg)
		mc.logger.Debugf("Sent permission denied reply to user %d in group %d", onebot.UserID, onebot.GroupID)
	} else {
		mc.logger.Warnf("Permission denied reply only supported for group messages, ignoring %s message", onebot.MessageType)
	}
}

// 检测是否为命令执行结果消息
func (mc *MessageConverter) isCommandExecutionMessage(gruni *types.GRUniChatMessage) bool {
	// 只检查事件类型的消息
	if gruni.Type != "event" {
		return false
	}

	eventDetail := strings.ToLower(gruni.Body.EventDetail)

	// 检查是否包含命令执行相关的关键词
	commandKeywords := []string{
		"executed command",  // 执行命令
		"player executed",   // 玩家执行
		"changed the block", // 更改方块
		"command ->",        // 命令箭头
	}

	for _, keyword := range commandKeywords {
		if strings.Contains(eventDetail, keyword) {
			return true
		}
	}

	return false
}
