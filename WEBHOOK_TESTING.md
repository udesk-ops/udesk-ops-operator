# Webhook Testing Guide

本文档介绍如何对ScaleNotifyConfig Webhook进行测试。

## 测试类型

### 1. 单元测试 (推荐)

单元测试使用fake client，无需真实的Kubernetes集群，运行速度快且可靠。

```bash
# 运行webhook单元测试
make test-webhook-unit

# 或者直接使用go test
go test -v ./internal/webhook/v1beta1 -run TestScaleNotifyConfigWebhook
```

### 2. 集成测试 (可选)

集成测试需要运行中的Kubernetes集群和webhook服务。

```bash
# 先启动webhook服务
make run-with-webhook

# 在另一个终端运行集成测试
make test-webhook-integration
```

### 3. 所有测试

运行所有测试，包括单元测试和控制器测试：

```bash
make test
```

## 测试覆盖的场景

单元测试覆盖以下webhook验证场景：

### ValidateCreate 测试
1. ✅ **首次创建某类型的默认配置** - 应该成功
2. ✅ **创建重复的同类型默认配置** - 应该被拒绝
3. ✅ **创建不同类型的默认配置** - 应该成功  
4. ✅ **创建非默认配置** - 应该成功
5. ✅ **创建无效类型的配置** - 应该被拒绝
6. ✅ **创建Email配置缺少必需字段** - 应该被拒绝
7. ✅ **创建WXWorkRobot配置缺少必需字段** - 应该被拒绝

### ValidateUpdate 测试
1. ✅ **将非默认配置更新为默认配置（无其他默认配置时）** - 应该成功
2. ✅ **将非默认配置更新为默认配置（已有其他默认配置时）** - 应该被拒绝
3. ✅ **更新默认配置的内容但不改变默认状态** - 应该成功

### ValidateDelete 测试
1. ✅ **删除任何配置** - 应该成功

## Webhook 验证规则

### 唯一默认配置约束
- 每种通知类型（Email、WXWorkRobot）只能有一个默认配置
- 创建新的默认配置时会检查是否已存在同类型的默认配置
- 更新配置为默认时也会进行相同检查

### 配置验证
- **Email配置必需字段**: `smtpServer`, `smtpUser`, `smtpPassword`, `fromEmail`, `toEmail`
- **WXWorkRobot配置必需字段**: `webhookURL`
- 支持的通知类型: `Email`, `WXWorkRobot`

## 开发建议

1. **优先运行单元测试**: 开发过程中主要依赖单元测试，它们快速、可靠且不依赖外部环境
2. **集成测试用于最终验证**: 在部署前运行集成测试确保webhook在真实环境中正常工作
3. **添加新功能时更新测试**: 修改webhook逻辑时，请相应更新单元测试

## 故障排除

如果测试失败，请检查：

1. **Import路径**: 确保所有导入路径正确
2. **依赖项**: 运行 `go mod tidy` 确保依赖项最新
3. **Webhook逻辑**: 检查webhook实现是否与测试期望一致
4. **错误消息匹配**: 测试中的错误消息断言是否与实际错误消息匹配
