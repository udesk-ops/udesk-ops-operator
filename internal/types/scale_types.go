package types

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

// ScaleStrategy 定义扩缩容策略接口
type ScaleStrategy interface {
	Scale(ctx context.Context, client client.Client, target *opsv1beta1.ScaleTarget, replicas int32) error
	GetCurrentReplicas(ctx context.Context, client client.Client, target *opsv1beta1.ScaleTarget) (int32, error)
	GetAvailableReplicas(ctx context.Context, client client.Client, target *opsv1beta1.ScaleTarget) (int32, error)
}

// ScaleContext 包含所有状态处理所需的上下文
type ScaleContext struct {
	AlertScale    *opsv1beta1.AlertScale
	Client        client.Client // 使用接口而不是具体类型
	Request       ctrl.Request
	Context       context.Context
	ScaleStrategy ScaleStrategy
}

// StateHandler 定义状态处理接口
type StateHandler interface {
	Handle(ctx *ScaleContext) (ctrl.Result, error)
	CanTransition(toState string) bool
}

// 状态常量
const (
	ScaleStatusPending   = "Pending"
	ScaleStatusScaling   = "Scaling"
	ScaleStatusScaled    = "Scaled"
	ScaleStatusCompleted = "Completed"
	ScaleStatusFailed    = "Failed"
	ScaleStatusArchived  = "Archived"
)

type ScaleNotificationClient interface {
	SendNotification(ctx context.Context, message string) error
	Validate(ctx context.Context) error
}
