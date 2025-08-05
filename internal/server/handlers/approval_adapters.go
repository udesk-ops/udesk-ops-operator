package handlers

import (
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/types"
)

// AlertScaleApprovalAdapter adapts AlertScale to work with common approval processor
type AlertScaleApprovalAdapter struct {
	*opsv1beta1.AlertScale
}

// NewAlertScaleApprovalAdapter creates a new adapter for AlertScale
func NewAlertScaleApprovalAdapter(alertScale *opsv1beta1.AlertScale) *AlertScaleApprovalAdapter {
	return &AlertScaleApprovalAdapter{AlertScale: alertScale}
}

// IsInApprovalState checks if AlertScale is in approvaling state
func (a *AlertScaleApprovalAdapter) IsInApprovalState() bool {
	return a.Status.ScaleStatus.Status == types.ScaleStatusApprovaling
}

// GetStatusFieldName returns the status field name for debugging
func (a *AlertScaleApprovalAdapter) GetStatusFieldName() string {
	return "Status.ScaleStatus.Status"
}

// PodRebalanceApprovalAdapter adapts PodRebalance to work with common approval processor
type PodRebalanceApprovalAdapter struct {
	*opsv1beta1.PodRebalance
}

// NewPodRebalanceApprovalAdapter creates a new adapter for PodRebalance
func NewPodRebalanceApprovalAdapter(podRebalance *opsv1beta1.PodRebalance) *PodRebalanceApprovalAdapter {
	return &PodRebalanceApprovalAdapter{PodRebalance: podRebalance}
}

// IsInApprovalState checks if PodRebalance is in approvaling state
func (a *PodRebalanceApprovalAdapter) IsInApprovalState() bool {
	return a.Status.Status == types.RebalanceStatusApprovaling
}

// GetStatusFieldName returns the status field name for debugging
func (a *PodRebalanceApprovalAdapter) GetStatusFieldName() string {
	return "Status.Status"
}
