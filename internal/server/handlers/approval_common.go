package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/constants"
)

// Error constants for approval processing
const (
	ErrResourceNotFound           = "resource not found"
	ErrResourceNotInApprovalState = "resource is not in approvaling state"
	ErrApproverRequired           = "approver is required"
	ErrReasonRequired             = "reason is required"
)

// CommonApprovalRequest represents a generic approval/rejection request
type CommonApprovalRequest struct {
	Approver string `json:"approver"`
	Reason   string `json:"reason"`
	Comment  string `json:"comment,omitempty"`
}

// ApprovalResource defines the interface that resources need to implement
// to be eligible for the common approval workflow
type ApprovalResource interface {
	client.Object
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
}

// ApprovalStatusChecker defines how to check if a resource is in approvaling state
type ApprovalStatusChecker interface {
	IsInApprovalState() bool
	GetStatusFieldName() string
}

// CommonApprovalProcessor handles the common approval/rejection logic
type CommonApprovalProcessor struct {
	client client.Client
}

// NewCommonApprovalProcessor creates a new common approval processor
func NewCommonApprovalProcessor(k8sClient client.Client) *CommonApprovalProcessor {
	return &CommonApprovalProcessor{
		client: k8sClient,
	}
}

// ProcessApprovalRequest processes a generic approval/rejection request
func (p *CommonApprovalProcessor) ProcessApprovalRequest(
	ctx context.Context,
	resourceKey client.ObjectKey,
	resourceObj ApprovalResource,
	action string, // "approve" or "reject"
	req CommonApprovalRequest,
) error {
	log := logf.FromContext(ctx).WithName("common-approval-processor")

	// Validate action
	if action != "approve" && action != "reject" {
		return fmt.Errorf("invalid action: %s, must be 'approve' or 'reject'", action)
	}

	// Validate request
	if req.Approver == "" {
		return fmt.Errorf("%s", ErrApproverRequired)
	}
	if req.Reason == "" {
		return fmt.Errorf("%s", ErrReasonRequired)
	}

	// Get the current resource state
	if err := p.client.Get(ctx, resourceKey, resourceObj); err != nil {
		log.Error(err, "Failed to get resource for approval",
			"resourceType", fmt.Sprintf("%T", resourceObj),
			"namespace", resourceKey.Namespace,
			"name", resourceKey.Name)
		return fmt.Errorf("failed to get resource: %w", err)
	}

	// Check if resource implements status checker interface
	if statusChecker, ok := resourceObj.(ApprovalStatusChecker); ok {
		if !statusChecker.IsInApprovalState() {
			return fmt.Errorf("%s", ErrResourceNotInApprovalState)
		}
	}

	// Prepare annotations
	annotations := resourceObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Set approval/rejection decision annotations
	annotations[constants.ApprovalDecisionAnnotation] = action
	annotations[constants.ApprovalTimestampAnnotation] = timestamp
	annotations[constants.ApprovalOperatorAnnotation] = req.Approver
	annotations[constants.ApprovalReasonAnnotation] = req.Reason
	if req.Comment != "" {
		annotations[constants.ApprovalCommentAnnotation] = req.Comment
	}
	annotations[constants.ApprovalProcessingAnnotation] = ApprovalProcessingPending

	// Update annotations
	resourceObj.SetAnnotations(annotations)

	// Single atomic update - controller will handle status transitions
	if err := p.client.Update(ctx, resourceObj); err != nil {
		log.Error(err, "Failed to update resource with approval decision",
			"resourceType", fmt.Sprintf("%T", resourceObj),
			"namespace", resourceKey.Namespace,
			"name", resourceKey.Name,
			"action", action)
		return fmt.Errorf("failed to update resource: %w", err)
	}

	log.Info("Approval decision recorded, controller will process the status transition",
		"resourceType", fmt.Sprintf("%T", resourceObj),
		"namespace", resourceKey.Namespace,
		"name", resourceKey.Name,
		"action", action,
		"approver", req.Approver)

	return nil
}

