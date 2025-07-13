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

## 部署指南

### 开发环境
```bash
# 1. 确保 Go 1.21+ 已安装
go version

# 2. 克隆项目
git clone <repository-url>
cd GRUniChat-Onebot

# 3. 安装依赖
go mod tidy

# 4. 运行项目
go run main.go
```

## 故障排除

### 1. 配置文件问题
**问题**: 程序启动后立即退出，提示配置文件创建
**解决**: 
- 检查程序是否有写入权限
- 修改生成的 `config.yaml` 文件中的关键配置
- 确保 GRUniChat 和 OneBot 服务器地址正确

### 2. WebSocket 连接失败
**问题**: 连接 GRUniChat 或 OneBot 服务器失败
**解决**:
- 检查服务器是否正在运行：`telnet <host> <port>`
- 确认 WebSocket URL 格式正确（以 `ws://` 或 `wss://` 开头）
- 检查防火墙和网络设置
- 查看详细日志：将日志级别设置为 `debug`

### 3. 消息过滤问题
**问题**: 消息没有被转发或被错误过滤
**解决**:
- 检查 `filter.service_groups` 配置，空数组表示服务所有群
- 确认用户不在黑名单 `filter.blacklist_users` 中
- 检查 `filter.message_types` 是否包含所需的消息类型

### 4. 性能问题
**问题**: 消息处理延迟或丢失
**解决**:
- 增加 `performance.message_queue_size`
- 调整 `performance.worker_count` 工作协程数量
- 检查系统资源使用情况
- 监控日志中的队列满载警告

### 5. 日志分析
启用详细日志进行问题诊断：
```yaml
log:
  level: "debug"
  format: "json"  # JSON格式便于分析
  file: "adapter.log"
```

常见日志关键字：
- `WebSocket connection established` - 连接成功
- `Failed to connect` - 连接失败
- `Message filtered` - 消息被过滤
- `Command confirmation sent` - 命令确认已发送

### 6. 版本检查问题
**问题**: 版本检查失败或超时
**解决**:
- 网络连接问题：检查是否可以访问 GitHub API
- 超时问题：版本检查会在5秒后超时，这不影响程序正常运行
- 跳过检查：使用 `--no-check-update` 参数跳过版本检查
- 企业环境：在受限网络环境中建议使用 `--no-check-update` 参数

**示例**:
```bash
# 跳过版本检查
.\grunichat-onebot-adapter.exe --no-check-update

# 或者结合其他参数
.\grunichat-onebot-adapter.exe -config ./my-config.yaml --no-check-update
```

## 开发指南

### 项目结构
```
GRUniChat-Onebot/
├── main.go                    # 程序入口点
├── go.mod                     # Go 模块依赖
├── go.sum                     # 依赖校验文件
├── config.yaml               # 配置文件（运行时生成）
├── .gitignore                # Git 忽略文件
├── README.md                 # 项目文档
└── internal/                 # 内部包目录
    ├── adapter/              # 主适配器逻辑
    │   └── adapter.go
    ├── config/               # 配置管理
    │   └── config.go
    ├── websocket/            # WebSocket 客户端
    │   └── websocket.go
    ├── types/                # 数据类型定义
    │   └── types.go
    ├── formatter/            # 消息格式化
    │   └── formatter.go
    ├── sender/               # 消息发送
    │   └── sender.go
    ├── confirmation/         # 命令确认
    │   └── confirmation.go
    └── converter/            # 消息转换
        └── converter.go
```

## 贡献指南

我们欢迎各种形式的贡献！在贡献之前，请阅读以下指南：

### 如何贡献

1. **报告问题**: 在 GitHub Issues 中报告 bug 或提出功能建议
2. **提交代码**: Fork 项目，创建功能分支，提交 Pull Request
3. **完善文档**: 改进 README、注释或添加示例
4. **测试**: 帮助测试新功能或报告兼容性问题

## 许可证

本项目采用 MIT 许可证 - 详情请参考 [LICENSE](LICENSE) 文件。

## 联系我们

- **项目主页**: [GitHub Repository](https://github.com/your-org/grunichat-onebot)
- **问题报告**: [GitHub Issues](https://github.com/your-org/grunichat-onebot/issues)
- **开发团队**: Glory Redstone Union
- **维护者**: caikun233

## 致谢

感谢以下开源项目的支持：
- [OneBot](https://11.onebot.dev) - QQ 机器人标准协议
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - Go WebSocket 实现
- [Logrus](https://github.com/sirupsen/logrus) - Go 结构化日志库

---

**🎯 让跨平台消息同步变得简单高效！**

