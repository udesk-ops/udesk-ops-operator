package handler

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/constants"
	"udesk.cn/ops/internal/types"
)

func parseDuration(duration string) (time.Duration, error) {
	if duration == "" {
		duration = "5m"
	}
	return time.ParseDuration(duration)
}

// BaseStateHandler 提供通用的状态处理功能
type BaseStateHandler struct{}

// 通用方法
func (h *BaseStateHandler) parseDuration(duration string) (time.Duration, error) {
	return parseDuration(duration)
}

func (h *BaseStateHandler) updateStatus(ctx *types.ScaleContext, status string) error {
	ctx.AlertScale.Status.ScaleStatus.Status = status
	return ctx.Client.Status().Update(ctx.Context, ctx.AlertScale)
}

func (h *BaseStateHandler) sendNotification(ctx *types.ScaleContext, status string) {
	log := logf.FromContext(ctx.Context)
	notificationService := NewNotificationService(ctx.Client)
	if err := notificationService.SendNotification(ctx.Context, ctx, status); err != nil {
		log.Error(err, "Failed to send notification", "status", status)
	}
}

func (h *BaseStateHandler) isTimeout(beginTime metav1.Time, timeoutDuration time.Duration) bool {
	return beginTime.IsZero() || beginTime.Time.Add(timeoutDuration).Before(time.Now())
}

// ApprovalingHandler 处理 Approvaling 状态
type ApprovalingHandler struct {
	BaseStateHandler
}

func (h *ApprovalingHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Handling Approvaling state", "alertScale", ctx.AlertScale.Name)

	// 检查API审批决策
	if result, err := h.processAPIApproval(ctx); result != nil {
		return *result, err
	}

	// 检查自动批准
	if ctx.AlertScale.Spec.ScaleAutoApproval {
		return h.processAutoApproval(ctx)
	}

	// 检查超时
	return h.processTimeout(ctx)
}

func (h *ApprovalingHandler) processAPIApproval(ctx *types.ScaleContext) (*ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)

	decision, exists := ctx.AlertScale.Annotations[constants.ApprovalDecisionAnnotation]
	if !exists {
		return nil, nil
	}

	processing := ctx.AlertScale.Annotations[constants.ApprovalProcessingAnnotation]
	if processing != "pending" {
		return nil, nil
	}

	log.Info("Processing API approval decision", "decision", decision, "alertScale", ctx.AlertScale.Name)

	var newStatus string
	switch decision {
	case "approve":
		newStatus = types.ScaleStatusApproved
		log.Info("API approval processed: Approved", "alertScale", ctx.AlertScale.Name)
	case "reject":
		newStatus = types.ScaleStatusRejected
		log.Info("API approval processed: Rejected", "alertScale", ctx.AlertScale.Name)
	default:
		log.Error(nil, "Unknown approval decision", "decision", decision)
		result := ctrl.Result{RequeueAfter: time.Second * 10}
		return &result, nil
	}

	// 更新状态
	if err := h.updateStatus(ctx, newStatus); err != nil {
		log.Error(err, "Failed to update status after API approval", "decision", decision)
		return &ctrl.Result{}, err
	}

	// 标记处理完成
	if err := h.markApprovalCompleted(ctx); err != nil {
		log.Error(err, "Failed to mark approval as completed")
	}

	result := ctrl.Result{Requeue: true}
	return &result, nil
}

