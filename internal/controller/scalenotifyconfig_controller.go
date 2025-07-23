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

var hasDefaultNotifyClient map[string]bool

func init() {
	hasDefaultNotifyClient = make(map[string]bool) // Initialize the map to track default notify clients
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

	// 如果这是一个默认配置且状态为Valid，设置相应的通知客户端
	if config.Spec.Default && config.Status.ValidationStatus == types.ValidationStatusValid {
		if _, exists := hasDefaultNotifyClient[config.Spec.Type]; !exists {
			log.Info("Setting up default ScaleNotifyClient for type", "type", config.Spec.Type, "name", config.Name)
			var notifyClient types.ScaleNotifyClient
			switch config.Spec.Type {
			case types.NotifyTypeWXWorkRobot:
				notifyClient = &strategy.WXWorkRobotNotificationClient{}
				strategy.DefaultNotifyClient = notifyClient
			case types.NotifyTypeEmail:
				notifyClient = &strategy.EmailNotificationClient{}
				strategy.DefaultNotifyClient = notifyClient
			default:
				log.Error(nil, "Unsupported notification type", "type", config.Spec.Type)
				// 标记配置为invalid
				config.Status.ValidationStatus = types.ValidationStatusInvalid
				if updateErr := r.Status().Update(ctx, &config); updateErr != nil {
					log.Error(updateErr, "Failed to update config status", "name", config.Name)
				}
				return ctrl.Result{}, nil
			}

			// 验证配置
			if err := notifyClient.Validate(ctx); err != nil {
				log.Error(err, "Invalid configuration for ScaleNotifyClient", "type", config.Spec.Type)
				// 标记配置为invalid
				config.Status.ValidationStatus = types.ValidationStatusInvalid
				if updateErr := r.Status().Update(ctx, &config); updateErr != nil {
					log.Error(updateErr, "Failed to update config status", "name", config.Name)
				}
				return ctrl.Result{}, nil
			}

			// 标记配置为valid
			if config.Status.ValidationStatus != types.ValidationStatusValid {
				config.Status.ValidationStatus = types.ValidationStatusValid
				if err := r.Status().Update(ctx, &config); err != nil {
					log.Error(err, "Failed to update config status", "name", config.Name)
					return ctrl.Result{}, err
				}
			}

			log.Info("Default ScaleNotifyClient set up successfully", "type", config.Spec.Type)
			hasDefaultNotifyClient[config.Spec.Type] = true
		} else {
			log.Info("Default ScaleNotifyClient already exists for type", "type", config.Spec.Type)
		}
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
