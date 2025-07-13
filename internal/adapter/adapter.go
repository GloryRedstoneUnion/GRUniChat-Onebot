package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"grunichat-onebot-adapter/internal/config"
	"grunichat-onebot-adapter/internal/confirmation"
	"grunichat-onebot-adapter/internal/converter"
	"grunichat-onebot-adapter/internal/formatter"
	"grunichat-onebot-adapter/internal/sender"
	"grunichat-onebot-adapter/internal/types"
	"grunichat-onebot-adapter/internal/websocket"
)

// 模块化适配器
type ModularAdapter struct {
	config              *config.Config
	logger              *logrus.Logger
	onebotWS            websocket.IWebSocketManager
	grunichatWS         websocket.IWebSocketManager
	messageConverter    *converter.MessageConverter
	wsFactory           *websocket.WebSocketManagerFactory
	formatter           *formatter.MessageFormatter
	confirmationManager confirmation.IConfirmationManager
	onebotSender        sender.IMessageSender
}

// 创建模块化适配器
func NewModularAdapter(cfg *config.Config, logger *logrus.Logger) *ModularAdapter {
	// 创建WebSocket工厂
	wsFactory := websocket.NewWebSocketManagerFactory(cfg, logger)

	// 创建WebSocket管理器
	onebotWS := wsFactory.CreateOneBotManager()
	grunichatWS := wsFactory.CreateGRUniChatManager()

	// 创建核心模块（需要按依赖顺序创建）
	formatter := formatter.NewMessageFormatter(cfg, logger)
	onebotSender := sender.NewOneBotMessageSender(onebotWS, logger)
	confirmationManager := confirmation.NewCommandConfirmationManager(formatter, onebotSender, grunichatWS, logger)
	messageConverter := converter.NewMessageConverter(cfg, logger, formatter, confirmationManager, onebotSender)

	return &ModularAdapter{
		config:              cfg,
		logger:              logger,
		onebotWS:            onebotWS,
		grunichatWS:         grunichatWS,
		messageConverter:    messageConverter,
		wsFactory:           wsFactory,
		formatter:           formatter,
		confirmationManager: confirmationManager,
		onebotSender:        onebotSender,
	}
}

// 启动适配器
func (adapter *ModularAdapter) Start(ctx context.Context) error {
	adapter.logger.Info("Starting GRUniChat-OneBot Modular Adapter")

	// 连接OneBot
	if err := adapter.connectOneBot(ctx); err != nil {
		return err
	}

	// 连接GRUniChat
	if err := adapter.connectGRUniChat(ctx); err != nil {
		return err
	}

	adapter.logger.Info("Modular adapter started successfully")

	// 等待信号或上下文取消
	return adapter.waitForShutdown(ctx)
}

// 连接OneBot
func (adapter *ModularAdapter) connectOneBot(ctx context.Context) error {
	maxAttempts := adapter.config.GRUniChat.MaxReconnectAttempts
	delay := time.Duration(adapter.config.GRUniChat.ReconnectInterval) * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		adapter.logger.Infof("Attempting to connect to OneBot (attempt %d/%d)", attempt, maxAttempts)

		if err := adapter.onebotWS.Connect(ctx); err != nil {
			adapter.logger.Errorf("Failed to connect to OneBot (attempt %d): %v", attempt, err)
			if attempt < maxAttempts {
				adapter.logger.Infof("Retrying in %v...", delay)
				time.Sleep(delay)
			}
			continue
		}

		// 设置消息处理器
		adapter.onebotWS.SetMessageHandler(adapter.handleOneBotMessage)
		break
	}

	if !adapter.onebotWS.IsConnected() {
		return fmt.Errorf("failed to connect to OneBot after %d attempts", maxAttempts)
	}

	return nil
}

// 连接GRUniChat
func (adapter *ModularAdapter) connectGRUniChat(ctx context.Context) error {
	adapter.logger.Info("Connecting to GRUniChat")

	if err := adapter.grunichatWS.Connect(ctx); err != nil {
		adapter.logger.Warnf("Failed to connect to GRUniChat: %v", err)
		adapter.logger.Info("Continuing without GRUniChat connection (OneBot-only mode)")
		return nil // 不返回错误，允许只连接OneBot
	}

	// 设置消息处理器
	adapter.grunichatWS.SetMessageHandler(adapter.handleGRUniChatMessage)
	return nil
}

// 处理OneBot消息
func (adapter *ModularAdapter) handleOneBotMessage(message []byte) {
	adapter.logger.Debugf("Received OneBot message: %s", string(message))

	var onebot types.OneBotMessage
	if err := json.Unmarshal(message, &onebot); err != nil {
		adapter.logger.Errorf("Failed to parse OneBot message: %v", err)
		return
	}

	// 基本过滤
	if onebot.PostType != "message" {
		adapter.logger.Debugf("Message filtered out: %+v", onebot)
		return
	}

	// 转换消息
	gruniMsg := adapter.messageConverter.OneBotToGRUniChat(&onebot)
	if gruniMsg == nil {
		return // 消息被过滤或已处理（如确认命令）
	}

	// 发送到GRUniChat
	if adapter.grunichatWS.IsConnected() {
		if err := adapter.grunichatWS.SendMessage(gruniMsg); err != nil {
			adapter.logger.Errorf("Failed to send message to GRUniChat: %v", err)
		} else {
			adapter.logger.Debugf("Sent message to GRUniChat: %+v", gruniMsg)
		}
	} else {
		adapter.logger.Debug("GRUniChat not connected, message not forwarded")
	}
}

// 处理GRUniChat消息
func (adapter *ModularAdapter) handleGRUniChatMessage(message []byte) {
	adapter.logger.Debugf("Received GRUniChat message: %s", string(message))

	var gruni types.GRUniChatMessage
	if err := json.Unmarshal(message, &gruni); err != nil {
		adapter.logger.Errorf("Failed to parse GRUniChat message: %v", err)
		return
	}

	// 转换并发送到OneBot
	adapter.messageConverter.GRUniChatToOneBot(&gruni)
}

// 等待关闭信号
func (adapter *ModularAdapter) waitForShutdown(ctx context.Context) error {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动清理任务
	go adapter.startCleanupTasks(ctx)

	select {
	case <-ctx.Done():
		adapter.logger.Info("Context cancelled, shutting down...")
	case sig := <-sigChan:
		adapter.logger.Infof("Received signal %v, shutting down...", sig)
	}

	return adapter.shutdown()
}

// 启动清理任务
func (adapter *ModularAdapter) startCleanupTasks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			adapter.confirmationManager.CleanupExpiredCommands()
		}
	}
}

// 关闭适配器
func (adapter *ModularAdapter) shutdown() error {
	adapter.logger.Info("Shutting down modular adapter...")

	// 关闭WebSocket连接
	if adapter.onebotWS != nil {
		adapter.onebotWS.Close()
	}
	if adapter.grunichatWS != nil {
		adapter.grunichatWS.Close()
	}

	adapter.logger.Info("Modular adapter shutdown complete")
	return nil
}
