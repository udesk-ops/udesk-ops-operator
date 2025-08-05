# 审批抽象开发指南

本文档介绍了udesk-ops-operator项目中的通用审批抽象系统，包括如何使用现有的审批处理器以及如何为新资源类型添加审批支持。

## 文件结构

审批抽象系统的相关文件组织如下：

```
internal/server/handlers/
├── approval_common.go      # 通用审批处理器和核心接口
├── approval_adapters.go    # 资源适配器实现
├── approval_test.go        # 审批系统测试
├── approval.go            # 批量审批和统计API
├── alertscale.go          # AlertScale HTTP处理器
└── podrebalance.go        # PodRebalance HTTP处理器
```

**核心文件说明**：
- `approval_common.go`: 包含 `CommonApprovalProcessor`、接口定义和错误常量
- `approval_adapters.go`: 包含 `AlertScaleApprovalAdapter` 和 `PodRebalanceApprovalAdapter`
- `approval_test.go`: 完整的测试覆盖，包括单元测试和集成测试
- `approval.go`: 批量审批、统计和列表API的实现

## 概述

审批抽象系统使用适配器模式消除了AlertScale和PodRebalance资源类型之间的重复审批逻辑。它提供了：

- 统一的审批/拒绝工作流程
- 类型安全的资源适配器
- 一致的错误处理和响应格式
- 基于注解的状态管理
- 可扩展的架构设计

## 架构组件

### 1. 核心接口

#### ApprovalResource
定义了可用于审批工作流程的资源必须实现的基本接口：

```go
type ApprovalResource interface {
    client.Object
    GetAnnotations() map[string]string
    SetAnnotations(annotations map[string]string)
}
```

#### ApprovalStatusChecker
定义了检查资源是否处于审批状态的方法：

```go
type ApprovalStatusChecker interface {
    IsInApprovalState() bool
    GetStatusFieldName() string
}
```

### 2. 通用审批处理器

`CommonApprovalProcessor` 是核心组件，提供统一的审批处理逻辑：

```go
type CommonApprovalProcessor struct {
    client client.Client
}
```

主要方法：
- `ProcessApprovalRequest()` - 通用审批请求处理
- `ProcessAlertScaleApproval()` - AlertScale特定处理
- `ProcessPodRebalanceApproval()` - PodRebalance特定处理
- `HandleApprovalWithResponseWriter()` - HTTP响应处理

### 3. 适配器实现

#### AlertScaleApprovalAdapter
```go
type AlertScaleApprovalAdapter struct {
    *opsv1beta1.AlertScale
}

func (a *AlertScaleApprovalAdapter) IsInApprovalState() bool {
    return a.Status.ScaleStatus.Status == types.ScaleStatusApprovaling
}
```

#### PodRebalanceApprovalAdapter
```go
type PodRebalanceApprovalAdapter struct {
    *opsv1beta1.PodRebalance
}

func (a *PodRebalanceApprovalAdapter) IsInApprovalState() bool {
    return a.Status.Status == types.RebalanceStatusApprovaling
}
```

## 使用方法

### 1. 在HTTP处理器中使用

```go
func (h *AlertScaleHandler) approveAlertScale(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    namespace := vars["namespace"]
    name := vars["name"]

    // 解析请求
    var req CommonApprovalRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        responseWriter.WriteError(w, http.StatusBadRequest, "Invalid request body", err)
        return
    }

    // 创建通用审批处理器
    processor := NewCommonApprovalProcessor(h.client)

    // 处理审批
    resourceKey := client.ObjectKey{Namespace: namespace, Name: name}
    if err := processor.ProcessAlertScaleApproval(r.Context(), resourceKey, "approve", req); err != nil {
        // 错误处理
        switch {
        case err.Error() == ErrResourceNotFound:
            responseWriter.WriteError(w, http.StatusNotFound, "AlertScale not found", err)
        case err.Error() == ErrResourceNotInApprovalState:
            responseWriter.WriteError(w, http.StatusBadRequest, "AlertScale is not in approvaling state", err)
        case err.Error() == ErrApproverRequired || err.Error() == ErrReasonRequired:
            responseWriter.WriteError(w, http.StatusBadRequest, err.Error(), nil)
        default:
            responseWriter.WriteError(w, http.StatusInternalServerError, "Failed to approve AlertScale", err)
        }
        return
    }

    // 返回成功响应
    responseData := map[string]interface{}{
        "message":   "AlertScale approved successfully",
        "approver":  req.Approver,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    }
    responseWriter.WriteSuccess(w, responseData)
}
```

### 2. 使用通用处理方法

对于需要统一响应格式的场景，可以使用 `HandleApprovalWithResponseWriter`:

```go
processor := NewCommonApprovalProcessor(h.client)
processor.HandleApprovalWithResponseWriter(
    r.Context(),
    w,
    r,
    resourceKey,
    &opsv1beta1.AlertScale{},
    "approve",
    responseWriter,
)
```

