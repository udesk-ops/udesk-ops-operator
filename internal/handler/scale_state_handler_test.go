package handler

import (
	"testing"
	"time"

	"udesk.cn/ops/internal/types"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name        string
		duration    string
		expected    time.Duration
		expectError bool
	}{
		{
			name:        "Valid duration",
			duration:    "10m",
			expected:    10 * time.Minute,
			expectError: false,
		},
		{
			name:        "Empty duration defaults to 5m",
			duration:    "",
			expected:    5 * time.Minute,
			expectError: false,
		},
		{
			name:        "Invalid duration",
			duration:    "invalid",
			expectError: true,
		},
		{
			name:        "Duration in seconds",
			duration:    "30s",
			expected:    30 * time.Second,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.duration)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected duration %v, got %v", tt.expected, result)
			}
		})
	}
}

// 简化的单元测试 - 主要测试接口和基本逻辑
func TestHandlerInterfaces(t *testing.T) {
	// Test that handlers implement the StateHandler interface
	var _ types.StateHandler = &ApprovalingHandler{}
	var _ types.StateHandler = &ApprovedHandler{}
	var _ types.StateHandler = &PendingHandler{}

	// Test CanTransition methods
	approvalingHandler := &ApprovalingHandler{}
	if !approvalingHandler.CanTransition(types.ScaleStatusApproved) {
		t.Errorf("ApprovalingHandler should be able to transition to Approved")
	}
	if !approvalingHandler.CanTransition(types.ScaleStatusRejected) {
		t.Errorf("ApprovalingHandler should be able to transition to Rejected")
	}
	if approvalingHandler.CanTransition(types.ScaleStatusScaling) {
		t.Errorf("ApprovalingHandler should not be able to transition to Scaling")
	}

	approvedHandler := &ApprovedHandler{}
	if !approvedHandler.CanTransition(types.ScaleStatusScaling) {
		t.Errorf("ApprovedHandler should be able to transition to Scaling")
	}
	if approvedHandler.CanTransition(types.ScaleStatusApproved) {
		t.Errorf("ApprovedHandler should not be able to transition to Approved")
	}
}
