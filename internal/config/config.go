package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// 配置结构体
type Config struct {
	GRUniChat struct {
		URL                  string `yaml:"url"`
		ClientID             string `yaml:"client_id"`
		ReconnectInterval    int    `yaml:"reconnect_interval"`
		MaxReconnectAttempts int    `yaml:"max_reconnect_attempts"`
	} `yaml:"grunichat"`

	OneBot struct {
		WebSocketURL string `yaml:"websocket_url"`
		AccessToken  string `yaml:"access_token"`
		Secret       string `yaml:"secret"`
	} `yaml:"onebot"`

	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		File   string `yaml:"file"`
	} `yaml:"log"`

	Filter struct {
		ServiceGroups           []int64  `yaml:"service_groups"` // 提供服务的群聊列表
		BlacklistUsers          []int64  `yaml:"blacklist_users"`
		MessageTypes            []string `yaml:"message_types"`
		FilterCommandExecutions bool     `yaml:"filter_command_executions"` // 是否过滤命令执行结果消息
	} `yaml:"filter"`

	Command struct {
		EnableCommandRouting bool    `yaml:"enable_command_routing"`
		AuthorizedUsers      []int64 `yaml:"authorized_users"`      // 有权限执行命令的用户ID列表
		RequirePermission    bool    `yaml:"require_permission"`    // 是否启用权限验证
		PermissionDeniedMsg  string  `yaml:"permission_denied_msg"` // 权限不足时的回复消息
	} `yaml:"command"`

	Format struct {
		GroupMessageFormat string `yaml:"group_message_format"`
		ShowGroupID        bool   `yaml:"show_group_id"`
	} `yaml:"format"`

	Performance struct {
		MessageQueueSize int `yaml:"message_queue_size"`
		WorkerCount      int `yaml:"worker_count"`
		MessageTimeout   int `yaml:"message_timeout"`
	} `yaml:"performance"`
}

// 加载配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	setConfigDefaults(&config)

	return &config, nil
}

// 加载配置文件，如果不存在则创建默认配置
func LoadConfigWithAutoCreate(path string) (*Config, bool, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 文件不存在，创建默认配置
		if err := createDefaultConfig(path); err != nil {
			return nil, false, fmt.Errorf("failed to create default config: %w", err)
		}
		return nil, true, nil // 返回true表示创建了新配置文件
	}

	// 尝试读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		// 读取失败，创建默认配置
		if err := createDefaultConfig(path); err != nil {
			return nil, false, fmt.Errorf("failed to create default config after read error: %w", err)
		}
		return nil, true, nil // 返回true表示创建了新配置文件
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		// 解析失败，创建默认配置
		if err := createDefaultConfig(path); err != nil {
			return nil, false, fmt.Errorf("failed to create default config after parse error: %w", err)
		}
		return nil, true, nil // 返回true表示创建了新配置文件
	}

	// 设置默认值
	setConfigDefaults(&config)

	return &config, false, nil // 返回false表示使用了现有配置文件
}

// 创建默认配置文件
func createDefaultConfig(path string) error {
	// 添加注释的默认配置内容
	configContent := `# GRUniChat-OneBot 适配器配置文件
# 请根据实际环境修改以下配置

# GRUniChat 配置
grunichat:
  url: "ws://localhost:8765/ws"           # GRUniChat WebSocket 服务器地址
  client_id: "QQ"                         # 客户端ID，建议改为有意义的名称
  reconnect_interval: 5                   # 重连间隔（秒）
  max_reconnect_attempts: 10              # 最大重连次数

# OneBot v11 配置
onebot:
  websocket_url: "ws://localhost:3001/"   # OneBot WebSocket 服务器地址
  access_token: ""                        # 访问令牌（如果需要）
  secret: ""                              # 签名密钥（如果需要）

# 日志配置
log:
  level: "debug"                          # 日志级别: debug, info, warn, error
  format: "text"                          # 日志格式: text, json
  file: ""                                # 日志文件路径（留空输出到控制台）

# 消息过滤配置
filter:
  service_groups: []                      # 提供服务的群聊ID列表，空数组表示所有群聊
  blacklist_users: []                     # 黑名单用户ID列表
  message_types: ["group"]                # 处理的消息类型（仅支持群聊）
  filter_command_executions: false        # 是否过滤命令执行结果消息

# 命令配置
command:
  enable_command_routing: true            # 是否启用命令路由
  require_permission: true                # 是否启用权限验证
  authorized_users: []                    # 有权限执行命令的用户QQ号列表，例如: [123456789, 987654321]
  permission_denied_msg: "权限不足，您无权执行此命令"  # 权限不足时的回复消息

# 消息格式配置
format:
  group_message_format: "{message}"       # 群消息格式模板
  show_group_id: false                    # 是否显示群ID

# 性能配置
performance:
  message_queue_size: 1000                # 消息队列大小
  worker_count: 5                         # 工作协程数量
  message_timeout: 10                     # 消息超时时间（秒）
`

	// 写入文件
	if err := os.WriteFile(path, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// 设置配置默认值
func setConfigDefaults(config *Config) {
	if config.GRUniChat.URL == "" {
		config.GRUniChat.URL = "ws://localhost:8765/ws"
	}
	if config.GRUniChat.ClientID == "" {
		config.GRUniChat.ClientID = "onebot_adapter"
	}
	if config.GRUniChat.ReconnectInterval == 0 {
		config.GRUniChat.ReconnectInterval = 5
	}
	if config.GRUniChat.MaxReconnectAttempts == 0 {
		config.GRUniChat.MaxReconnectAttempts = 10
	}

	if config.OneBot.WebSocketURL == "" {
		config.OneBot.WebSocketURL = "ws://localhost:5700/ws"
	}

	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.Log.Format == "" {
		config.Log.Format = "text"
	}

	if len(config.Filter.MessageTypes) == 0 {
		config.Filter.MessageTypes = []string{"group"} // 仅支持群聊消息
	}

	if config.Performance.MessageQueueSize == 0 {
		config.Performance.MessageQueueSize = 1000
	}
	if config.Performance.WorkerCount == 0 {
		config.Performance.WorkerCount = 5
	}
	if config.Performance.MessageTimeout == 0 {
		config.Performance.MessageTimeout = 10
	}

	// 设置默认的消息格式模板
	if config.Format.GroupMessageFormat == "" {
		config.Format.GroupMessageFormat = "{message}"
	}

	// 设置命令权限默认值
	if config.Command.PermissionDeniedMsg == "" {
		config.Command.PermissionDeniedMsg = "权限不足，您无权执行此命令"
	}
}

// 检查用户是否有命令执行权限
func (c *Config) HasCommandPermission(userID int64) bool {
	// 如果没有启用权限验证，允许所有用户
	if !c.Command.RequirePermission {
		return true
	}

	// 检查用户是否在授权列表中
	for _, authorizedUser := range c.Command.AuthorizedUsers {
		if authorizedUser == userID {
			return true
		}
	}

	return false
}

// 解析命令行黑名单参数
func ParseBlacklistGroups(blacklist string) []int64 {
	if blacklist == "" {
		return nil
	}

	var groups []int64
	for _, idStr := range strings.Split(blacklist, ",") {
		idStr = strings.TrimSpace(idStr)
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			groups = append(groups, id)
		}
	}

	return groups
}
