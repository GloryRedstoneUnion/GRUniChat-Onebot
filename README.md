# GRUniChat-OneBot 适配器

一个用Go语言开发的模块化适配器，用于连接 OneBot v11 协议（QQ机器人标准协议）和 GRUniChat WebSocket消息广播服务。

## 功能特性

- ✅ **双向消息转换**：OneBot v11 ⇔ GRUniChat 无缝对接
- ✅ **自动配置管理**：首次运行自动创建配置文件，支持配置验证和热加载
- ✅ **群聊专用服务**：专为群聊场景优化，不支持私聊消息处理
- ✅ **智能消息过滤**：支持群组白名单、用户黑名单和消息类型过滤
- ✅ **命令权限控制**：支持用户权限验证，只有授权用户才能执行!!command命令
- ✅ **命令确认机制**：命令确认，确保命令执行状态同步
- ✅ **自动重连机制**：WebSocket连接断开时自动重连，提高服务稳定性
- ✅ **详细日志系统**：支持多级别日志、JSON格式输出和文件日志
- ✅ **自动更新检查**：启动时自动检查新版本，支持跳过检查

## 快速开始

### 1. 构建项目

```bash
# 克隆项目
git clone <repository-url>
cd GRUniChat-Onebot

# 构建可执行文件
go build -o grunichat-onebot-adapter.exe .
```

### 2. 运行适配器

#### 首次运行（自动创建配置）
```bash
# 直接运行，程序会自动创建默认配置文件
.\grunichat-onebot-adapter.exe

# 或指定配置文件路径
.\grunichat-onebot-adapter.exe -config ./my-config.yaml

# 跳过版本更新检查
.\grunichat-onebot-adapter.exe --no-check-update
```

程序首次运行时会：
1. 显示启动横幅
2. 检查版本更新（可通过 `--no-check-update` 跳过）
3. 自动创建配置文件并在五秒后退出

#### 正常运行
配置文件修改完成后，再次运行即可启动服务：
```bash
.\grunichat-onebot-adapter.exe
```

### 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-config` | `./config.yaml` | 指定配置文件路径 |
| `--no-check-update` | `false` | 跳过启动时的版本更新检查 |

## 配置文件详解

程序会自动生成详细的配置文件 `config.yaml`，包含以下主要配置：

### GRUniChat 连接配置
```yaml
grunichat:
  url: "ws://localhost:8765/ws"           # GRUniChat WebSocket 服务器地址
  client_id: "QQ"                         # 客户端标识，建议改为有意义的名称
  reconnect_interval: 5                   # 重连间隔（秒）
  max_reconnect_attempts: 10              # 最大重连次数
```

### OneBot v11 配置
```yaml
onebot:
  websocket_url: "ws://localhost:3001/"   # OneBot WebSocket 服务器地址
  access_token: ""                        # 访问令牌（如果需要）
  secret: ""                              # 签名密钥（如果需要）
```

### 消息过滤配置
```yaml
filter:
  service_groups: []                      # 提供服务的群聊ID列表，空数组表示所有群聊
  blacklist_users: []                     # 黑名单用户ID列表
  message_types: ["group"]                # 处理的消息类型（仅支持群聊）
  filter_command_executions: false        # 是否过滤命令执行结果消息（防止刷屏）
```

### 命令权限配置
```yaml
command:
  enable_command_routing: true            # 是否启用命令路由
  require_permission: true                # 是否启用权限验证
  authorized_users: []                    # 有权限执行命令的用户QQ号列表，例如: [123456789, 987654321]
  permission_denied_msg: "❌ 权限不足，您无权执行此命令"  # 权限不足时的回复消息
```

### 日志配置
```yaml
log:
  level: "debug"                          # 日志级别: debug, info, warn, error
  format: "text"                          # 日志格式: text, json
  file: ""                                # 日志文件路径（留空输出到控制台）
```

### 性能配置
```yaml
performance:
  message_queue_size: 1000                # 消息队列大小
  worker_count: 5                         # 工作协程数量
  message_timeout: 10                     # 消息超时时间（秒）
```

## 架构设计

### 模块化结构
```
internal/
├── adapter/         # 主适配器模块，协调各组件
├── config/          # 配置管理模块
├── websocket/       # WebSocket 连接管理
├── types/           # 数据类型定义
├── formatter/       # 消息格式化模块
├── sender/          # 消息发送模块
├── confirmation/    # 命令确认机制
└── converter/       # 消息转换模块
```

### 消息处理流程
1. **接收阶段**：WebSocket 客户端接收来自 OneBot 和 GRUniChat 的消息
2. **过滤阶段**：根据配置的过滤规则筛选消息
3. **转换阶段**：将消息格式在两种协议间转换
4. **确认阶段**：对于命令消息，处理确认回复
5. **发送阶段**：将转换后的消息发送到目标服务

## 协议转换详解

### 连接和握手

适配器启动后会：
1. 连接到 GRUniChat WebSocket 服务器
2. 连接到 OneBot WebSocket 服务器
3. 自动发送握手消息建立通信
4. 开始双向消息转换和路由


### 命令权限机制

系统支持基于用户ID的命令权限控制：

**权限配置**：
```yaml
command:
  require_permission: true                # 启用权限验证
  authorized_users: [123456789, 987654321]  # 授权用户QQ号列表
  permission_denied_msg: "❌ 权限不足，您无权执行此命令"
```

**权限验证流程**：
1. 用户发送 `!!command` 格式的命令
2. 系统检查用户ID是否在授权列表中
3. 如果有权限，正常处理并转发命令
4. 如果无权限，在群聊中回复权限不足消息

**权限拒绝回复示例**：
```
用户 (无权限): !!command survival /weather clear
系统回复: ❌ 权限不足，您无权执行此命令
```

### 命令执行结果过滤

系统支持过滤来自游戏客户端的命令执行结果消息，避免批量操作时的消息刷屏：

**过滤配置**：
```yaml
filter:
  filter_command_executions: true    # 启用命令执行结果过滤
```

**会被过滤的消息类型**：
- `<[redstone]> [redstone] Player The_skyM executed command: command -> - 易碎深板岩`
- `<[creative]> [creative] Player awszqsedxzaw executed command: command -> Changed the block at -912, 139, -603`
- `<[redstone]> [redstone] Player _XuanMing_ executed command: command -> Successfully filled 2 block(s)`

**使用场景**：
- 大批量方块编辑操作
- 自动化脚本执行
- 避免QQ群聊被命令执行日志刷屏

**注意**：启用此过滤器后，所有包含 "executed command"、"changed the block"、"successfully filled" 等关键词的事件消息都会被过滤。