## 为新资源添加审批支持

### 步骤1: 确认资源实现ApprovalResource接口

大多数Kubernetes资源已经实现了这个接口，因为它们嵌入了 `metav1.Object`。

### 步骤2: 创建适配器

```go
// MyResourceApprovalAdapter adapts MyResource to work with common approval processor
type MyResourceApprovalAdapter struct {
    *opsv1beta1.MyResource
}

// NewMyResourceApprovalAdapter creates a new adapter for MyResource
func NewMyResourceApprovalAdapter(resource *opsv1beta1.MyResource) *MyResourceApprovalAdapter {
    return &MyResourceApprovalAdapter{MyResource: resource}
}

// IsInApprovalState checks if MyResource is in approvaling state
func (a *MyResourceApprovalAdapter) IsInApprovalState() bool {
    return a.Status.Status == types.MyResourceStatusApprovaling
}

// GetStatusFieldName returns the status field name for debugging
func (a *MyResourceApprovalAdapter) GetStatusFieldName() string {
    return "Status.Status"
}
```

### 步骤3: 添加专用处理方法

```go
// ProcessMyResourceApproval processes approval for MyResource
func (p *CommonApprovalProcessor) ProcessMyResourceApproval(
    ctx context.Context,
    resourceKey client.ObjectKey,
    action string,
    req CommonApprovalRequest,
) error {
    // Create the actual MyResource object for k8s operations
    myResource := &opsv1beta1.MyResource{}

    // Get the current resource state first
    if err := p.client.Get(ctx, resourceKey, myResource); err != nil {
        return fmt.Errorf("failed to get resource: %w", err)
    }

    // Create adapter for status checking
    adapter := NewMyResourceApprovalAdapter(myResource)

    // Use the common processing logic with adapter for interface checks
    return p.processApprovalWithAdapter(ctx, resourceKey, myResource, adapter, action, req)
}
```

### 步骤4: 更新HTTP处理器

```go
func (h *MyResourceHandler) approveMyResource(w http.ResponseWriter, r *http.Request) {
    // 解析路径参数
    vars := mux.Vars(r)
    namespace := vars["namespace"]
    name := vars["name"]

    // 解析请求体
    var req CommonApprovalRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // 创建处理器并处理审批
    processor := NewCommonApprovalProcessor(h.client)
    resourceKey := client.ObjectKey{Namespace: namespace, Name: name}
    
    if err := processor.ProcessMyResourceApproval(r.Context(), resourceKey, "approve", req); err != nil {
        // 使用标准错误处理
        switch {
        case err.Error() == ErrResourceNotFound:
            http.Error(w, "MyResource not found", http.StatusNotFound)
        case err.Error() == ErrResourceNotInApprovalState:
            http.Error(w, "MyResource is not in approvaling state", http.StatusBadRequest)
        case err.Error() == ErrApproverRequired || err.Error() == ErrReasonRequired:
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            http.Error(w, "Failed to approve MyResource", http.StatusInternalServerError)
        }
        return
    }

    // 返回成功响应
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":   "MyResource approved successfully",
        "approver":  req.Approver,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}
```

### 步骤5: 更新批量审批支持（可选）

如果需要在批量审批中支持新资源，更新 `approval.go` 中的相关方法：

```go
// 在 listPendingApprovals 中添加
myResources := &opsv1beta1.MyResourceList{}
if err := c.Client.List(ctx, myResources, listOpts...); err == nil {
    for _, resource := range myResources.Items {
        adapter := NewMyResourceApprovalAdapter(&resource)
        if adapter.IsInApprovalState() {
            pendingApprovals = append(pendingApprovals, PendingApproval{
                Type:      "MyResource",
                Namespace: resource.Namespace,
                Name:      resource.Name,
                Reason:    resource.Spec.Reason, // 根据实际字段调整
            })
        }
    }
}

// 在统计方法中添加计数逻辑
```

## 测试指南

### 1. 单元测试结构

```go
var _ = Describe("MyResourceApprovalAdapter", func() {
    It("should correctly identify approval state", func() {
        resource := &opsv1beta1.MyResource{
            Status: opsv1beta1.MyResourceStatus{
                Status: types.MyResourceStatusApprovaling,
            },
        }
        adapter := NewMyResourceApprovalAdapter(resource)
        Expect(adapter.IsInApprovalState()).To(BeTrue())

        resource.Status.Status = types.MyResourceStatusApproved
        Expect(adapter.IsInApprovalState()).To(BeFalse())
    })

    It("should return correct status field name", func() {
        resource := &opsv1beta1.MyResource{}
        adapter := NewMyResourceApprovalAdapter(resource)
        Expect(adapter.GetStatusFieldName()).To(Equal("Status.Status"))
    })
})
```

### 2. 集成测试

