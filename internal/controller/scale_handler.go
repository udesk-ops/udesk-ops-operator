package controller

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

// PendingHandler 处理 Pending 状态
type PendingHandler struct{}

func (h *PendingHandler) Handle(ctx *ScaleContext) (ctrl.Result, error) {
	ctx.Logger.Info("Handling Pending state", "alertScale", ctx.AlertScale.Name)

	// 解析持续时间
	duration, err := h.parseDuration(ctx.AlertScale.Spec.ScaleDuration)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 更新状态
	status := &ctx.AlertScale.Status.ScaleStatus
	status.Status = ScaleStatusScaling
	status.ScaleBeginTime = metav1.Now()
	status.ScaleEndTime = metav1.NewTime(status.ScaleBeginTime.Time.Add(duration))
	status.OriginReplicas = *ctx.Deployment.Spec.Replicas

	if err := ctx.Reconciler.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
		ctx.Logger.Error(err, "failed to update status to scaling")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (h *PendingHandler) CanTransition(toState string) bool {
	return toState == ScaleStatusScaling
}

func (h *PendingHandler) parseDuration(duration string) (time.Duration, error) {
	if duration == "" {
		duration = "5m"
	}
	return time.ParseDuration(duration)
}

// ScalingHandler 处理 Scaling 状态
type ScalingHandler struct{}

func (h *ScalingHandler) Handle(ctx *ScaleContext) (ctrl.Result, error) {
	ctx.Logger.Info("Handling Scaling state", "alertScale", ctx.AlertScale.Name)

	// 检查是否需要扩缩容
	if err := h.scaleIfNeeded(ctx); err != nil {
		return ctrl.Result{}, err
	}

	// 检查扩缩容是否完成
	if h.isScalingCompleted(ctx) {
		ctx.AlertScale.Status.ScaleStatus.Status = ScaleStatusScaled
	}

	// 更新扩缩容后的副本数
	ctx.AlertScale.Status.ScaleStatus.ScaledReplicas = ctx.Deployment.Status.AvailableReplicas

	if err := ctx.Reconciler.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

func (h *ScalingHandler) CanTransition(toState string) bool {
	return toState == ScaleStatusScaled || toState == ScaleStatusFailed
}

func (h *ScalingHandler) scaleIfNeeded(ctx *ScaleContext) error {
	if ctx.Deployment.Spec.Replicas != nil &&
		*ctx.Deployment.Spec.Replicas != ctx.AlertScale.Spec.ScaleThreshold {

		patch := client.MergeFrom(ctx.Deployment.DeepCopy())
		ctx.Deployment.Spec.Replicas = &ctx.AlertScale.Spec.ScaleThreshold

		return ctx.Reconciler.Patch(ctx.Context, ctx.Deployment, patch)
	}
	return nil
}

func (h *ScalingHandler) isScalingCompleted(ctx *ScaleContext) bool {
	return ctx.Deployment.Status.AvailableReplicas == ctx.AlertScale.Spec.ScaleThreshold
}

// ScaledHandler 处理 Scaled 状态
type ScaledHandler struct{}

func (h *ScaledHandler) Handle(ctx *ScaleContext) (ctrl.Result, error) {
	ctx.Logger.Info("Handling Scaled state", "alertScale", ctx.AlertScale.Name)

	status := &ctx.AlertScale.Status.ScaleStatus

	// 检查是否到达结束时间
	if status.ScaleEndTime.Time.Before(time.Now()) {
		status.Status = ScaleStatusCompleted
		if err := ctx.Reconciler.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 等待到结束时间
	if !status.ScaleEndTime.Time.IsZero() && status.ScaleEndTime.Time.After(time.Now()) {
		return ctrl.Result{RequeueAfter: time.Until(status.ScaleEndTime.Time)}, nil
	}

	return ctrl.Result{Requeue: true}, nil
}

func (h *ScaledHandler) CanTransition(toState string) bool {
	return toState == ScaleStatusCompleted
}

// CompletedHandler 处理 Completed 状态
type CompletedHandler struct{}

func (h *CompletedHandler) Handle(ctx *ScaleContext) (ctrl.Result, error) {
	ctx.Logger.Info("Handling Completed state", "alertScale", ctx.AlertScale.Name)

	status := &ctx.AlertScale.Status.ScaleStatus

	// 检查是否已恢复到原始副本数
	if ctx.Deployment.Spec.Replicas != nil &&
		*ctx.Deployment.Spec.Replicas == status.OriginReplicas {

		status.Status = ScaleStatusArchived
		if err := ctx.Reconciler.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 恢复原始副本数
	if err := h.restoreOriginalReplicas(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 5}, nil
}

func (h *CompletedHandler) CanTransition(toState string) bool {
	return toState == ScaleStatusArchived
}

func (h *CompletedHandler) restoreOriginalReplicas(ctx *ScaleContext) error {
	patch := client.MergeFrom(ctx.Deployment.DeepCopy())
	ctx.Deployment.Spec.Replicas = &ctx.AlertScale.Status.ScaleStatus.OriginReplicas
	return ctx.Reconciler.Patch(ctx.Context, ctx.Deployment, patch)
}

// FailedHandler 处理 Failed 状态
type FailedHandler struct{}

func (h *FailedHandler) Handle(ctx *ScaleContext) (ctrl.Result, error) {
	ctx.Logger.Info("Handling Failed state", "alertScale", ctx.AlertScale.Name)
	return ctrl.Result{}, nil
}

func (h *FailedHandler) CanTransition(toState string) bool {
	return false
}

// ArchivedHandler 处理 Archived 状态
type ArchivedHandler struct{}

func (h *ArchivedHandler) Handle(ctx *ScaleContext) (ctrl.Result, error) {
	ctx.Logger.Info("Handling Archived state", "alertScale", ctx.AlertScale.Name)
	return ctrl.Result{}, nil
}

func (h *ArchivedHandler) CanTransition(toState string) bool {
	return false
}

// DefaultHandler 处理默认/初始化状态
type DefaultHandler struct{}

func (h *DefaultHandler) Handle(ctx *ScaleContext) (ctrl.Result, error) {
	ctx.Logger.Info("Initializing AlertScale status", "alertScale", ctx.AlertScale.Name)

	ctx.AlertScale.Status.ScaleStatus = opsv1beta1.ScaleStatus{
		Status:         ScaleStatusPending,
		ScaleBeginTime: metav1.Now(),
		ScaleEndTime:   metav1.Time{},
		OriginReplicas: 0,
		ScaledReplicas: 0,
	}

	if err := ctx.Reconciler.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (h *DefaultHandler) CanTransition(toState string) bool {
	return toState == ScaleStatusPending
}
