# 消息模板语法指南

AlertScale Operator 支持使用 Go 模板语法来自定义通知消息格式。

## 可用的模板变量

在 `messageTemplate` 中，你可以使用以下变量：

| 变量名 | 类型 | 描述 | 示例值 |
|--------|------|------|--------|
| `{{.Message}}` | string | 扩容消息内容 | `"deployment/my-app needs to scale to 5 replicas"` |
| `{{.Time}}` | string | 通知发送时间 (RFC3339 格式) | `"2025-07-25T19:45:30+08:00"` |

## 模板语法示例

### 企业微信机器人通知

#### 基础模板
```yaml
messageTemplate: "🚨 AlertScale 通知\n\n{{.Message}}\n\n时间: {{.Time}}"
```

**渲染结果**:
```
🚨 AlertScale 通知

deployment/my-app needs to scale to 5 replicas

时间: 2025-07-25T19:45:30+08:00
```

#### 高级模板（带条件和格式化）
```yaml
messageTemplate: |
  🔔 **AlertScale 扩容通知**
  
  📊 **详情**: {{.Message}}
  ⏰ **时间**: {{.Time}}
  🏷️ **环境**: 生产环境
  
  请及时关注资源使用情况！
```

### 邮件通知

#### HTML 格式模板
```yaml
messageTemplate: |
  <html>
  <body>
    <h2>🚨 AlertScale 扩容通知</h2>
    <p><strong>消息</strong>: {{.Message}}</p>
    <p><strong>时间</strong>: {{.Time}}</p>
    <hr>
    <p><em>此邮件由 AlertScale Operator 自动发送</em></p>
  </body>
  </html>
```

#### 纯文本格式模板
```yaml
messageTemplate: |
  AlertScale 扩容通知
  ==================
  
  消息: {{.Message}}
  时间: {{.Time}}
  
  ---
  此消息由 AlertScale Operator 自动发送
```

## 完整配置示例

### 企业微信机器人配置
```yaml
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyConfig
metadata:
  name: wxwork-notification
spec:
  type: WXWorkRobot
  default: true
  config:
    webhookURL: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=your-key"
    messageTemplate: |
      🚨 **AlertScale 通知**
      
      📋 **消息**: {{.Message}}
      ⏰ **时间**: {{.Time}}
      🔧 **操作**: 自动扩容
      
      请关注应用状态！
    atUsers: ["@all"]
    atAll: true
```

### 邮件通知配置
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
    username: "alerts@company.com" 
    password: "app-password"
    fromEmail: "noreply@company.com"
    toEmails: ["admin@company.com", "ops@company.com"]
    subject: "🚨 AlertScale 扩容通知"
    messageTemplate: |
      AlertScale 自动扩容通知
      
      详细信息:
      - 消息: {{.Message}}
      - 时间: {{.Time}}
      - 系统: Kubernetes AlertScale Operator
      
      请及时检查应用状态和资源使用情况。
      
      ---
      此邮件由 AlertScale Operator 自动发送，请勿回复。
```

## Go 模板语法参考

### 基本语法
- `{{.Variable}}` - 输出变量值
- `{{.Variable | printf "%s"}}` - 格式化输出
- `{{if .Variable}}...{{end}}` - 条件判断
- `{{range .Array}}...{{end}}` - 循环

### 常用函数
- `{{.Time | printf "发送时间: %s"}}` - 格式化字符串
- `{{len .Message}}` - 获取字符串长度
- `{{.Message | printf "%.50s"}}` - 截取前50个字符

### 多行文本
使用 YAML 的 `|` 或 `>` 语法来定义多行模板：

```yaml
messageTemplate: |
  第一行
  第二行
  {{.Message}}
```

## 错误处理

如果模板语法有错误，通知发送会失败并记录错误日志：

```
Failed to parse message template: template: wxwork:1:15: unexpected "}" in operand
```

## 最佳实践

1. **测试模板**: 在生产环境使用前，先在测试环境验证模板格式
2. **保持简洁**: 避免过于复杂的模板逻辑
3. **转义特殊字符**: 在企业微信中使用 Markdown 格式时注意转义
4. **多语言支持**: 可以根据需要使用中文或英文模板
5. **统一格式**: 在团队中保持模板格式的一致性

## 故障排除

### 常见问题

1. **模板解析失败**
   - 检查 `{{}}` 语法是否正确
   - 确保变量名拼写正确（区分大小写）

2. **消息格式异常**
   - 检查换行符和特殊字符
   - 验证 YAML 缩进是否正确

3. **变量为空**
   - 确认使用的变量名存在
   - 检查是否有拼写错误

### 调试方法

1. 查看 operator 日志：
   ```bash
   kubectl logs -f deployment/udesk-ops-operator-controller-manager -n udesk-ops-operator-system
   ```

2. 使用简单模板测试：
   ```yaml
   messageTemplate: "Test: {{.Message}}"
   ```

3. 逐步添加复杂度，确认每个部分都正常工作