// HandleApprovalHTTPRequest provides a common HTTP handler for approval/rejection
func (p *CommonApprovalProcessor) HandleApprovalHTTPRequest(
	w http.ResponseWriter,
	r *http.Request,
	resourceKey client.ObjectKey,
	resourceObj ApprovalResource,
	action string,
	responseWriter ResponseWriter,
) {
	ctx := r.Context()

	// Parse request body
	var req CommonApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if responseWriter != nil {
			responseWriter.WriteError(w, http.StatusBadRequest, "Invalid request body", err)
		} else {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
		}
		return
	}

	// Process the approval request
	if err := p.ProcessApprovalRequest(ctx, resourceKey, resourceObj, action, req); err != nil {
		if responseWriter != nil {
			switch {
			case err.Error() == ErrResourceNotFound:
				responseWriter.WriteError(w, http.StatusNotFound, "Resource not found", err)
			case err.Error() == ErrResourceNotInApprovalState:
				responseWriter.WriteError(w, http.StatusBadRequest, "Resource is not in approvaling state", err)
			case err.Error() == ErrApproverRequired || err.Error() == ErrReasonRequired:
				responseWriter.WriteError(w, http.StatusBadRequest, err.Error(), nil)
			default:
				responseWriter.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to %s resource", action), err)
			}
		} else {
			switch {
			case err.Error() == ErrResourceNotFound:
				http.Error(w, "Resource not found", http.StatusNotFound)
			case err.Error() == ErrResourceNotInApprovalState:
				http.Error(w, "Resource is not in approvaling state", http.StatusBadRequest)
			case err.Error() == ErrApproverRequired || err.Error() == ErrReasonRequired:
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, fmt.Sprintf("Failed to %s resource", action), http.StatusInternalServerError)
			}
		}
		return
	}

	// Prepare success response
	responseData := map[string]interface{}{
		"namespace": resourceKey.Namespace,
		"name":      resourceKey.Name,
		"status":    capitalizeFirst(action) + "d",
		"approver":  req.Approver,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	successMessage := fmt.Sprintf("Resource %sd successfully", action)

	if responseWriter != nil {
		responseWriter.WriteSuccess(w, successMessage, responseData)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": successMessage,
			"data":    responseData,
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// ProcessAlertScaleApproval processes approval for AlertScale
func (p *CommonApprovalProcessor) ProcessAlertScaleApproval(
	ctx context.Context,
	resourceKey client.ObjectKey,
	action string,
	req CommonApprovalRequest,
) error {
	// Create the actual AlertScale object for k8s operations
	alertScale := &opsv1beta1.AlertScale{}

	// Get the current resource state first
	if err := p.client.Get(ctx, resourceKey, alertScale); err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}

	// Create adapter for status checking
	adapter := NewAlertScaleApprovalAdapter(alertScale)

	// Use the common processing logic with adapter for interface checks
	return p.processApprovalWithAdapter(ctx, resourceKey, alertScale, adapter, action, req)
}

// ProcessPodRebalanceApproval processes approval for PodRebalance
func (p *CommonApprovalProcessor) ProcessPodRebalanceApproval(
	ctx context.Context,
	resourceKey client.ObjectKey,
	action string,
	req CommonApprovalRequest,
) error {
	// Create the actual PodRebalance object for k8s operations
	podRebalance := &opsv1beta1.PodRebalance{}

	// Get the current resource state first
	if err := p.client.Get(ctx, resourceKey, podRebalance); err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}

	// Create adapter for status checking
	adapter := NewPodRebalanceApprovalAdapter(podRebalance)

	// Use the common processing logic with adapter for interface checks
	return p.processApprovalWithAdapter(ctx, resourceKey, podRebalance, adapter, action, req)
}

// processApprovalWithAdapter processes approval with separate k8s object and adapter
func (p *CommonApprovalProcessor) processApprovalWithAdapter(
	ctx context.Context,
	resourceKey client.ObjectKey,
	k8sObj client.Object,
	statusChecker ApprovalStatusChecker,
	action string,
	req CommonApprovalRequest,
) error {
	log := logf.FromContext(ctx).WithName("common-approval-processor")

	// Validate action
	if action != "approve" && action != "reject" {
		return fmt.Errorf("invalid action: %s, must be 'approve' or 'reject'", action)
	}

	// Validate request
	if req.Approver == "" {
		return fmt.Errorf("%s", ErrApproverRequired)
	}
	if req.Reason == "" {
		return fmt.Errorf("%s", ErrReasonRequired)
	}

	// Check if resource is in approval state using the status checker
	if !statusChecker.IsInApprovalState() {
		return fmt.Errorf("%s", ErrResourceNotInApprovalState)
	}

	// Prepare annotations
	annotations := k8sObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Set approval/rejection decision annotations
	annotations[constants.ApprovalDecisionAnnotation] = action
	annotations[constants.ApprovalTimestampAnnotation] = timestamp
	annotations[constants.ApprovalOperatorAnnotation] = req.Approver
	annotations[constants.ApprovalReasonAnnotation] = req.Reason
	if req.Comment != "" {
		annotations[constants.ApprovalCommentAnnotation] = req.Comment
	}
	annotations[constants.ApprovalProcessingAnnotation] = ApprovalProcessingPending

	// Update annotations
	k8sObj.SetAnnotations(annotations)

	// Single atomic update - controller will handle status transitions
	if err := p.client.Update(ctx, k8sObj); err != nil {
		log.Error(err, "Failed to update resource with approval decision",
			"resourceType", fmt.Sprintf("%T", k8sObj),
			"namespace", resourceKey.Namespace,
			"name", resourceKey.Name,
			"action", action)
		return fmt.Errorf("failed to update resource: %w", err)
	}

	log.Info("Approval decision recorded, controller will process the status transition",
		"resourceType", fmt.Sprintf("%T", k8sObj),
		"namespace", resourceKey.Namespace,
		"name", resourceKey.Name,
		"action", action,
		"approver", req.Approver)

	return nil
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
