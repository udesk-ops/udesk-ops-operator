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

	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

type scaleStatus string

const (
	ScaleStatusPending   scaleStatus = "Pending"
	ScaleStatusScaling   scaleStatus = "Scaling"
	ScaleStatusScaled    scaleStatus = "Scaled"
	ScaleStatusFailed    scaleStatus = "Failed"
	ScaleStatusCompleted scaleStatus = "Completed"
	ScaleStatusArchived  scaleStatus = "Archived"
)

type scaleTargetKind string

const (
	ScaleTargetKindDeployment  scaleTargetKind = "Deployment"
	ScaleTargetKindStatefulSet scaleTargetKind = "StatefulSet"
)

// AlertScaleReconciler reconciles a AlertScale object
type AlertScaleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ops.udesk.cn,resources=alertscales,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ops.udesk.cn,resources=alertscales/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ops.udesk.cn,resources=alertscales/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AlertScale object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *AlertScaleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := logf.FromContext(ctx)

	// Fetch the AlertScale resource
	alertScale := &opsv1beta1.AlertScale{}
	if err := r.Get(ctx, req.NamespacedName, alertScale); err != nil {
		log.Error(err, "unable to fetch AlertScale")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	currentStatus := &alertScale.Status.ScaleStatus
	currentTarget := alertScale.Spec.ScaleTarget
	currentTargetKind := currentTarget.Kind

	switch scaleTargetKind(currentTargetKind) {
	case ScaleTargetKindDeployment:
		// Handle Deployment scaling
		// Use the apps/v1beta1 scale client to scale the Deployment
		// get the deployment resource
		deployment := &appv1.Deployment{}
		if err := r.Get(ctx, client.ObjectKey{Name: currentTarget.Name, Namespace: req.Namespace}, deployment); err != nil {
			log.Error(err, "unable to fetch Deployment for scaling", "name", currentTarget.Name, "namespace", req.Namespace)
			return ctrl.Result{}, err
		}
		switch scaleStatus(currentStatus.Status) {
		case ScaleStatusPending:
			log.Info("ScaleStatusPending: AlertScale is pending", "name", alertScale.Name, "status", currentStatus)
			scaleDuration := alertScale.Spec.ScaleDuration
			if scaleDuration == "" {
				log.Info("ScaleStatusPending: No scale duration specified, defaulting to 5 minutes")
				scaleDuration = "5m" // Default scale duration
			}
			// trans 5m to time.Duration
			duration, err := time.ParseDuration(scaleDuration)
			if err != nil {
				log.Error(err, "ScaleStatusPending: unable to parse scale duration", "duration", scaleDuration)
				return ctrl.Result{}, err
			}

			// begin to process the scaling operation
			currentStatus.Status = string(ScaleStatusScaling)
			currentStatus.ScaleBeginTime = metav1.Now()
			currentStatus.ScaleEndTime = metav1.NewTime(currentStatus.ScaleBeginTime.Time.Add(duration))
			currentStatus.OriginReplicas = *deployment.Spec.Replicas

			if err := r.Status().Update(ctx, alertScale); err != nil {
				log.Error(err, "ScaleStatusPending: unable to update AlertScale status to scaling", "name", alertScale.Name)
				return ctrl.Result{}, err
			}
			// 状态更新后立即重新处理
			return ctrl.Result{Requeue: true}, nil
		case ScaleStatusScaling:
			// Check if the scaling operation is still in progress
			log.Info("ScaleStatusScaling: Scaling Deployment", "name", deployment.Name, "namespace", req.Namespace, "status", currentStatus)
			if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas != alertScale.Spec.ScaleThreshold {
				patch := client.MergeFrom(deployment.DeepCopy())
				// Update the Deployment replicas to the scale threshold
				log.Info("ScaleStatusScaling: Updating Deployment replicas", "name", deployment.Name, "namespace", req.Namespace, "scaleThreshold", alertScale.Spec.ScaleThreshold)

				deployment.Spec.Replicas = &alertScale.Spec.ScaleThreshold

				if err := r.Patch(ctx, deployment, patch); err != nil {
					log.Error(err, "ScaleStatusScaling: unable to patch Deployment for scaling", "name", deployment.Name, "namespace", req.Namespace)
					return ctrl.Result{}, err
				}
				log.Info("ScaleStatusScaling: Deployment replicas updated", "name", deployment.Name, "namespace", req.Namespace, "scaleThreshold", alertScale.Spec.ScaleThreshold)
			}

			// Check if the scaling operation is completed
			if deployment.Status.AvailableReplicas == alertScale.Spec.ScaleThreshold {
				currentStatus.Status = string(ScaleStatusScaled)
			}

			// update the current status with scaled replicas
			currentStatus.ScaledReplicas = deployment.Status.AvailableReplicas

			// Update the AlertScale status
			if err := r.Status().Update(ctx, alertScale); err != nil {
				log.Error(err, "ScaleStatusScaling: unable to update AlertScale status to scaled", "name", alertScale.Name)
				return ctrl.Result{}, err
			}
			// 每10秒检查一次扩缩容状态
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		case ScaleStatusScaled:
			log.Info("ScaleStatusScaled: AlertScale has been scaled", "name", alertScale.Name, "status", currentStatus)
			// check if the scaling operation is completed
			// if the scaleEndTime is before time.NOW(), modify the status to completed
			if currentStatus.ScaleEndTime.Time.Before(time.Now()) {
				currentStatus.Status = string(ScaleStatusCompleted)
				if err := r.Status().Update(ctx, alertScale); err != nil {
					log.Error(err, "unable to update AlertScale status to completed", "name", alertScale.Name)
					return ctrl.Result{}, err
				}
				log.Info("ScaleStatusScaled: AlertScale scaling completed", "name", alertScale.Name)
			}
			// 等待到扩缩容结束时间
			if !currentStatus.ScaleEndTime.Time.IsZero() && currentStatus.ScaleEndTime.Time.After(time.Now()) {
				return ctrl.Result{RequeueAfter: time.Until(currentStatus.ScaleEndTime.Time)}, nil
			}
			// 立即转换为 Completed 状态
			return ctrl.Result{Requeue: true}, nil
		case ScaleStatusCompleted:
			log.Info("ScaleStatusCompleted: AlertScale scaling is already completed", "name", alertScale.Name, "status", currentStatus)

			// 安全地检查副本数是否已恢复
			if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == currentStatus.OriginReplicas {
				// Update the AlertScale status to archived
				currentStatus.Status = string(ScaleStatusArchived)
				if err := r.Status().Update(ctx, alertScale); err != nil {
					log.Error(err, "unable to update AlertScale status to archived", "name", alertScale.Name)
					return ctrl.Result{}, err
				}
				log.Info("ScaleStatusCompleted: AlertScale scaling has been archived", "name", alertScale.Name)
				return ctrl.Result{Requeue: true}, nil
			} else {
				// 使用 Patch 更新 Deployment 副本数
				patch := client.MergeFrom(deployment.DeepCopy())
				deployment.Spec.Replicas = &currentStatus.OriginReplicas

				if err := r.Patch(ctx, deployment, patch); err != nil {
					log.Error(err, "unable to patch Deployment to original replicas", "name", deployment.Name, "namespace", req.Namespace)
					return ctrl.Result{}, err
				}

				log.Info("ScaleStatusCompleted: Deployment replicas restored to original", "name", deployment.Name, "original", currentStatus.OriginReplicas)
				// 恢复副本数后重新检查
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}
		case ScaleStatusFailed:
			log.Info("ScaleStatusFailed: AlertScale scaling failed", "name", alertScale.Name)
			// 失败状态不需要重新调度
			return ctrl.Result{}, nil
		case ScaleStatusArchived:
			log.Info("ScaleStatusArchived: AlertScale scaling has been archived", "name", alertScale.Name)
			// 归档状态不需要重新调度
			return ctrl.Result{}, nil
		default:
			alertScale.Status.ScaleStatus = opsv1beta1.ScaleStatus{
				Status:         string(ScaleStatusPending),
				ScaleBeginTime: metav1.Now(),
				ScaleEndTime:   metav1.Time{},
				OriginReplicas: 0,
				ScaledReplicas: 0,
			}
			if err := r.Status().Update(ctx, alertScale); err != nil {
				log.Error(err, "unable to update AlertScale status to pending", "name", alertScale.Name)
				return ctrl.Result{}, err
			}
			log.Info("AlertScale status initialized to pending", "name", alertScale.Name, "status", currentStatus)
			// 状态初始化后立即重新处理
			return ctrl.Result{Requeue: true}, nil
		}

	default:
		log.Info("Unsupported scale target kind", "kind", currentTargetKind, "name", currentTarget.Name)
		return ctrl.Result{}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertScaleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1beta1.AlertScale{}).
		Owns(&appv1.Deployment{}).
		Owns(&appv1.StatefulSet{}).
		Named("alertscale").
		Complete(r)
}
