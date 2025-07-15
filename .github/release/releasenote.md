# GRUniChat-OneBot 适配器 v0.1.1

添加了一个配置项，可以决定是否转发指令执行结果的消息。

## 下载

请选择适合您系统的版本进行下载：

| 平台 | 架构 | 文件名 |
|------|------|--------|
| **Windows** | x64 | `GRUniChat-OneBot-Adapter-v1.0.0-windows-amd64.exe` |
| **Windows** | ARM64 | `GRUniChat-OneBot-Adapter-v1.0.0-windows-arm64.exe` |
| **Linux** | x64 | `GRUniChat-OneBot-Adapter-v1.0.0-linux-amd64` |
| **Linux** | ARM64 | `GRUniChat-OneBot-Adapter-v1.0.0-linux-arm64` |
| **Linux** | 32位 | `GRUniChat-OneBot-Adapter-v1.0.0-linux-386` |
| **Linux** | ARM | `GRUniChat-OneBot-Adapter-v1.0.0-linux-arm` |
| **macOS** | Intel | `GRUniChat-OneBot-Adapter-v1.0.0-darwin-amd64` |
| **macOS** | Apple Silicon | `GRUniChat-OneBot-Adapter-v1.0.0-darwin-arm64` |
| **FreeBSD** | x64 | `GRUniChat-OneBot-Adapter-v1.0.0-freebsd-amd64` |
| **FreeBSD** | ARM64 | `GRUniChat-OneBot-Adapter-v1.0.0-freebsd-arm64` |

## 使用方法

### 基础使用

```bash
# 1. 下载适合您系统的可执行文件
# 2. 首次运行，自动创建配置文件
./GRUniChat-OneBot-Adapter

# 3. 修改配置文件 config.yaml
# 4. 再次运行程序
./GRUniChat-OneBot-Adapter
```

### 配置示例

```yaml
# 命令权限配置
command:
  require_permission: true                # 启用权限验证
  authorized_users: [123456789, 987654321]  # 授权用户QQ号列表
  permission_denied_msg: "权限不足，您无权执行此命令"

# 群聊过滤配置  
filter:
  service_groups: [111111111, 222222222]  # 提供服务的群聊ID列表
  blacklist_users: [333333333]           # 黑名单用户ID列表
```

### 命令示例

```bash
# 用户在QQ群中发送（需要权限）
!!command survival /weather clear

# 系统自动转换为GRUniChat格式并转发
# 无权限用户会收到拒绝消息
```

## 主要功能

### 权限控制系统
```yaml
command:
  require_permission: true
  authorized_users: [123456789, 987654321]
  permission_denied_msg: "权限不足，您无权执行此命令"
```

### 智能消息路由
```go
// OneBot 群消息自动转换为 GRUniChat 格式
{
  "from": "QQ",
  "type": "chat", 
  "body": {
    "sender": "用户昵称",
    "chatMessage": "消息内容"
  }
}
```

### 群组过滤系统
```yaml
filter:
  service_groups: [111111111, 222222222]  # 白名单群组
  blacklist_users: [333333333]           # 黑名单用户
  message_types: ["group"]               # 仅支持群聊
```
---

## 快速开始

1. **下载**: 选择适合您系统的版本
2. **配置**: 首次运行自动生成配置文件
3. **修改**: 根据环境调整配置参数
4. **启动**: 再次运行即可开始使用

**详细文档**: [README.md](https://github.com/your-org/grunichat-onebot/blob/main/README.md)

**🐛 问题反馈**: [GitHub Issues](https://github.com/your-org/grunichat-onebot/issues)

---

**🎯 让跨平台消息同步变得简单高效！**