func (h *ApprovalingHandler) processAutoApproval(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Auto approval enabled, transitioning to Approved state")

	if err := h.updateStatus(ctx, types.ScaleStatusApproved); err != nil {
		log.Error(err, "Failed to update status to Approved")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (h *ApprovalingHandler) processTimeout(ctx *types.ScaleContext) (ctrl.Result, error) {
	timeout, err := h.parseDuration(ctx.AlertScale.Spec.ScaleTimeout)
	if err != nil {
		log := logf.FromContext(ctx.Context)
		log.Error(err, "Failed to parse scale timeout duration")
		return ctrl.Result{}, err
	}

	beginTime := ctx.AlertScale.Status.ScaleStatus.ScaleBeginTime
	if beginTime.IsZero() {
		ctx.AlertScale.Status.ScaleStatus.ScaleBeginTime = metav1.Now()
		if err := ctx.Client.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
			return ctrl.Result{}, err
		}
	}

	if h.isTimeout(beginTime, timeout) {
		log := logf.FromContext(ctx.Context)
		log.Info("Approval timeout reached, transitioning to Rejected state")

		if err := h.updateStatus(ctx, types.ScaleStatusRejected); err != nil {
			log.Error(err, "Failed to update status to Rejected")
			return ctrl.Result{}, err
		}

		h.sendNotification(ctx, "rejected")
		return ctrl.Result{}, nil
	}

	log := logf.FromContext(ctx.Context)
	log.Info("Waiting for approval", "alertScale", ctx.AlertScale.Name)
	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

func (h *ApprovalingHandler) markApprovalCompleted(ctx *types.ScaleContext) error {
	if ctx.AlertScale.Annotations == nil {
		ctx.AlertScale.Annotations = make(map[string]string)
	}
	ctx.AlertScale.Annotations[constants.ApprovalProcessingAnnotation] = "completed"
	return ctx.Client.Update(ctx.Context, ctx.AlertScale)
}

func (h *ApprovalingHandler) CanTransition(toState string) bool {
	return toState == types.ScaleStatusApproved || toState == types.ScaleStatusRejected
}

// ApprovedHandler 处理 Approved 状态
type ApprovedHandler struct {
	BaseStateHandler
}

func (h *ApprovedHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Handling Approved state", "alertScale", ctx.AlertScale.Name)

	if err := h.updateStatus(ctx, types.ScaleStatusScaling); err != nil {
		log.Error(err, "Failed to update status to Scaling")
		return ctrl.Result{}, err
	}

	h.sendNotification(ctx, "approved")

	log.Info("Transitioning to Scaling state", "alertScale", ctx.AlertScale.Name)
	return ctrl.Result{Requeue: true}, nil
}

func (h *ApprovedHandler) CanTransition(toState string) bool {
	return toState == types.ScaleStatusScaling
}

// RejectedHandler 处理 Rejected 状态
type RejectedHandler struct {
	BaseStateHandler
}

func (h *RejectedHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Handling Rejected state", "alertScale", ctx.AlertScale.Name)

	if err := h.updateStatus(ctx, types.ScaleStatusCompleted); err != nil {
		log.Error(err, "Failed to update status to Completed")
		return ctrl.Result{}, err
	}

	h.sendNotification(ctx, types.ScaleStatusRejected)
	return ctrl.Result{}, nil
}

func (h *RejectedHandler) CanTransition(toState string) bool {
	return toState == types.ScaleStatusCompleted
}

// PendingHandler 处理 Pending 状态
type PendingHandler struct {
	BaseStateHandler
}

func (h *PendingHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)

	log.Info("Handling Pending state", "alertScale", ctx.AlertScale.Name)
	// 获取当前副本数作为原始副本数
	originReplicas, err := ctx.ScaleStrategy.GetCurrentReplicas(
		ctx.Context,
		ctx.Client,
		&ctx.AlertScale.Spec.ScaleTarget,
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	status := &ctx.AlertScale.Status.ScaleStatus
	status.OriginReplicas = originReplicas
	status.ScaledReplicas = originReplicas // 初始化为原始副本数
	status.Status = types.ScaleStatusApprovaling

	if err := ctx.Client.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
		log.Error(err, "failed to update status to scaling")
		return ctrl.Result{}, err
	}

	// 发送通知
	h.sendNotification(ctx, "pending")

	log.Info("Transitioning to Approvaling state for AlertScale", "alertScale", ctx.AlertScale.Name)
	// 返回结果，继续处理 Approvaling 状态
	return ctrl.Result{Requeue: true}, nil
}

func (h *PendingHandler) CanTransition(toState string) bool {
	return toState == types.ScaleStatusScaling
}

// ScalingHandler 处理 Scaling 状态
type ScalingHandler struct {
	BaseStateHandler
}

func (h *ScalingHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Handling Scaling state", "alertScale", ctx.AlertScale.Name)

	// 使用策略进行扩缩容
	if err := h.scaleIfNeeded(ctx); err != nil {
		return ctrl.Result{}, err
	}

	// 检查扩缩容是否完成
	if isCompleted, err := h.isScalingCompleted(ctx); err != nil {
		return ctrl.Result{}, err
	} else if isCompleted {
		// 解析持续时间
		duration, err := h.parseDuration(ctx.AlertScale.Spec.ScaleDuration)
		if err != nil {
			return ctrl.Result{}, err
		}
		// 更新状态
		status := &ctx.AlertScale.Status.ScaleStatus
		status.Status = types.ScaleStatusScaled
		status.ScaleBeginTime = metav1.Now()
		status.ScaleEndTime = metav1.NewTime(status.ScaleBeginTime.Add(duration))

		// 发送扩缩容完成通知
		h.sendNotification(ctx, "scaled")
	}

	// 检查是否超时
	currentScaleBeginTime := ctx.AlertScale.Status.ScaleStatus.ScaleBeginTime
	timeoutDuration, err := h.parseDuration(ctx.AlertScale.Spec.ScaleTimeout)
	if err != nil {
		return ctrl.Result{}, err
	}
	if h.isTimeout(currentScaleBeginTime, timeoutDuration) {
		status := &ctx.AlertScale.Status.ScaleStatus
		status.Status = types.ScaleStatusFailed
		status.ScaleEndTime = metav1.Now()
	}

	// 更新扩缩容后的副本数
	if availableReplicas, err := ctx.ScaleStrategy.GetAvailableReplicas(
		ctx.Context,
		ctx.Client,
		&ctx.AlertScale.Spec.ScaleTarget,
	); err != nil {
		log.Error(err, "failed to get available replicas")
	} else {
		ctx.AlertScale.Status.ScaleStatus.ScaledReplicas = availableReplicas
	}

	if err := ctx.Client.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

func (h *ScalingHandler) CanTransition(toState string) bool {
	return toState == types.ScaleStatusScaled || toState == types.ScaleStatusFailed
}

func (h *ScalingHandler) scaleIfNeeded(ctx *types.ScaleContext) error {
	currentReplicas, err := ctx.ScaleStrategy.GetCurrentReplicas(
		ctx.Context,
		ctx.Client,
		&ctx.AlertScale.Spec.ScaleTarget,
	)
	if err != nil {
		return err
	}

	if currentReplicas != ctx.AlertScale.Spec.ScaleThreshold {
		return ctx.ScaleStrategy.Scale(
			ctx.Context,
			ctx.Client,
			&ctx.AlertScale.Spec.ScaleTarget,
			ctx.AlertScale.Spec.ScaleThreshold,
		)
	}
	return nil
}

func (h *ScalingHandler) isScalingCompleted(ctx *types.ScaleContext) (bool, error) {
	availableReplicas, err := ctx.ScaleStrategy.GetAvailableReplicas(
		ctx.Context,
		ctx.Client,
		&ctx.AlertScale.Spec.ScaleTarget,
	)
	if err != nil {
		return false, err
	}

	return availableReplicas == ctx.AlertScale.Spec.ScaleThreshold, nil
}

// ScaledHandler 处理 Scaled 状态
type ScaledHandler struct {
	BaseStateHandler
}

func (h *ScaledHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Handling Scaled state", "alertScale", ctx.AlertScale.Name)

	status := &ctx.AlertScale.Status.ScaleStatus

	// 检查是否到达结束时间
	if status.ScaleEndTime.Time.Before(time.Now()) {
		status.Status = types.ScaleStatusCompleted
		if err := ctx.Client.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
			return ctrl.Result{}, err
		}

		// 发送扩缩容完成通知
		h.sendNotification(ctx, "completed")

		return ctrl.Result{Requeue: true}, nil
	}

	// 等待到结束时间
	if !status.ScaleEndTime.IsZero() && status.ScaleEndTime.After(time.Now()) {
		return ctrl.Result{RequeueAfter: time.Until(status.ScaleEndTime.Time)}, nil
	}

	return ctrl.Result{Requeue: true}, nil
}

func (h *ScaledHandler) CanTransition(toState string) bool {
	return toState == types.ScaleStatusCompleted
}

// CompletedHandler 处理 Completed 状态
type CompletedHandler struct {
	BaseStateHandler
}

func (h *CompletedHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Handling Completed state", "alertScale", ctx.AlertScale.Name)

	status := &ctx.AlertScale.Status.ScaleStatus

	// 检查是否已恢复到原始副本数
	currentReplicas, err := ctx.ScaleStrategy.GetCurrentReplicas(
		ctx.Context,
		ctx.Client,
		&ctx.AlertScale.Spec.ScaleTarget,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentReplicas == status.OriginReplicas {
		status.Status = types.ScaleStatusArchived
		if err := ctx.Client.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
			return ctrl.Result{}, err
		}

		// 发送归档通知
		h.sendNotification(ctx, "archived")

		return ctrl.Result{Requeue: true}, nil
	}

	// 恢复原始副本数
	if err := ctx.ScaleStrategy.Scale(
		ctx.Context,
		ctx.Client,
		&ctx.AlertScale.Spec.ScaleTarget,
		status.OriginReplicas,
	); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 5}, nil
}

