package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	scaletypes "udesk.cn/ops/internal/types"
)

// Constants for approval actions
const (
	ApprovalActionApprove = "approve"
	ApprovalActionReject  = "reject"
)

// Constants for approval processing status
const (
	ApprovalProcessingPending   = "pending"
	ApprovalProcessingCompleted = "completed"
)

// init registers the Approval handler automatically
func init() {
	RegisterHandler("approval", func(k8sClient client.Client) Handler {
		return NewApprovalHandler(k8sClient)
	})
}

// ApprovalHandler handles generic approval workflows
type ApprovalHandler struct {
	client client.Client
}

// NewApprovalHandler creates a new approval handler
func NewApprovalHandler(k8sClient client.Client) *ApprovalHandler {
	return &ApprovalHandler{
		client: k8sClient,
	}
}

// RegisterRoutes registers approval routes to the router
func (h *ApprovalHandler) RegisterRoutes(router *mux.Router, responseWriter ResponseWriter) {
	api := GetAPIRouter(router)

	// Generic approval endpoints
	api.HandleFunc("/approvals/pending", h.withResponseWriter(responseWriter, h.listPendingApprovals)).Methods("GET")
	api.HandleFunc("/approvals/batch", h.withResponseWriter(responseWriter, h.batchApproval)).Methods("POST")
	api.HandleFunc("/approvals/stats", h.withResponseWriter(responseWriter, h.getApprovalStats)).Methods("GET")
}

// ApprovalHandlerFuncWithWriter wrapper type
type ApprovalHandlerFuncWithWriter func(ResponseWriter, http.ResponseWriter, *http.Request)

// withResponseWriter wraps handler functions with ResponseWriter
func (h *ApprovalHandler) withResponseWriter(rw ResponseWriter, handler ApprovalHandlerFuncWithWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(rw, w, r)
	}
}

