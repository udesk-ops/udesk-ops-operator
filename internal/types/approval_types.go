package types

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

// ApprovableResource 定义支持审批流的资源接口
type ApprovableResource interface {
	client.Object

	// GetAutoApproval 获取是否自动审批
	GetAutoApproval() bool

	// GetTimeout 获取超时设置
	GetTimeout() string

	// GetStatus 获取当前状态
	GetStatus() string

	// SetStatus 设置状态
	SetStatus(status string)

	// GetBeginTime 获取开始时间
	GetBeginTime() *metav1.Time

	// SetBeginTime 设置开始时间
	SetBeginTime(time metav1.Time)
}

// ApprovalContext 通用审批上下文
type ApprovalContext struct {
	Context  context.Context
	Client   client.Client
	Request  ctrl.Request
	Resource ApprovableResource
}

// ApprovalHandler 通用审批处理器接口
type ApprovalHandler interface {
	HandleApprovaling(ctx *ApprovalContext) (ctrl.Result, error)
	HandleApproved(ctx *ApprovalContext) (ctrl.Result, error)
	HandleRejected(ctx *ApprovalContext) (ctrl.Result, error)
}

// PodRebalanceContext Pod自平衡特定上下文
type PodRebalanceContext struct {
	Context      context.Context
	Client       client.Client
	Request      ctrl.Request
	PodRebalance *opsv1beta1.PodRebalance
}

// 状态常量 - PodRebalance
const (
	RebalanceStatusPending     = "Pending"
	RebalanceStatusApprovaling = "Approvaling"
	RebalanceStatusApproved    = "Approved"
	RebalanceStatusRejected    = "Rejected"
	RebalanceStatusExecuting   = "Executing"
	RebalanceStatusCompleted   = "Completed"
	RebalanceStatusFailed      = "Failed"
)
