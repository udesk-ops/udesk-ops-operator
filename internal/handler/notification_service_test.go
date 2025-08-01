package handler

import (
	"context"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/types"
)

func setupNotificationTestClient() client.Client {
	s := runtime.NewScheme()
	_ = opsv1beta1.AddToScheme(s)
	_ = scheme.AddToScheme(s)
	return fake.NewClientBuilder().WithScheme(s).Build()
}

func TestNotificationService_PrepareTemplateData(t *testing.T) {
	k8sClient := setupNotificationTestClient()
	ns := NewNotificationService(k8sClient)

	// 创建测试用的 ScaleContext
	alertScale := &opsv1beta1.AlertScale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-alert",
			Namespace: "default",
		},
		Spec: opsv1beta1.AlertScaleSpec{
			ScaleReason:       "High CPU Usage",
			ScaleDuration:     "30m",
			ScaleThreshold:    80,
			ScaleTimeout:      "5m",
			ScaleAutoApproval: true,
			ScaleTarget: opsv1beta1.ScaleTarget{
				Name:       "nginx-deployment",
				Kind:       "Deployment",
				Namespace:  "default",
				APIVersion: "apps/v1",
			},
		},
		Status: opsv1beta1.AlertScaleStatus{
			ScaleStatus: opsv1beta1.ScaleStatus{
				Status:         types.ScaleStatusScaling,
				OriginReplicas: 3,
				ScaledReplicas: 5,
				ScaleBeginTime: metav1.NewTime(time.Now().Add(-10 * time.Minute)),
				ScaleEndTime:   metav1.NewTime(time.Now().Add(20 * time.Minute)),
			},
		},
	}

	scaleCtx := &types.ScaleContext{
		Context:    context.Background(),
		Client:     k8sClient,
		AlertScale: alertScale,
	}

	// 测试准备模板数据
	templateData := ns.prepareTemplateData(scaleCtx)

	// 验证数据
	if templateData.ScaleReason != "High CPU Usage" {
		t.Errorf("Expected ScaleReason to be 'High CPU Usage', got %s", templateData.ScaleReason)
	}
	if templateData.ScaleDuration != "30m" {
		t.Errorf("Expected ScaleDuration to be '30m', got %s", templateData.ScaleDuration)
	}
	if templateData.ScaleThreshold != 80 {
		t.Errorf("Expected ScaleThreshold to be 80, got %d", templateData.ScaleThreshold)
	}
	if templateData.ScaleAutoApproval != true {
		t.Errorf("Expected ScaleAutoApproval to be true, got %v", templateData.ScaleAutoApproval)
	}
	if templateData.ScaleTarget.Name != "nginx-deployment" {
		t.Errorf("Expected ScaleTarget.Name to be 'nginx-deployment', got %s", templateData.ScaleTarget.Name)
	}
	if templateData.ScaleTarget.Kind != "Deployment" {
		t.Errorf("Expected ScaleTarget.Kind to be 'Deployment', got %s", templateData.ScaleTarget.Kind)
	}
	if templateData.Status != types.ScaleStatusScaling {
		t.Errorf("Expected Status to be '%s', got %s", types.ScaleStatusScaling, templateData.Status)
	}
	if templateData.OriginReplicas != 3 {
		t.Errorf("Expected OriginReplicas to be 3, got %d", templateData.OriginReplicas)
	}
	if templateData.ScaledReplicas != 5 {
		t.Errorf("Expected ScaledReplicas to be 5, got %d", templateData.ScaledReplicas)
	}
}

