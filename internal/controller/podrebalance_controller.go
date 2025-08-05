/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/constants"
	"udesk.cn/ops/internal/server/handlers"
)

// PodRebalance status constants
const (
	StatusPending     = "Pending"
	StatusApprovaling = "Approvaling"
	StatusApproved    = "Approved"
	StatusRejected    = "Rejected"
	StatusExecuting   = "Executing"
	StatusCompleted   = "Completed"
	StatusFailed      = "Failed"
)

// PodRebalanceReconciler reconciles a PodRebalance object
type PodRebalanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ops.udesk.cn,resources=podrebalances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ops.udesk.cn,resources=podrebalances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ops.udesk.cn,resources=podrebalances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PodRebalanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 获取 PodRebalance 实例
	var podRebalance opsv1beta1.PodRebalance
	if err := r.Get(ctx, req.NamespacedName, &podRebalance); err != nil {
		if apierrors.IsNotFound(err) {
			// 对象被删除，忽略
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch PodRebalance")
		return ctrl.Result{}, err
	}

	// 处理删除逻辑
	if podRebalance.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &podRebalance)
	}

	// 确保 finalizer 存在
	if !controllerutil.ContainsFinalizer(&podRebalance, "ops.udesk.cn/podrebalance-finalizer") {
		controllerutil.AddFinalizer(&podRebalance, "ops.udesk.cn/podrebalance-finalizer")
		if err := r.Update(ctx, &podRebalance); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// 根据当前状态处理审批流
	switch podRebalance.Status.Status {
	case "":
		// 初始状态，设置为 Pending
		return r.handlePending(ctx, &podRebalance)
	case StatusPending:
		// 等待审批，根据配置决定是否自动审批
		return r.handleApprovaling(ctx, &podRebalance)
	case StatusApprovaling:
		// 审批中，检查审批结果
		return r.handleApprovaling(ctx, &podRebalance)
	case StatusApproved:
		// 已批准，开始执行
		return r.handleExecuting(ctx, &podRebalance)
	case StatusRejected:
		// 已拒绝，等待重新提交或删除
		return ctrl.Result{}, nil
	case StatusExecuting:
		// 执行中，监控执行状态
		return r.handleExecuting(ctx, &podRebalance)
	case StatusCompleted:
		// 已完成
		return ctrl.Result{}, nil
	case StatusFailed:
		// 执行失败
		return ctrl.Result{}, nil
	default:
		// 未知状态，重置为 Pending
		return r.handlePending(ctx, &podRebalance)
	}
}

