# 通知功能使用指南

AlertScale Operator 现在支持真正的通知发送功能，包括企业微信机器人和邮件通知。

## 功能特性

### 企业微信机器人通知 (WXWorkRobot)

- ✅ **HTTP POST 请求**: 直接向企业微信机器人 webhook 发送消息
- ✅ **签名验证**: 支持企业微信机器人的签名验证机制
- ✅ **@用户功能**: 支持 @ 特定用户或 @ 所有人
- ✅ **消息模板**: 支持自定义消息格式模板
- ✅ **错误处理**: 完整的错误处理和响应验证

#### 配置示例

```yaml
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyConfig
metadata:
  name: wxwork-notification
spec:
  type: WXWorkRobot
  default: true
  config:
    webhookURL: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=your-key-here"
    secret: "your-secret-key"  # 可选，用于签名验证
    atUsers: ["user1", "user2"]  # @ 特定用户
    atAll: true  # @ 所有人
    messageTemplate: "🚨 AlertScale 通知\n\n%s\n\n时间: %s"
```

#### 主要功能

1. **签名验证**: 如果配置了 `secret`，会自动生成时间戳和签名
2. **@ 功能**: 支持 @ 特定用户列表或所有人
3. **消息格式**: 支持富文本格式，包括 emoji
4. **错误响应处理**: 解析企业微信 API 响应，处理错误码

### 邮件通知 (Email)

- ✅ **SMTP 发送**: 使用标准 SMTP 协议发送邮件
- ✅ **多收件人**: 支持同时发送给多个收件人
- ✅ **SMTP 认证**: 支持用户名/密码认证
- ✅ **自定义主题**: 支持自定义邮件主题和模板
- ✅ **UTF-8 编码**: 支持中文内容

#### 配置示例

```yaml
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyConfig
metadata:
  name: email-notification
spec:
  type: Email
  default: true
  config:
    smtpServer: "smtp.gmail.com"
    smtpPort: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    fromEmail: "noreply@yourcompany.com"
    toEmails: ["admin@yourcompany.com", "ops@yourcompany.com"]
    subject: "AlertScale 扩容通知"
    messageTemplate: "AlertScale 通知\n\n%s\n\n发送时间: %s"
```

#### 主要功能

1. **多收件人**: `toEmails` 字段支持数组，可以同时发送给多个收件人
2. **SMTP 认证**: 支持大多数邮件服务提供商的 SMTP 认证
3. **自定义模板**: 支持自定义邮件主题和正文模板
4. **错误处理**: 对每个收件人单独处理，失败时提供详细错误信息

## 使用流程

1. **创建 ScaleNotifyConfig**: 配置通知参数
2. **设置默认通知**: 将 `default: true` 设置为默认通知方式
3. **AlertScale 自动通知**: 当 AlertScale 触发扩容时，会自动发送通知

## 验证功能

可以运行示例程序来测试通知功能：

```bash
go run examples/notify_example.go
```

## 常见问题

### 企业微信机器人

1. **获取 Webhook URL**: 在企业微信群聊中添加机器人，获取 webhook URL
2. **签名验证**: 如果机器人配置了签名验证，需要提供 `secret` 字段
3. **消息格式**: 支持 markdown 格式，可以使用 emoji 和格式化文本

### 邮件通知

1. **SMTP 配置**: 不同邮件服务商的 SMTP 设置不同
   - Gmail: smtp.gmail.com:587 (需要应用专用密码)
   - Outlook: smtp-mail.outlook.com:587
   - 企业邮箱: 请咨询邮箱管理员
2. **认证问题**: 建议使用应用专用密码而不是账户密码
3. **防火墙**: 确保 Kubernetes 集群可以访问 SMTP 服务器

## 错误处理

- 所有通知方法都包含完整的错误处理
- 失败时会记录详细的错误日志
- 网络问题或配置错误会返回具体的错误信息

## 测试

项目包含完整的单元测试：

```bash
make test  # 运行所有测试
```

测试覆盖了：
- 配置验证
- 字段格式检查
- Webhook 验证逻辑
- Controller 状态管理
