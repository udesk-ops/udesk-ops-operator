# ScaleNotifyConfig Webhook

这个项目实现了一个 Kubernetes Admission Webhook，用于验证 `ScaleNotifyConfig` 资源，确保每个通知类型只能有一个默认配置。

## 功能特性

### 1. 验证Webhook (ValidatingAdmissionWebhook)
- **唯一性验证**: 确保每个通知类型（Email, WXWorkRobot）只能有一个 `default: true` 的配置
- **字段验证**: 验证必需字段和字段值的有效性
- **类型验证**: 确保 `type` 字段的值在允许范围内

### 2. 变更Webhook (MutatingAdmissionWebhook)
- **默认值设置**: 自动设置默认的 `ValidationStatus` 为 "Pending"
- **字段标准化**: 确保字段符合预期格式

## 使用方法

### 1. 部署Operator
```bash
# 构建和部署
make install
make deploy
```

### 2. 创建第一个默认配置 (成功)
```yaml
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyConfig
metadata:
  name: email-default
spec:
  default: true
  type: Email
  config:
    smtpServer: "smtp.example.com"
    smtpPort: 587
    # ... 其他配置
```

### 3. 尝试创建重复的默认配置 (失败)
```yaml
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyConfig
metadata:
  name: email-duplicate
spec:
  default: true  # 这会被webhook拒绝
  type: Email     # 因为已经存在Email类型的默认配置
  config:
    # ... 配置
```

当尝试创建上述重复配置时，webhook会返回类似以下的错误：
```
error validating data: ValidationError(ScaleNotifyConfig): 
a default ScaleNotifyConfig of type 'Email' already exists: email-default
```

## 工作流程

### 创建资源时
1. **Mutating Webhook**: 设置默认值
2. **Validating Webhook**: 检查是否违反唯一性约束
3. **Controller**: 配置验证通过后，设置通知客户端

### 更新资源时
1. **Mutating Webhook**: 更新默认值（如需要）
2. **Validating Webhook**: 检查新的配置是否冲突
3. **Controller**: 重新配置通知客户端

## 验证规则

### 基本字段验证
- `spec.type` 必须为 "Email" 或 "WXWorkRobot"
- `spec.config` 不能为空

### 唯一性验证
- 在同一命名空间中，每个通知类型只能有一个 `default: true` 的配置
- 更新现有配置不受此限制（除非更改了类型）

## 测试

### 测试样例文件
项目提供了以下测试样例：
- `config/samples/ops_v1beta1_scalenotifyconfig_email.yaml` - Email配置示例
- `config/samples/ops_v1beta1_scalenotifyconfig_wxwork.yaml` - 微信工作机器人配置示例  
- `config/samples/ops_v1beta1_scalenotifyconfig_duplicate.yaml` - 冲突配置示例（应该被拒绝）

### 测试步骤
```bash
# 1. 创建第一个默认配置（应该成功）
kubectl apply -f config/samples/ops_v1beta1_scalenotifyconfig_email.yaml

# 2. 尝试创建重复配置（应该失败）
kubectl apply -f config/samples/ops_v1beta1_scalenotifyconfig_duplicate.yaml

# 3. 查看webhook拒绝的错误信息
kubectl get events --field-selector reason=FailedCreate
```

## 架构优势

1. **预防而非修复**: Webhook在资源创建时就阻止冲突，而不是事后处理
2. **即时反馈**: 用户立即知道为什么配置被拒绝
3. **简化Controller**: Controller逻辑更加简单，主要负责业务逻辑而非验证
4. **符合Kubernetes最佳实践**: 使用标准的Admission Controller机制

## 故障排除

### Webhook不工作
1. 检查webhook配置是否正确部署：
   ```bash
   kubectl get validatingwebhookconfiguration
   kubectl get mutatingwebhookconfiguration
   ```

2. 检查webhook服务是否运行：
   ```bash
   kubectl get pods -n udesk-ops-operator-system
   kubectl logs -n udesk-ops-operator-system deployment/udesk-ops-operator-controller-manager
   ```

### 证书问题
Webhook需要TLS证书，确保cert-manager正确安装并配置。

### 调试模式
启用详细日志以调试webhook行为：
```bash
kubectl patch deployment -n udesk-ops-operator-system udesk-ops-operator-controller-manager \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","args":["--zap-log-level=debug"]}]}}}}'
```