// handleDeletion 处理删除逻辑
func (r *PodRebalanceReconciler) handleDeletion(ctx context.Context, podRebalance *opsv1beta1.PodRebalance) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 清理资源
	log.Info("cleaning up PodRebalance resources", "name", podRebalance.Name)

	// 移除 finalizer
	controllerutil.RemoveFinalizer(podRebalance, "ops.udesk.cn/podrebalance-finalizer")
	if err := r.Update(ctx, podRebalance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// handlePending 处理初始状态
func (r *PodRebalanceReconciler) handlePending(ctx context.Context, podRebalance *opsv1beta1.PodRebalance) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 设置状态为 Pending
	podRebalance.Status.Status = StatusPending
	podRebalance.Status.Message = "Waiting for approval"

	if err := r.Status().Update(ctx, podRebalance); err != nil {
		log.Error(err, "failed to update PodRebalance status to Pending")
		return ctrl.Result{}, err
	}

	log.Info("PodRebalance set to Pending status", "name", podRebalance.Name)
	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

// handleApprovaling 处理审批状态
func (r *PodRebalanceReconciler) handleApprovaling(ctx context.Context, podRebalance *opsv1beta1.PodRebalance) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 更新状态为 Approvaling（如果还不是）
	if podRebalance.Status.Status != StatusApprovaling {
		podRebalance.Status.Status = StatusApprovaling
		podRebalance.Status.Message = "Under approval review"

		if err := r.Status().Update(ctx, podRebalance); err != nil {
			log.Error(err, "failed to update PodRebalance status to Approvaling")
			return ctrl.Result{}, err
		}
	}

	// 检查是否启用自动审批
	if podRebalance.Spec.AutoApproval {
		log.Info("auto-approval enabled, approving PodRebalance", "name", podRebalance.Name)

		// 设置自动审批注解
		if podRebalance.Annotations == nil {
			podRebalance.Annotations = make(map[string]string)
		}
		podRebalance.Annotations[constants.ApprovalDecisionAnnotation] = constants.ApprovalDecisionApprove
		podRebalance.Annotations[constants.ApprovalOperatorAnnotation] = constants.ApprovalDefaultUser
		podRebalance.Annotations[constants.ApprovalTimestampAnnotation] = metav1.Now().Format(time.RFC3339)
		podRebalance.Annotations[constants.ApprovalReasonAnnotation] = constants.ApprovalReasonAutoApproved

		if err := r.Update(ctx, podRebalance); err != nil {
			log.Error(err, "failed to set auto-approval annotations")
			return ctrl.Result{}, err
		}

		// 直接设置为已批准状态
		podRebalance.Status.Status = StatusApproved
		podRebalance.Status.Message = "Auto-approved by system"

		if err := r.Status().Update(ctx, podRebalance); err != nil {
			log.Error(err, "failed to update PodRebalance status to Approved")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// 检查是否有外部审批决策
	annotations := podRebalance.GetAnnotations()
	if annotations != nil {
		decision, exists := annotations[constants.ApprovalDecisionAnnotation]
		if exists {
			processing := annotations[constants.ApprovalProcessingAnnotation]
			if processing == handlers.ApprovalProcessingPending {
				// 审批决策还在处理中，等待
				return ctrl.Result{RequeueAfter: time.Second * 10}, nil
			}

			switch decision {
			case "approve":
				// 批准
				podRebalance.Status.Status = StatusApproved
				podRebalance.Status.Message = "Approved by " + annotations[constants.ApprovalOperatorAnnotation]

				if err := r.Status().Update(ctx, podRebalance); err != nil {
					log.Error(err, "failed to update PodRebalance status to Approved")
					return ctrl.Result{}, err
				}

				log.Info("PodRebalance approved", "name", podRebalance.Name, "operator", annotations[constants.ApprovalOperatorAnnotation])
				return ctrl.Result{}, nil

			case "reject":
				// 拒绝
				podRebalance.Status.Status = StatusRejected
				podRebalance.Status.Message = "Rejected by " + annotations[constants.ApprovalOperatorAnnotation]

				if err := r.Status().Update(ctx, podRebalance); err != nil {
					log.Error(err, "failed to update PodRebalance status to Rejected")
					return ctrl.Result{}, err
				}

				log.Info("PodRebalance rejected", "name", podRebalance.Name, "operator", annotations[constants.ApprovalOperatorAnnotation])
				return ctrl.Result{}, nil
			}
		}
	}

	// 检查审批超时
	creationTime := podRebalance.CreationTimestamp.Time
	approvalTimeout := 24 * time.Hour // 默认24小时超时
	// 解析 Timeout 字段
	if podRebalance.Spec.Timeout != "" {
		if duration, err := time.ParseDuration(podRebalance.Spec.Timeout); err == nil {
			approvalTimeout = duration
		}
	}

	if time.Since(creationTime) > approvalTimeout {
		log.Info("approval timeout reached, rejecting PodRebalance", "name", podRebalance.Name)

		// 设置超时拒绝状态
		podRebalance.Status.Status = StatusRejected
		podRebalance.Status.Message = "Approval timeout"

		if err := r.Status().Update(ctx, podRebalance); err != nil {
			log.Error(err, "failed to update PodRebalance status to Rejected due to timeout")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// 继续等待审批
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

// handleExecuting 处理执行状态
func (r *PodRebalanceReconciler) handleExecuting(ctx context.Context, podRebalance *opsv1beta1.PodRebalance) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 更新状态为 Executing（如果还不是）
	if podRebalance.Status.Status != StatusExecuting {
		podRebalance.Status.Status = StatusExecuting
		podRebalance.Status.Message = "Executing pod rebalance"

		if err := r.Status().Update(ctx, podRebalance); err != nil {
			log.Error(err, "failed to update PodRebalance status to Executing")
			return ctrl.Result{}, err
		}
	}

	log.Info("executing pod rebalance", "name", podRebalance.Name, "strategy", podRebalance.Spec.Strategy)

	// TODO: 实现具体的 Pod 重平衡逻辑
	// 这里应该包含：
	// 1. 根据策略分析当前 Pod 分布
	// 2. 计算目标分布
	// 3. 执行 Pod 迁移/重新调度
	// 4. 监控执行进度

	// 模拟执行完成
	podRebalance.Status.Status = StatusCompleted
	podRebalance.Status.Message = "Pod rebalance completed successfully"

	if err := r.Status().Update(ctx, podRebalance); err != nil {
		log.Error(err, "failed to update PodRebalance status to Completed")
		return ctrl.Result{}, err
	}

	log.Info("PodRebalance completed", "name", podRebalance.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodRebalanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1beta1.PodRebalance{}).
		Complete(r)
}
