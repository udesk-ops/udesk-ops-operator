package types

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

func TestScaleStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ScaleStatusPending", ScaleStatusPending, "Pending"},
		{"ScaleStatusScaling", ScaleStatusScaling, "Scaling"},
		{"ScaleStatusScaled", ScaleStatusScaled, "Scaled"},
		{"ScaleStatusCompleted", ScaleStatusCompleted, "Completed"},
		{"ScaleStatusApprovaling", ScaleStatusApprovaling, "Approvaling"},
		{"ScaleStatusApproved", ScaleStatusApproved, "Approved"},
		{"ScaleStatusRejected", ScaleStatusRejected, "Rejected"},
		{"ScaleStatusFailed", ScaleStatusFailed, "Failed"},
		{"ScaleStatusArchived", ScaleStatusArchived, "Archived"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s to be %s, got %s", tt.name, tt.expected, tt.constant)
			}
		})
	}
}

func TestNotifyTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"NotifyTypeWXWorkRobot", NotifyTypeWXWorkRobot, "WXWorkRobot"},
		{"NotifyTypeEmail", NotifyTypeEmail, "Email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s to be %s, got %s", tt.name, tt.expected, tt.constant)
			}
		})
	}
}

func TestValidationStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ValidationStatusValid", ValidationStatusValid, "Valid"},
		{"ValidationStatusInvalid", ValidationStatusInvalid, "Invalid"},
		{"ValidationStatusPending", ValidationStatusPending, "Pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s to be %s, got %s", tt.name, tt.expected, tt.constant)
			}
		})
	}
}

func TestScaleContext(t *testing.T) {
	ctx := context.Background()
	s := runtime.NewScheme()
	_ = opsv1beta1.AddToScheme(s)
	_ = scheme.AddToScheme(s)
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()

	alertScale := &opsv1beta1.AlertScale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-alertscale",
			Namespace: "default",
		},
		Spec: opsv1beta1.AlertScaleSpec{
			ScaleAutoApproval: true,
		},
	}

	request := ctrl.Request{
		NamespacedName: apitypes.NamespacedName{
			Name:      "test-alertscale",
			Namespace: "default",
		},
	}

	scaleStrategy := &mockScaleStrategy{}

	scaleContext := &ScaleContext{
		Context:       ctx,
		Client:        fakeClient,
		AlertScale:    alertScale,
		Request:       request,
		ScaleStrategy: scaleStrategy,
	}

	// Test that ScaleContext holds all required fields
	if scaleContext.Context == nil {
		t.Errorf("Expected Context to be set")
	}
	if scaleContext.Client == nil {
		t.Errorf("Expected Client to be set")
	}
	if scaleContext.AlertScale == nil {
		t.Errorf("Expected AlertScale to be set")
	}
	if scaleContext.ScaleStrategy == nil {
		t.Errorf("Expected ScaleStrategy to be set")
	}

	// Test field values
	if scaleContext.AlertScale.Name != "test-alertscale" {
		t.Errorf("Expected AlertScale name to be 'test-alertscale', got %s", scaleContext.AlertScale.Name)
	}
	if scaleContext.Request.Name != "test-alertscale" {
		t.Errorf("Expected Request name to be 'test-alertscale', got %s", scaleContext.Request.Name)
	}
}

// Mock implementations for testing interfaces

type mockScaleStrategy struct {
	scaleError             error
	currentReplicas        int32
	currentReplicasError   error
	availableReplicas      int32
	availableReplicasError error
}

func (m *mockScaleStrategy) Scale(ctx context.Context, c client.Client, target *opsv1beta1.ScaleTarget, replicas int32) error {
	return m.scaleError
}

func (m *mockScaleStrategy) GetCurrentReplicas(ctx context.Context, c client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	return m.currentReplicas, m.currentReplicasError
}

func (m *mockScaleStrategy) GetAvailableReplicas(ctx context.Context, c client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	return m.availableReplicas, m.availableReplicasError
}

type mockStateHandler struct {
	handleResult      ctrl.Result
	handleError       error
	canTransitionFunc func(toState string) bool
}

func (m *mockStateHandler) Handle(ctx *ScaleContext) (ctrl.Result, error) {
	return m.handleResult, m.handleError
}

func (m *mockStateHandler) CanTransition(toState string) bool {
	if m.canTransitionFunc != nil {
		return m.canTransitionFunc(toState)
	}
	return true
}

type mockScaleNotifyClient struct {
	sendNotifyError error
	validateError   error
}

func (m *mockScaleNotifyClient) SendNotify(ctx context.Context, message string) error {
	return m.sendNotifyError
}

func (m *mockScaleNotifyClient) Validate(ctx context.Context) error {
	return m.validateError
}

func TestScaleStrategyInterface(t *testing.T) {
	strategy := &mockScaleStrategy{
		currentReplicas:   3,
		availableReplicas: 3,
	}

	ctx := context.Background()
	s := runtime.NewScheme()
	_ = opsv1beta1.AddToScheme(s)
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()

	target := &opsv1beta1.ScaleTarget{
		Kind: "Deployment",
		Name: "test-deployment",
	}

	// Test Scale method
	err := strategy.Scale(ctx, fakeClient, target, 5)
	if err != nil {
		t.Errorf("Expected no error from Scale, got %v", err)
	}

	// Test GetCurrentReplicas method
	replicas, err := strategy.GetCurrentReplicas(ctx, fakeClient, target)
	if err != nil {
		t.Errorf("Expected no error from GetCurrentReplicas, got %v", err)
	}
	if replicas != 3 {
		t.Errorf("Expected current replicas to be 3, got %d", replicas)
	}

	// Test GetAvailableReplicas method
	availableReplicas, err := strategy.GetAvailableReplicas(ctx, fakeClient, target)
	if err != nil {
		t.Errorf("Expected no error from GetAvailableReplicas, got %v", err)
	}
	if availableReplicas != 3 {
		t.Errorf("Expected available replicas to be 3, got %d", availableReplicas)
	}
}

func TestStateHandlerInterface(t *testing.T) {
	handler := &mockStateHandler{
		handleResult: ctrl.Result{Requeue: true},
		canTransitionFunc: func(toState string) bool {
			return toState == ScaleStatusScaling
		},
	}

	// Test Handle method
	scaleContext := &ScaleContext{}
	result, err := handler.Handle(scaleContext)
	if err != nil {
		t.Errorf("Expected no error from Handle, got %v", err)
	}
	if !result.Requeue {
		t.Errorf("Expected result.Requeue to be true")
	}

	// Test CanTransition method
	canTransition := handler.CanTransition(ScaleStatusScaling)
	if !canTransition {
		t.Errorf("Expected CanTransition to return true for Scaling status")
	}

	canTransition = handler.CanTransition(ScaleStatusCompleted)
	if canTransition {
		t.Errorf("Expected CanTransition to return false for Completed status")
	}
}

func TestScaleNotifyClientInterface(t *testing.T) {
	notifyClient := &mockScaleNotifyClient{}

	ctx := context.Background()

	// Test SendNotify method
	err := notifyClient.SendNotify(ctx, "test message")
	if err != nil {
		t.Errorf("Expected no error from SendNotify, got %v", err)
	}

	// Test Validate method
	err = notifyClient.Validate(ctx)
	if err != nil {
		t.Errorf("Expected no error from Validate, got %v", err)
	}
}
