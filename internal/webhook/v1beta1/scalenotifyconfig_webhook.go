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

package v1beta1

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/strategy"
)

// nolint:unused
// log is for logging in this package.
var scalenotifyconfiglog = logf.Log.WithName("scalenotifyconfig-resource")

// SetupScaleNotifyConfigWebhookWithManager registers the webhook for ScaleNotifyConfig in the manager.
func SetupScaleNotifyConfigWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsv1beta1.ScaleNotifyConfig{}).
		WithValidator(&ScaleNotifyConfigCustomValidator{Client: mgr.GetClient()}).
		WithDefaulter(&ScaleNotifyConfigCustomDefaulter{Client: mgr.GetClient()}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-ops-udesk-cn-v1beta1-scalenotifyconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=ops.udesk.cn,resources=scalenotifyconfigs,verbs=create;update,versions=v1beta1,name=mscalenotifyconfig-v1beta1.kb.io,admissionReviewVersions=v1

// ScaleNotifyConfigCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ScaleNotifyConfig when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ScaleNotifyConfigCustomDefaulter struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &ScaleNotifyConfigCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ScaleNotifyConfig.
func (d *ScaleNotifyConfigCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	scalenotifyconfig, ok := obj.(*opsv1beta1.ScaleNotifyConfig)

	if !ok {
		return fmt.Errorf("expected an ScaleNotifyConfig object but got %T", obj)
	}
	scalenotifyconfiglog.Info("Defaulting for ScaleNotifyConfig", "name", scalenotifyconfig.GetName())

	// 设置默认的ValidationStatus为Pending
	if scalenotifyconfig.Status.ValidationStatus == "" {
		scalenotifyconfig.Status.ValidationStatus = "Pending"
	}

	// 设置默认的Default为false
	if !scalenotifyconfig.Spec.Default {
		scalenotifyconfig.Spec.Default = false
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-ops-udesk-cn-v1beta1-scalenotifyconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=ops.udesk.cn,resources=scalenotifyconfigs,verbs=create;update,versions=v1beta1,name=vscalenotifyconfig-v1beta1.kb.io,admissionReviewVersions=v1

// ScaleNotifyConfigCustomValidator struct is responsible for validating the ScaleNotifyConfig resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ScaleNotifyConfigCustomValidator struct {
	Client client.Client
}

var _ webhook.CustomValidator = &ScaleNotifyConfigCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ScaleNotifyConfig.
func (v *ScaleNotifyConfigCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scalenotifyconfig, ok := obj.(*opsv1beta1.ScaleNotifyConfig)
	if !ok {
		return nil, fmt.Errorf("expected a ScaleNotifyConfig object but got %T", obj)
	}
	scalenotifyconfiglog.Info("Validation for ScaleNotifyConfig upon creation", "name", scalenotifyconfig.GetName())

	// 验证基本字段
	if err := v.validateBasicFields(scalenotifyconfig, ctx); err != nil {
		return nil, err
	}

	// 如果这个配置要设置为默认，检查是否已经存在同类型的默认配置
	if scalenotifyconfig.Spec.Default {
		if err := v.validateUniqueDefault(ctx, scalenotifyconfig, ""); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ScaleNotifyConfig.
func (v *ScaleNotifyConfigCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scalenotifyconfig, ok := newObj.(*opsv1beta1.ScaleNotifyConfig)
	if !ok {
		return nil, fmt.Errorf("expected a ScaleNotifyConfig object for the newObj but got %T", newObj)
	}

	oldConfig, ok := oldObj.(*opsv1beta1.ScaleNotifyConfig)
	if !ok {
		return nil, fmt.Errorf("expected a ScaleNotifyConfig object for the oldObj but got %T", oldObj)
	}

	scalenotifyconfiglog.Info("Validation for ScaleNotifyConfig upon update", "name", scalenotifyconfig.GetName())

	// 验证基本字段
	if err := v.validateBasicFields(scalenotifyconfig, ctx); err != nil {
		return nil, err
	}

	// 如果新配置要设置为默认，或者类型发生了变化，检查是否已经存在同类型的默认配置
	if scalenotifyconfig.Spec.Default && (!oldConfig.Spec.Default || oldConfig.Spec.Type != scalenotifyconfig.Spec.Type) {
		if err := v.validateUniqueDefault(ctx, scalenotifyconfig, oldConfig.Name); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ScaleNotifyConfig.
func (v *ScaleNotifyConfigCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scalenotifyconfig, ok := obj.(*opsv1beta1.ScaleNotifyConfig)
	if !ok {
		return nil, fmt.Errorf("expected a ScaleNotifyConfig object but got %T", obj)
	}
	scalenotifyconfiglog.Info("Validation for ScaleNotifyConfig upon deletion", "name", scalenotifyconfig.GetName())

	// TODO: 如果需要，可以在此添加删除验证逻辑
	// 例如：检查是否有依赖此配置的其他资源

	return nil, nil
}

// validateBasicFields 验证ScaleNotifyConfig的基本字段
func (v *ScaleNotifyConfigCustomValidator) validateBasicFields(config *opsv1beta1.ScaleNotifyConfig, ctx context.Context) error {
	// 验证Type字段
	if config.Spec.Type == "" {
		return fmt.Errorf("spec.type is required")
	}

	// 验证Type字段的值
	validTypes := []string{"Email", "WXWorkRobot"}
	isValidType := false
	for _, validType := range validTypes {
		if config.Spec.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("spec.type must be one of: %v, got: %s", validTypes, config.Spec.Type)
	}

	// 验证Config字段（如果需要的话）
	if len(config.Spec.Config.Raw) == 0 {
		return fmt.Errorf("spec.config is required and cannot be empty")
	} else {
		// 根据Type字段的值，验证Config的内容
		switch config.Spec.Type {
		case "Email":
			var emailConfig strategy.EmailNotificationClient
			if err := json.Unmarshal(config.Spec.Config.Raw, &emailConfig); err != nil {
				return fmt.Errorf("failed to unmarshal spec.config for Email type: %v", err)
			}
			if err := emailConfig.Validate(ctx); err != nil {
				return fmt.Errorf("invalid Email configuration: %v", err)
			}
		case "WXWorkRobot":
			var wxConfig strategy.WXWorkRobotNotificationClient
			if err := json.Unmarshal(config.Spec.Config.Raw, &wxConfig); err != nil {
				return fmt.Errorf("failed to unmarshal spec.config for WXWorkRobot type: %v", err)
			}
			if err := wxConfig.Validate(ctx); err != nil {
				return fmt.Errorf("invalid WXWorkRobot configuration: %v", err)
			}
		default:
			return fmt.Errorf("unsupported notification type: %s", config.Spec.Type)
		}
	}

	return nil
}

// validateUniqueDefault 验证同一类型只能有一个默认配置
func (v *ScaleNotifyConfigCustomValidator) validateUniqueDefault(ctx context.Context, config *opsv1beta1.ScaleNotifyConfig, excludeName string) error {
	// 列出所有相同类型的ScaleNotifyConfig
	var configList opsv1beta1.ScaleNotifyConfigList
	if err := v.Client.List(ctx, &configList, client.InNamespace(config.Namespace)); err != nil {
		return fmt.Errorf("failed to list ScaleNotifyConfig: %v", err)
	}

	// 检查是否已经存在同类型的默认配置
	for _, existingConfig := range configList.Items {
		// 跳过当前正在更新的配置
		if excludeName != "" && existingConfig.Name == excludeName {
			continue
		}

		// 检查是否存在同类型且为默认的配置
		if existingConfig.Spec.Type == config.Spec.Type && existingConfig.Spec.Default {
			return fmt.Errorf("a default ScaleNotifyConfig of type '%s' already exists: %s",
				config.Spec.Type, existingConfig.Name)
		}
	}

	return nil
}
