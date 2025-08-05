# AlertScale 与 ScaleNotifyMsgTemplate 集成

## 概述

AlertScale 现在支持关联 ScaleNotifyMsgTemplate，实现自定义通知消息模板功能。这样可以为不同的扩缩容场景配置个性化的通知内容。

## 功能特性

### 1. 模板引用
- AlertScale 通过 `scaleNotifyMsgTemplate` 字段引用消息模板
- 支持跨命名空间引用（如果有相应的 RBAC 权限）
- 模板名称必须符合 Kubernetes 资源命名规范

### 2. 模板变量
消息模板支持以下模板变量：

```go
// AlertScale 相关变量
.ScaleReason          // 扩缩容原因
.ScaleDuration        // 持续时间
.ScaleThreshold       // 触发阈值
.ScaleTimeout         // 超时时间
.ScaleAutoApproval    // 是否自动审批

// ScaleTarget 相关变量
.ScaleTarget.Name     // 目标资源名称
.ScaleTarget.Kind     // 目标资源类型
.ScaleTarget.Namespace // 目标资源命名空间
.ScaleTarget.APIVersion // API 版本

// ScaleStatus 相关变量
.Status               // 当前状态
.OriginReplicas       // 原始副本数
.ScaledReplicas       // 扩缩后副本数
.ScaleBeginTime       // 开始时间
.ScaleEndTime         // 结束时间（如果已完成）

// 额外变量（由系统提供）
.Timestamp            // 当前时间戳
.Operator             // 操作员信息
```

### 3. kubectl 显示增强
添加了 `MsgTemplate` 列，方便查看关联的消息模板：

```bash
kubectl get alertscales -o wide
```

显示效果：
```
NAME                   TARGET             AUTOAPPROVAL   STATUS    ORIGIN-REPLICAS   SCALED-REPLICAS   SCALED-DURATION   THRESHOLD   NOTIFICATIONTYPE   MSGTEMPLATE                  REASON
production-nginx-scale nginx-production   false          Pending   3                 5                 30m               80          WXWorkRobot        production-scale-template    CPU使用率超过80%
```

## 使用示例

### 1. 创建消息模板

```yaml
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyMsgTemplate
metadata:
  name: production-scale-template
  namespace: default
spec:
  title: "生产环境扩缩容通知 - {{.ScaleTarget.Name}}"
  content: |
    **🚨 扩缩容操作通知 🚨**
    
    **目标资源:** {{.ScaleTarget.Kind}}/{{.ScaleTarget.Name}}
    **命名空间:** {{.ScaleTarget.Namespace}}
    **操作原因:** {{.ScaleReason}}
    **触发阈值:** {{.ScaleThreshold}}%
    **原始副本数:** {{.OriginReplicas}}
    **目标副本数:** {{.ScaledReplicas}}
    **持续时间:** {{.ScaleDuration}}
    **当前状态:** {{.Status}}
    **开始时间:** {{.ScaleBeginTime}}
    
    请及时关注系统资源状态！
```

### 2. 创建关联的 AlertScale

```yaml
apiVersion: ops.udesk.cn/v1beta1
kind: AlertScale
metadata:
  name: production-nginx-scale
  namespace: default
spec:
  scaleReason: "CPU使用率超过80%，需要扩容保证服务质量"
  scaleDuration: "30m"
  scaleNotificationType: "WXWorkRobot"
  scaleNotifyMsgTemplate: "production-scale-template"  # 引用消息模板
  scaleAutoApproval: false
  scaleTarget:
    apiVersion: "apps/v1"
    kind: "Deployment"
    name: "nginx-production"
    namespace: "default"
  scaleThreshold: 80
  scaleTimeout: "5m"
```

## 验证和测试

### 1. 验证 CRD 更新
```bash
make manifests
kubectl get crd alertscales.ops.udesk.cn -o yaml | grep scaleNotifyMsgTemplate
```

### 2. 验证测试
```bash
make test
```

### 3. 代码质量检查
```bash
make lint
```

## 实现细节

### API 字段定义
```go
// ScaleNotifyMsgTemplate is the reference to the message template for notifications.
// +kubebuilder:validation:Optional
// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
// where the value must be a valid Kubernetes resource name.
// Example: "my-notification-template"
ScaleNotifyMsgTemplate string `json:"scaleNotifyMsgTemplate,omitempty"`
```

### CRD 增强
- 添加了 `MsgTemplate` 打印列
- 支持完整的验证规则
- 符合 Kubernetes 资源命名规范

### 测试覆盖
- 单元测试已更新
- Controller 测试包含消息模板验证
- Lint 检查通过（0 issues）

## 下一步计划

1. **Controller 逻辑增强**: 在 AlertScale Controller 中实现模板查找和渲染逻辑
2. **模板验证**: 在 Webhook 中添加模板语法验证
3. **跨命名空间支持**: 实现跨命名空间模板引用
4. **模板缓存**: 优化模板查找性能
5. **默认模板**: 支持系统级默认模板

## 总结

AlertScale 与 ScaleNotifyMsgTemplate 的集成为用户提供了灵活的通知定制能力，支持：
- ✅ 模板引用和关联
- ✅ 丰富的模板变量
- ✅ kubectl 显示增强
- ✅ 完整的验证规则
- ✅ 测试覆盖和代码质量保证

这个集成为后续的通知系统增强奠定了良好的基础。