func (h *CompletedHandler) CanTransition(toState string) bool {
	return toState == types.ScaleStatusArchived
}

// FailedHandler 处理 Failed 状态
type FailedHandler struct {
	BaseStateHandler
}

func (h *FailedHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Handling Failed state", "alertScale", ctx.AlertScale.Name)

	// 发送失败通知
	h.sendNotification(ctx, "failed")

	// 如果副本数 和原始副本数不一致，恢复原始副本数
	availableReplicas, err := ctx.ScaleStrategy.GetAvailableReplicas(
		ctx.Context,
		ctx.Client,
		&ctx.AlertScale.Spec.ScaleTarget,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	if availableReplicas != ctx.AlertScale.Status.ScaleStatus.OriginReplicas {
		if err := ctx.ScaleStrategy.Scale(
			ctx.Context,
			ctx.Client,
			&ctx.AlertScale.Spec.ScaleTarget,
			ctx.AlertScale.Status.ScaleStatus.OriginReplicas,
		); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (h *FailedHandler) CanTransition(toState string) bool {
	return false
}

// ArchivedHandler 处理 Archived 状态
type ArchivedHandler struct {
	BaseStateHandler
}

func (h *ArchivedHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Handling Archived state", "alertScale", ctx.AlertScale.Name)
	return ctrl.Result{}, nil
}

func (h *ArchivedHandler) CanTransition(toState string) bool {
	return false
}

// DefaultHandler 处理默认/初始化状态
type DefaultHandler struct {
	BaseStateHandler
}

func (h *DefaultHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
	log := logf.FromContext(ctx.Context)
	log.Info("Initializing AlertScale status", "alertScale", ctx.AlertScale.Name)

	ctx.AlertScale.Status.ScaleStatus = opsv1beta1.ScaleStatus{
		Status:         types.ScaleStatusPending,
		ScaleBeginTime: metav1.Now(), // 设置开始时间为当前时间，初始化时开始计算scale超时时间
		ScaleEndTime:   metav1.Time{},
		OriginReplicas: 0,
		ScaledReplicas: 0,
	}

	if err := ctx.Client.Status().Update(ctx.Context, ctx.AlertScale); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (h *DefaultHandler) CanTransition(toState string) bool {
	return toState == types.ScaleStatusPending
}