```go
Context("when processing MyResource approval", func() {
    var myResource *opsv1beta1.MyResource

    BeforeEach(func() {
        myResource = &opsv1beta1.MyResource{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "test-myresource",
                Namespace: "default",
            },
            Spec: opsv1beta1.MyResourceSpec{
                // 填入必要的字段
            },
            Status: opsv1beta1.MyResourceStatus{
                Status: types.MyResourceStatusApprovaling,
            },
        }
        Expect(k8sClient.Create(ctx, myResource)).To(Succeed())

        // 在测试环境中更新状态
        myResource.Status.Status = types.MyResourceStatusApprovaling
        Expect(k8sClient.Status().Update(ctx, myResource)).To(Succeed())
    })

    AfterEach(func() {
        Expect(k8sClient.Delete(ctx, myResource)).To(Succeed())
    })

    It("should approve MyResource successfully", func() {
        req := CommonApprovalRequest{
            Approver: "test-approver",
            Reason:   "Test reason",
        }

        resourceKey := client.ObjectKey{Namespace: myResource.Namespace, Name: myResource.Name}
        err := processor.ProcessMyResourceApproval(ctx, resourceKey, "approve", req)
        Expect(err).ToNot(HaveOccurred())

        // 验证注解设置
        var updated opsv1beta1.MyResource
        Expect(k8sClient.Get(ctx, resourceKey, &updated)).To(Succeed())

        Expect(updated.Annotations).To(HaveKey(constants.ApprovalDecisionAnnotation))
        Expect(updated.Annotations[constants.ApprovalDecisionAnnotation]).To(Equal("approve"))
        Expect(updated.Annotations[constants.ApprovalOperatorAnnotation]).To(Equal("test-approver"))
        Expect(updated.Annotations[constants.ApprovalReasonAnnotation]).To(Equal("Test reason"))
    })
})
```

## 注解和状态管理

### 审批注解

系统使用以下注解来跟踪审批状态：

- `ops.udesk.cn/approval-decision`: "approve" 或 "reject"
- `ops.udesk.cn/approval-timestamp`: RFC3339格式的时间戳
- `ops.udesk.cn/approval-operator`: 审批者身份
- `ops.udesk.cn/approval-reason`: 审批原因
- `ops.udesk.cn/approval-comment`: 可选的审批评论
- `ops.udesk.cn/approval-processing`: "pending" 表示正在处理

### 状态转换

审批流程遵循以下状态转换：

1. **资源创建** → `Approvaling` 状态
2. **设置审批注解** → 控制器检测到注解
3. **控制器处理** → 状态转换为 `Approved`/`Rejected`
4. **执行操作** → 状态转换为 `Running`/`Failed`

## 最佳实践

### 1. 错误处理
- 使用预定义的错误常量（`ErrResourceNotFound`, `ErrResourceNotInApprovalState` 等）
- 提供有意义的错误消息
- 适当的HTTP状态码

### 2. 类型安全
- 使用适配器模式而不是接口断言
- 分离Kubernetes客户端操作和状态检查逻辑
- 利用Go的类型系统防止运行时错误

### 3. 测试
- 为每个新适配器编写单元测试
- 在测试环境中显式设置资源状态
- 验证注解正确设置

### 4. 性能考虑
- 使用专用处理方法而不是通用方法以获得更好的性能
- 避免不必要的资源获取操作
- 考虑批量操作的性能影响

## 故障排除

### 常见问题

1. **"resource is not in approvaling state" 错误**
   - 检查资源的当前状态
   - 确认适配器的 `IsInApprovalState()` 方法正确实现
   - 在测试环境中确保状态正确设置

2. **注解未设置**
   - 检查资源是否正确实现 `ApprovalResource` 接口
   - 确认Kubernetes客户端有适当的权限
   - 验证资源对象不为nil

3. **类型转换错误**
   - 使用适配器模式而不是直接类型断言
   - 确保适配器正确嵌入了资源类型

### 调试技巧

1. 启用详细日志记录
2. 使用 `GetStatusFieldName()` 方法进行调试
3. 检查资源的实际状态和注解
4. 验证Kubernetes客户端配置

## 未来扩展

### 计划的改进

1. **异步审批处理**: 支持长时间运行的审批流程
2. **审批历史**: 跟踪完整的审批历史记录
3. **权限集成**: 与RBAC系统集成进行权限检查
4. **通知系统**: 与现有通知系统集成
5. **批量操作优化**: 改进大规模批量审批的性能

### 贡献指南

当添加新的资源类型或改进现有功能时：

1. 遵循现有的命名约定
2. 为所有新功能编写测试
3. 更新相关文档
4. 确保向后兼容性
5. 遵循Go最佳实践和项目编码标准

---

## 参考

- [Kubernetes Client-go文档](https://pkg.go.dev/k8s.io/client-go)
- [Controller-runtime文档](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [项目开发指南](./dev.prompt.md)
- [审批架构文档](./approval-architecture.md)
