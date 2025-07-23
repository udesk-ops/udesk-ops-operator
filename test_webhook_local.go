package main

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	webhookv1beta1 "udesk.cn/ops/internal/webhook/v1beta1"
)

func createRawExtension(config map[string]interface{}) runtime.RawExtension {
	data, _ := json.Marshal(config)
	return runtime.RawExtension{Raw: data}
}

func main() {
	// 设置scheme
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = opsv1beta1.AddToScheme(scheme)

	fmt.Println("=== Testing Webhook Validation Logic Locally ===")

	// 创建fake client用于测试
	// 先创建一个已存在的默认Email配置
	existingEmailConfig := &opsv1beta1.ScaleNotifyConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-email-default",
		},
		Spec: opsv1beta1.ScaleNotifyConfigSpec{
			Type:    "Email",
			Default: true,
			Config: createRawExtension(map[string]interface{}{
				"smtpServer":   "smtp.example.com",
				"smtpPort":     587,
				"smtpUser":     "test@example.com",
				"smtpPassword": "password",
				"fromEmail":    "test@example.com",
				"toEmail":      "admin@example.com",
			}),
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(existingEmailConfig).
		Build()

	// 创建webhook实例
	webhook := &webhookv1beta1.ScaleNotifyConfigCustomValidator{
		Client: fakeClient,
	}

	fmt.Println("1. Testing valid new WXWork default config (should pass)...")
	newWXWorkConfig := &opsv1beta1.ScaleNotifyConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-wxwork-default",
		},
		Spec: opsv1beta1.ScaleNotifyConfigSpec{
			Type:    "WXWorkRobot",
			Default: true,
			Config: createRawExtension(map[string]interface{}{
				"webhookURL": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
			}),
		},
	}

	if _, err := webhook.ValidateCreate(context.Background(), newWXWorkConfig); err != nil {
		fmt.Printf("❌ Validation failed (unexpected): %v\n", err)
	} else {
		fmt.Println("✅ WXWork default config validation passed")
	}

	fmt.Println("2. Testing duplicate Email default config (should fail)...")
	duplicateEmailConfig := &opsv1beta1.ScaleNotifyConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "duplicate-email-default",
		},
		Spec: opsv1beta1.ScaleNotifyConfigSpec{
			Type:    "Email",
			Default: true,
			Config: createRawExtension(map[string]interface{}{
				"smtpServer":   "smtp.example.com",
				"smtpPort":     587,
				"smtpUser":     "test2@example.com",
				"smtpPassword": "password",
				"fromEmail":    "test2@example.com",
				"toEmail":      "admin2@example.com",
			}),
		},
	}

	if _, err := webhook.ValidateCreate(context.Background(), duplicateEmailConfig); err != nil {
		fmt.Printf("✅ Duplicate validation correctly failed: %v\n", err)
	} else {
		fmt.Println("❌ Duplicate validation should have failed but passed")
	}

	fmt.Println("3. Testing non-default Email config (should pass)...")
	nonDefaultEmailConfig := &opsv1beta1.ScaleNotifyConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "non-default-email",
		},
		Spec: opsv1beta1.ScaleNotifyConfigSpec{
			Type:    "Email",
			Default: false,
			Config: createRawExtension(map[string]interface{}{
				"smtpServer":   "smtp.example.com",
				"smtpPort":     587,
				"smtpUser":     "test3@example.com",
				"smtpPassword": "password",
				"fromEmail":    "test3@example.com",
				"toEmail":      "admin3@example.com",
			}),
		},
	}

	if _, err := webhook.ValidateCreate(context.Background(), nonDefaultEmailConfig); err != nil {
		fmt.Printf("❌ Non-default validation failed (unexpected): %v\n", err)
	} else {
		fmt.Println("✅ Non-default Email config validation passed")
	}

	fmt.Println("=== Webhook Validation Test Completed ===")
}