// PendingApprovalItem represents a pending approval item
type PendingApprovalItem struct {
	Type        string    `json:"type"`
	Namespace   string    `json:"namespace"`
	Name        string    `json:"name"`
	Reason      string    `json:"reason"`
	Duration    string    `json:"duration,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	Priority    string    `json:"priority,omitempty"`
	RequestedBy string    `json:"requestedBy,omitempty"`
	TargetKind  string    `json:"targetKind,omitempty"`
	TargetName  string    `json:"targetName,omitempty"`
}

// BatchApprovalRequest represents a batch approval request
type BatchApprovalRequest struct {
	Items    []ApprovalItem `json:"items"`
	Approver string         `json:"approver"`
	Reason   string         `json:"reason"`
	Action   string         `json:"action"` // ApprovalActionApprove or ApprovalActionReject
}

// ApprovalItem represents an item to be approved/rejected
type ApprovalItem struct {
	Type      string `json:"type"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// ApprovalStats represents approval statistics
type ApprovalStats struct {
	TotalPending  int               `json:"totalPending"`
	TotalApproved int               `json:"totalApproved"`
	TotalRejected int               `json:"totalRejected"`
	AlertScales   ApprovalTypeStats `json:"alertScales"`
}

// ApprovalTypeStats represents statistics for a specific approval type
type ApprovalTypeStats struct {
	Pending  int `json:"pending"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
}

// listPendingApprovals handles GET /api/v1/approvals/pending
func (h *ApprovalHandler) listPendingApprovals(responseWriter ResponseWriter, w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	log := logf.FromContext(ctx)

	var pendingItems []PendingApprovalItem

	// Query AlertScales that need approval
	var alertScaleList opsv1beta1.AlertScaleList
	if err := h.client.List(ctx, &alertScaleList); err != nil {
		log.Error(err, "Failed to list AlertScales for pending approvals")
		responseWriter.WriteError(w, http.StatusInternalServerError, "Failed to list pending approvals", err)
		return
	}

	for _, alertScale := range alertScaleList.Items {
		if alertScale.Status.ScaleStatus.Status == scaletypes.ScaleStatusApprovaling {
			item := PendingApprovalItem{
				Type:       "AlertScale",
				Namespace:  alertScale.Namespace,
				Name:       alertScale.Name,
				Reason:     alertScale.Spec.ScaleReason,
				Duration:   alertScale.Spec.ScaleDuration,
				CreatedAt:  alertScale.CreationTimestamp.Time,
				TargetKind: alertScale.Spec.ScaleTarget.Kind,
				TargetName: alertScale.Spec.ScaleTarget.Name,
			}

			// Extract priority from annotations if available
			if priority, exists := alertScale.Annotations["ops.udesk.cn/priority"]; exists {
				item.Priority = priority
			}

			// Extract requester from annotations if available
			if requester, exists := alertScale.Annotations["ops.udesk.cn/requested-by"]; exists {
				item.RequestedBy = requester
			}

			pendingItems = append(pendingItems, item)
		}
	}

	responseData := map[string]interface{}{
		"items": pendingItems,
		"count": len(pendingItems),
	}

	responseWriter.WriteSuccess(w, "Pending approvals retrieved successfully", responseData)
}

// batchApproval handles POST /api/v1/approvals/batch
func (h *ApprovalHandler) batchApproval(responseWriter ResponseWriter, w http.ResponseWriter, r *http.Request) {
	var req BatchApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWriter.WriteError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Approver == "" {
		responseWriter.WriteError(w, http.StatusBadRequest, "Approver is required", nil)
		return
	}

	if req.Action != ApprovalActionApprove && req.Action != ApprovalActionReject {
		responseWriter.WriteError(w, http.StatusBadRequest, "Action must be either 'approve' or 'reject'", nil)
		return
	}

	ctx := context.Background()
	log := logf.FromContext(ctx)

	var results []map[string]interface{}
	successCount := 0
	failureCount := 0

	for _, item := range req.Items {
		switch item.Type {
		case "AlertScale":
			success := h.processAlertScaleApproval(ctx, item, req.Approver, req.Reason, req.Action)
			result := map[string]interface{}{
				"type":      item.Type,
				"namespace": item.Namespace,
				"name":      item.Name,
				"success":   success,
				"action":    req.Action,
			}

			if success {
				successCount++
			} else {
				failureCount++
				result["error"] = "Failed to process approval"
			}

			results = append(results, result)

		default:
			log.Info("Unsupported approval type", "type", item.Type)
			result := map[string]interface{}{
				"type":      item.Type,
				"namespace": item.Namespace,
				"name":      item.Name,
				"success":   false,
				"error":     "Unsupported approval type",
			}
			failureCount++
			results = append(results, result)
		}
	}

	responseData := map[string]interface{}{
		"results":    results,
		"total":      len(req.Items),
		"successful": successCount,
		"failed":     failureCount,
		"action":     req.Action,
		"approver":   req.Approver,
	}

	if failureCount > 0 {
		responseWriter.WriteResponse(w, http.StatusPartialContent, true, "Batch approval completed with some failures", responseData, "")
	} else {
		responseWriter.WriteSuccess(w, "Batch approval completed successfully", responseData)
	}
}

// processAlertScaleApproval processes approval for AlertScale resources using annotation-based declarative approach
func (h *ApprovalHandler) processAlertScaleApproval(ctx context.Context, item ApprovalItem, approver, reason, action string) bool {
	log := logf.FromContext(ctx)

	var alertScale opsv1beta1.AlertScale
	key := client.ObjectKey{
		Namespace: item.Namespace,
		Name:      item.Name,
	}

	if err := h.client.Get(ctx, key, &alertScale); err != nil {
		log.Error(err, "Failed to get AlertScale for batch approval", "namespace", item.Namespace, "name", item.Name)
		return false
	}

	// Declarative approach: Only update annotations, controller will reconcile the desired state
	if alertScale.Annotations == nil {
		alertScale.Annotations = make(map[string]string)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Set approval decision annotations - controller will detect and process
	alertScale.Annotations["ops.udesk.cn/approval-decision"] = action // "approve" or "reject"
	alertScale.Annotations["ops.udesk.cn/approval-timestamp"] = timestamp
	alertScale.Annotations["ops.udesk.cn/approval-operator"] = approver
	alertScale.Annotations["ops.udesk.cn/approval-reason"] = reason

	// Add processing state to prevent duplicate processing
	alertScale.Annotations["ops.udesk.cn/approval-processing"] = ApprovalProcessingPending

	// Single atomic update - no status changes here
	if err := h.client.Update(ctx, &alertScale); err != nil {
		log.Error(err, "Failed to update AlertScale approval annotations", "namespace", item.Namespace, "name", item.Name)
		return false
	}

	log.Info("Approval decision recorded, controller will process the status transition",
		"namespace", item.Namespace, "name", item.Name, "decision", action, "approver", approver)

	return true
}

// getApprovalStats handles GET /api/v1/approvals/stats
func (h *ApprovalHandler) getApprovalStats(responseWriter ResponseWriter, w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	log := logf.FromContext(ctx)

	var stats ApprovalStats

	// Get AlertScale statistics
	var alertScaleList opsv1beta1.AlertScaleList
	if err := h.client.List(ctx, &alertScaleList); err != nil {
		log.Error(err, "Failed to list AlertScales for statistics")
		responseWriter.WriteError(w, http.StatusInternalServerError, "Failed to get approval statistics", err)
		return
	}

	for _, alertScale := range alertScaleList.Items {
		switch alertScale.Status.ScaleStatus.Status {
		case scaletypes.ScaleStatusApprovaling:
			stats.AlertScales.Pending++
			stats.TotalPending++
		case scaletypes.ScaleStatusApproved:
			stats.AlertScales.Approved++
			stats.TotalApproved++
		case scaletypes.ScaleStatusRejected:
			stats.AlertScales.Rejected++
			stats.TotalRejected++
		}
	}

	responseWriter.WriteSuccess(w, "Approval statistics retrieved successfully", stats)
}