func TestNotificationService_RenderDefaultMessage(t *testing.T) {
	k8sClient := setupNotificationTestClient()
	ns := NewNotificationService(k8sClient)

	templateData := &TemplateData{
		ScaleReason:       "High CPU Usage",
		ScaleDuration:     "30m",
		ScaleThreshold:    80,
		ScaleTimeout:      "5m",
		ScaleAutoApproval: true,
		Status:            types.ScaleStatusScaling,
		OriginReplicas:    3,
		ScaledReplicas:    5,
		Timestamp:         time.Now(),
		ScaleBeginTime:    time.Now().Add(-10 * time.Minute),
	}
	templateData.ScaleTarget.Name = "nginx-deployment"
	templateData.ScaleTarget.Kind = "Deployment"
	templateData.ScaleTarget.Namespace = "default"

	message := ns.renderDefaultMessage(templateData)

	// 验证消息包含关键信息
	if !strings.Contains(message, "nginx-deployment") {
		t.Error("Message should contain target name")
	}
	if !strings.Contains(message, "Deployment") {
		t.Error("Message should contain target kind")
	}
	if !strings.Contains(message, "High CPU Usage") {
		t.Error("Message should contain scale reason")
	}
	if !strings.Contains(message, types.ScaleStatusScaling) {
		t.Error("Message should contain status")
	}
	if !strings.Contains(message, "80%") {
		t.Error("Message should contain threshold")
	}
}

func TestNotificationService_RenderWithTemplate(t *testing.T) {
	k8sClient := setupNotificationTestClient()
	ns := NewNotificationService(k8sClient)

	// 创建消息模板
	msgTemplate := &opsv1beta1.ScaleNotifyMsgTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-template",
			Namespace: "default",
		},
		Spec: opsv1beta1.ScaleNotifyMsgTemplateSpec{
			Title:   "Scale Alert: {{.ScaleTarget.Name}}",
			Content: "Target: {{.ScaleTarget.Kind}}/{{.ScaleTarget.Name}}\nReason: {{.ScaleReason}}\nStatus: {{.Status}}",
		},
	}

	// 将模板添加到 fake client
	if err := k8sClient.Create(context.Background(), msgTemplate); err != nil {
		t.Fatalf("Failed to create message template: %v", err)
	}

	// 创建测试用的 ScaleContext
	alertScale := &opsv1beta1.AlertScale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-alert",
			Namespace: "default",
		},
		Spec: opsv1beta1.AlertScaleSpec{
			ScaleReason:            "High CPU Usage",
			ScaleNotifyMsgTemplate: "test-template",
			ScaleTarget: opsv1beta1.ScaleTarget{
				Name: "nginx-deployment",
				Kind: "Deployment",
			},
		},
		Status: opsv1beta1.AlertScaleStatus{
			ScaleStatus: opsv1beta1.ScaleStatus{
				Status: types.ScaleStatusScaling,
			},
		},
	}

	scaleCtx := &types.ScaleContext{
		Context:    context.Background(),
		Client:     k8sClient,
		AlertScale: alertScale,
	}

	templateData := ns.prepareTemplateData(scaleCtx)

	// 测试模板渲染
	message, err := ns.renderWithTemplate(context.Background(), scaleCtx, templateData)
	if err != nil {
		t.Fatalf("Failed to render message with template: %v", err)
	}

	// 验证渲染结果
	if !strings.Contains(message, "Scale Alert: nginx-deployment") {
		t.Error("Message should contain rendered title")
	}
	if !strings.Contains(message, "Target: Deployment/nginx-deployment") {
		t.Error("Message should contain rendered target info")
	}
	if !strings.Contains(message, "Reason: High CPU Usage") {
		t.Error("Message should contain rendered reason")
	}
	if !strings.Contains(message, "Status: "+types.ScaleStatusScaling) {
		t.Error("Message should contain rendered status")
	}
}

func TestNotificationService_SendNotification_NoClient(t *testing.T) {
	k8sClient := setupNotificationTestClient()
	ns := NewNotificationService(k8sClient)

	// 创建没有通知类型的 AlertScale
	alertScale := &opsv1beta1.AlertScale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-alert",
			Namespace: "default",
		},
		Spec: opsv1beta1.AlertScaleSpec{
			ScaleNotificationType: "", // 没有配置通知类型
		},
	}

	scaleCtx := &types.ScaleContext{
		Context:    context.Background(),
		Client:     k8sClient,
		AlertScale: alertScale,
	}

	// 应该不会报错，只是跳过通知
	err := ns.SendNotification(context.Background(), scaleCtx, "test")
	if err != nil {
		t.Errorf("Expected no error when no notification type is configured, got %v", err)
	}
}
