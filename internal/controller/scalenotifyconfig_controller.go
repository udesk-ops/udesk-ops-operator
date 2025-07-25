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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/strategy"
	"udesk.cn/ops/internal/types"
)

// ScaleNotifyConfigReconciler reconciles a ScaleNotifyConfig object
type ScaleNotifyConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ops.udesk.cn,resources=scalenotifyconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ops.udesk.cn,resources=scalenotifyconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ops.udesk.cn,resources=scalenotifyconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ScaleNotifyConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *ScaleNotifyConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling ScaleNotifyConfig", "request", req)

	// 获取当前的ScaleNotifyConfig
	var config opsv1beta1.ScaleNotifyConfig
	if err := r.Get(ctx, req.NamespacedName, &config); err != nil {
		log.Info("ScaleNotifyConfig not found, possibly deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 初始化状态（如果为空）
	if config.Status.ValidationStatus == "" {
		config.Status.ValidationStatus = types.ValidationStatusPending
		if err := r.Status().Update(ctx, &config); err != nil {
			log.Error(err, "Failed to initialize config status", "name", config.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 验证配置并更新状态
	if config.Status.ValidationStatus == types.ValidationStatusPending {
		if err := r.validateConfigAndUpdateStatus(ctx, &config); err != nil {
			log.Error(err, "Failed to validate config", "name", config.Name)
			// 如果是资源不存在错误，忽略它（可能已被删除）
			if client.IgnoreNotFound(err) == nil {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 如果状态是valid且是默认配置，则可以设置默认通知客户端
	if config.Status.ValidationStatus == types.ValidationStatusValid && config.Spec.Default {
		strategy.DefaultNotifyClientMap[config.Spec.Type] = strategy.NewScaleNotifyClient(config.Spec.Type, config.Spec.Config)
		log.Info("Default notification client set", "name", config.Name, "type", config.Spec.Type)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScaleNotifyConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1beta1.ScaleNotifyConfig{}).
		Named("scalenotifyconfig").
		Complete(r)
}

// validateConfigAndUpdateStatus 验证配置并更新状态
func (r *ScaleNotifyConfigReconciler) validateConfigAndUpdateStatus(ctx context.Context, config *opsv1beta1.ScaleNotifyConfig) error {
	log := logf.FromContext(ctx)

	// // 创建相应的通知客户端进行验证
	var notifyClient types.ScaleNotifyClient
	var err error
	// 根据配置类型创建不同的通知客户端
	switch config.Spec.Type {
	case types.NotifyTypeWXWorkRobot:
		if notifyClient, err = strategy.NewWXWorkRobotNotificationClient(config.Spec.Config); err != nil {
			log.Error(err, "Failed to create WeChat Work notification client", "type", config.Spec.Type)
			config.Status.ValidationStatus = types.ValidationStatusInvalid
			return r.Status().Update(ctx, config)
		}
	case types.NotifyTypeEmail:
		if notifyClient, err = strategy.NewEmailNotificationClient(config.Spec.Config); err != nil {
			log.Error(err, "Failed to create Email notification client", "type", config.Spec.Type)
			config.Status.ValidationStatus = types.ValidationStatusInvalid
			return r.Status().Update(ctx, config)
		}
	default:
		log.Error(nil, "Unsupported notification type", "type", config.Spec.Type)
		config.Status.ValidationStatus = types.ValidationStatusInvalid
		return r.Status().Update(ctx, config)
	}

	// 验证配置
	if err := notifyClient.Validate(ctx); err != nil {
		log.Error(err, "Configuration validation failed", "type", config.Spec.Type)
		config.Status.ValidationStatus = types.ValidationStatusInvalid
		return r.Status().Update(ctx, config)
	}

	// 验证成功
	config.Status.ValidationStatus = types.ValidationStatusValid

	log.Info("Configuration validated successfully", "name", config.Name, "type", config.Spec.Type)
	return r.Status().Update(ctx, config)
}
