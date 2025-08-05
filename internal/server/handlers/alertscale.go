package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

// init registers the AlertScale handler automatically
func init() {
	RegisterHandler("alertscale", func(k8sClient client.Client) Handler {
		return NewAlertScaleHandler(k8sClient)
	})
}

// AlertScaleHandler handles AlertScale-related API endpoints
type AlertScaleHandler struct {
	client client.Client
}

// NewAlertScaleHandler creates a new AlertScale handler
func NewAlertScaleHandler(k8sClient client.Client) *AlertScaleHandler {
	return &AlertScaleHandler{
		client: k8sClient,
	}
}

// RegisterRoutes registers AlertScale routes to the router
func (h *AlertScaleHandler) RegisterRoutes(router *mux.Router, responseWriter ResponseWriter) {
	api := GetAPIRouter(router)

	// AlertScale API endpoints
	api.HandleFunc("/alertscales", h.withResponseWriter(responseWriter, h.listAlertScales)).Methods("GET")
	api.HandleFunc("/alertscales/{namespace}/{name}", h.withResponseWriter(responseWriter, h.getAlertScale)).Methods("GET")
	api.HandleFunc("/alertscales/{namespace}/{name}/approve", h.withResponseWriter(responseWriter, h.approveAlertScale)).Methods("POST")
	api.HandleFunc("/alertscales/{namespace}/{name}/reject", h.withResponseWriter(responseWriter, h.rejectAlertScale)).Methods("POST")
}

// HandlerFuncWithWriter wrapper type
type AlertScaleHandlerFuncWithWriter func(ResponseWriter, http.ResponseWriter, *http.Request)

// withResponseWriter wraps handler functions with ResponseWriter
func (h *AlertScaleHandler) withResponseWriter(rw ResponseWriter, handler AlertScaleHandlerFuncWithWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(rw, w, r)
	}
}

// AlertScaleInfo represents AlertScale information for API responses
type AlertScaleInfo struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Reason       string `json:"reason"`
	Status       string `json:"status"`
	Duration     string `json:"duration"`
	Template     string `json:"template"`
	AutoApproval bool   `json:"autoApproval"`
	CreatedAt    string `json:"createdAt,omitempty"`
}

// ApprovalRequest represents an approval/rejection request
type ApprovalRequest struct {
	Approver string `json:"approver"`
	Reason   string `json:"reason"`
	Comment  string `json:"comment,omitempty"`
}

// listAlertScales handles GET /api/v1/alertscales
func (h *AlertScaleHandler) listAlertScales(responseWriter ResponseWriter, w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	log := logf.FromContext(ctx)

	var alertScaleList opsv1beta1.AlertScaleList
	if err := h.client.List(ctx, &alertScaleList); err != nil {
		log.Error(err, "Failed to list AlertScales")
		responseWriter.WriteError(w, http.StatusInternalServerError, "Failed to list AlertScales", err)
		return
	}

	items := make([]AlertScaleInfo, 0, len(alertScaleList.Items))
	for _, alertScale := range alertScaleList.Items {
		info := AlertScaleInfo{
			Name:         alertScale.Name,
			Namespace:    alertScale.Namespace,
			Reason:       alertScale.Spec.ScaleReason,
			Status:       alertScale.Status.ScaleStatus.Status,
			Duration:     alertScale.Spec.ScaleDuration,
			Template:     alertScale.Spec.ScaleNotifyMsgTemplate,
			AutoApproval: alertScale.Spec.ScaleAutoApproval,
			CreatedAt:    alertScale.CreationTimestamp.Format(time.RFC3339),
		}
		items = append(items, info)
	}

	responseData := map[string]interface{}{
		"items": items,
		"count": len(items),
	}

	responseWriter.WriteSuccess(w, "AlertScales retrieved successfully", responseData)
}

// getAlertScale handles GET /api/v1/alertscales/{namespace}/{name}
func (h *AlertScaleHandler) getAlertScale(responseWriter ResponseWriter, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	ctx := context.Background()
	log := logf.FromContext(ctx)

	var alertScale opsv1beta1.AlertScale
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	if err := h.client.Get(ctx, key, &alertScale); err != nil {
		log.Error(err, "Failed to get AlertScale", "namespace", namespace, "name", name)
		if client.IgnoreNotFound(err) == nil {
			responseWriter.WriteError(w, http.StatusNotFound, "AlertScale not found", err)
		} else {
			responseWriter.WriteError(w, http.StatusInternalServerError, "Failed to get AlertScale", err)
		}
		return
	}

	info := AlertScaleInfo{
		Name:         alertScale.Name,
		Namespace:    alertScale.Namespace,
		Reason:       alertScale.Spec.ScaleReason,
		Status:       alertScale.Status.ScaleStatus.Status,
		Duration:     alertScale.Spec.ScaleDuration,
		Template:     alertScale.Spec.ScaleNotifyMsgTemplate,
		AutoApproval: alertScale.Spec.ScaleAutoApproval,
		CreatedAt:    alertScale.CreationTimestamp.Format(time.RFC3339),
	}

	responseWriter.WriteSuccess(w, "AlertScale retrieved successfully", info)
}

// approveAlertScale handles POST /api/v1/alertscales/{namespace}/{name}/approve
func (h *AlertScaleHandler) approveAlertScale(responseWriter ResponseWriter, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	// Parse request body
	var req CommonApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWriter.WriteError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Create common approval processor
	processor := NewCommonApprovalProcessor(h.client)

	// Process approval using specialized method
	resourceKey := client.ObjectKey{Namespace: namespace, Name: name}
	if err := processor.ProcessAlertScaleApproval(r.Context(), resourceKey, "approve", req); err != nil {
		switch {
		case err.Error() == ErrResourceNotFound:
			responseWriter.WriteError(w, http.StatusNotFound, "AlertScale not found", err)
		case err.Error() == ErrResourceNotInApprovalState:
			responseWriter.WriteError(w, http.StatusBadRequest, "AlertScale is not in approvaling state", err)
		case err.Error() == ErrApproverRequired || err.Error() == ErrReasonRequired:
			responseWriter.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		default:
			responseWriter.WriteError(w, http.StatusInternalServerError, "Failed to approve AlertScale", err)
		}
		return
	}

	responseData := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"status":    "Approved",
		"approver":  req.Approver,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	responseWriter.WriteSuccess(w, "AlertScale approved successfully", responseData)
}

// rejectAlertScale handles POST /api/v1/alertscales/{namespace}/{name}/reject
func (h *AlertScaleHandler) rejectAlertScale(responseWriter ResponseWriter, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	// Parse request body
	var req CommonApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWriter.WriteError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Create common approval processor
	processor := NewCommonApprovalProcessor(h.client)

	// Process rejection using specialized method
	resourceKey := client.ObjectKey{Namespace: namespace, Name: name}
	if err := processor.ProcessAlertScaleApproval(r.Context(), resourceKey, "reject", req); err != nil {
		switch {
		case err.Error() == ErrResourceNotFound:
			responseWriter.WriteError(w, http.StatusNotFound, "AlertScale not found", err)
		case err.Error() == ErrResourceNotInApprovalState:
			responseWriter.WriteError(w, http.StatusBadRequest, "AlertScale is not in approvaling state", err)
		case err.Error() == ErrApproverRequired || err.Error() == ErrReasonRequired:
			responseWriter.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		default:
			responseWriter.WriteError(w, http.StatusInternalServerError, "Failed to reject AlertScale", err)
		}
		return
	}

	responseData := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"status":    "Rejected",
		"rejector":  req.Approver,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	responseWriter.WriteSuccess(w, "AlertScale rejected successfully", responseData)
}
