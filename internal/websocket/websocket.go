package websocket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"grunichat-onebot-adapter/internal/config"
)

// WebSocket管理器接口
type IWebSocketManager interface {
	Connect(ctx context.Context) error
	SendMessage(message interface{}) error
	SetMessageHandler(handler func(message []byte))
	Close() error
	IsConnected() bool
}

// OneBot WebSocket客户端管理器
type OneBotWebSocketManager struct {
	config         *config.Config
	logger         *logrus.Logger
	conn           *websocket.Conn
	handler        func(message []byte)
	connected      bool
	reconnectDelay time.Duration
}

// 创建OneBot WebSocket管理器
func NewOneBotWebSocketManager(cfg *config.Config, logger *logrus.Logger) *OneBotWebSocketManager {
	return &OneBotWebSocketManager{
		config:         cfg,
		logger:         logger,
		reconnectDelay: 5 * time.Second,
	}
}

// 连接OneBot WebSocket
func (ws *OneBotWebSocketManager) Connect(ctx context.Context) error {
	wsURL := ws.config.OneBot.WebSocketURL

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	if ws.config.OneBot.AccessToken != "" {
		headers.Set("Authorization", "Bearer "+ws.config.OneBot.AccessToken)
	}

	ws.logger.Infof("Connecting to OneBot at %s", wsURL)

	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect to OneBot: %w", err)
	}

	ws.conn = conn
	ws.connected = true
	ws.logger.Info("Connected to OneBot WebSocket")

	// 启动消息读取协程
	go ws.readMessages(ctx)

	return nil
}

// 发送消息到OneBot
func (ws *OneBotWebSocketManager) SendMessage(message interface{}) error {
	if !ws.connected || ws.conn == nil {
		return fmt.Errorf("OneBot WebSocket not connected")
	}

	return ws.conn.WriteJSON(message)
}

// 设置消息处理器
func (ws *OneBotWebSocketManager) SetMessageHandler(handler func(message []byte)) {
	ws.handler = handler
}

// 关闭连接
func (ws *OneBotWebSocketManager) Close() error {
	ws.connected = false
	if ws.conn != nil {
		return ws.conn.Close()
	}
	return nil
}

// 检查连接状态
func (ws *OneBotWebSocketManager) IsConnected() bool {
	return ws.connected
}

// 读取消息协程
func (ws *OneBotWebSocketManager) readMessages(ctx context.Context) {
	defer func() {
		ws.connected = false
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				ws.logger.Errorf("OneBot WebSocket read error: %v", err)
				ws.connected = false
				return
			}

			if ws.handler != nil {
				ws.handler(message)
			}
		}
	}
}

// GRUniChat WebSocket客户端管理器
type GRUniChatWebSocketManager struct {
	config         *config.Config
	logger         *logrus.Logger
	conn           *websocket.Conn
	handler        func(message []byte)
	connected      bool
	reconnectDelay time.Duration
}

// 创建GRUniChat WebSocket管理器
func NewGRUniChatWebSocketManager(cfg *config.Config, logger *logrus.Logger) *GRUniChatWebSocketManager {
	return &GRUniChatWebSocketManager{
		config:         cfg,
		logger:         logger,
		reconnectDelay: 5 * time.Second,
	}
}

// 连接GRUniChat WebSocket服务器
func (ws *GRUniChatWebSocketManager) Connect(ctx context.Context) error {
	wsURL := ws.config.GRUniChat.URL

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	ws.logger.Infof("Connecting to GRUniChat at %s", wsURL)

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to GRUniChat: %w", err)
	}

	ws.conn = conn
	ws.connected = true
	ws.logger.Info("Connected to GRUniChat WebSocket")

	// 发送hello消息进行认证
	if err := ws.sendHelloMessage(); err != nil {
		ws.logger.Errorf("Failed to send hello message: %v", err)
		conn.Close()
		ws.connected = false
		return err
	}

	// 启动消息读取协程
	go ws.readMessages(ctx)

	return nil
}

// 发送hello认证消息
func (ws *GRUniChatWebSocketManager) sendHelloMessage() error {
	helloMsg := map[string]interface{}{
		"type": "hello",
		"from": ws.config.GRUniChat.ClientID,
	}

	ws.logger.Debugf("Sending hello message: %+v", helloMsg)

	if err := ws.conn.WriteJSON(helloMsg); err != nil {
		return fmt.Errorf("failed to send hello message: %w", err)
	}

	ws.logger.Debug("Hello message sent successfully")
	return nil
}

// 读取消息协程
func (ws *GRUniChatWebSocketManager) readMessages(ctx context.Context) {
	defer func() {
		ws.connected = false
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				ws.logger.Errorf("GRUniChat WebSocket read error: %v", err)
				ws.connected = false
				return
			}

			if ws.handler != nil {
				ws.handler(message)
			}
		}
	}
}

// 发送消息到GRUniChat
func (ws *GRUniChatWebSocketManager) SendMessage(message interface{}) error {
	if !ws.connected || ws.conn == nil {
		ws.logger.Warn("GRUniChat client not connected, skipping message")
		return nil // 不返回错误，因为客户端可能没有连接
	}

	return ws.conn.WriteJSON(message)
}

// 设置消息处理器
func (ws *GRUniChatWebSocketManager) SetMessageHandler(handler func(message []byte)) {
	ws.handler = handler
}

// 关闭连接
func (ws *GRUniChatWebSocketManager) Close() error {
	ws.connected = false
	if ws.conn != nil {
		return ws.conn.Close()
	}
	return nil
}

// 检查连接状态
func (ws *GRUniChatWebSocketManager) IsConnected() bool {
	return ws.connected
}

// WebSocket管理器工厂
type WebSocketManagerFactory struct {
	config *config.Config
	logger *logrus.Logger
}

// 创建WebSocket管理器工厂
func NewWebSocketManagerFactory(cfg *config.Config, logger *logrus.Logger) *WebSocketManagerFactory {
	return &WebSocketManagerFactory{
		config: cfg,
		logger: logger,
	}
}

// 创建OneBot WebSocket管理器
func (f *WebSocketManagerFactory) CreateOneBotManager() IWebSocketManager {
	return NewOneBotWebSocketManager(f.config, f.logger)
}

// 创建GRUniChat WebSocket管理器
func (f *WebSocketManagerFactory) CreateGRUniChatManager() IWebSocketManager {
	return NewGRUniChatWebSocketManager(f.config, f.logger)
}
