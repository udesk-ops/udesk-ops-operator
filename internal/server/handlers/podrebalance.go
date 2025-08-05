package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/constants"
	"udesk.cn/ops/internal/types"
)

const (
	// DefaultNamespace is the default namespace for resources
	DefaultNamespace = "default"
)

// init registers the PodRebalance handler automatically
func init() {
	RegisterHandler("podrebalance", func(k8sClient client.Client) Handler {
		return NewPodRebalanceHandler(k8sClient)
	})
}

// PodRebalanceHandler handles PodRebalance API requests
type PodRebalanceHandler struct {
	client client.Client
}

// NewPodRebalanceHandler creates a new PodRebalance handler
func NewPodRebalanceHandler(k8sClient client.Client) *PodRebalanceHandler {
	return &PodRebalanceHandler{
		client: k8sClient,
	}
}

// RegisterRoutes registers PodRebalance routes to the router
func (h *PodRebalanceHandler) RegisterRoutes(router *mux.Router, responseWriter ResponseWriter) {
	api := GetAPIRouter(router)

	// PodRebalance resource routes
	api.HandleFunc("/podrebalances", h.handleList).Methods("GET")
	api.HandleFunc("/podrebalances", h.handleCreate).Methods("POST")
	api.HandleFunc("/podrebalances/{name}", h.handleGet).Methods("GET")
	api.HandleFunc("/podrebalances/{name}", h.handleUpdate).Methods("PUT")
	api.HandleFunc("/podrebalances/{name}", h.handleDelete).Methods("DELETE")

	// PodRebalance approval routes (复用通用审批接口)
	api.HandleFunc("/podrebalances/{name}/approve", h.handleApprove).Methods("POST")
	api.HandleFunc("/podrebalances/{name}/reject", h.handleReject).Methods("POST")
}

// PodRebalanceInfo represents PodRebalance information for API response
type PodRebalanceInfo struct {
	Name         string                          `json:"name"`
	Namespace    string                          `json:"namespace"`
	Status       string                          `json:"status"`
	AutoApproval bool                            `json:"autoApproval"`
	Strategy     opsv1beta1.PodRebalanceStrategy `json:"strategy"`
	DryRun       bool                            `json:"dryRun"`
	CreatedAt    time.Time                       `json:"createdAt"`
	BeginTime    *time.Time                      `json:"beginTime,omitempty"`
	EndTime      *time.Time                      `json:"endTime,omitempty"`
	Message      string                          `json:"message,omitempty"`
}

// handleList lists PodRebalance resources
func (h *PodRebalanceHandler) handleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logf.FromContext(ctx).WithName("podrebalance-list")

	var podRebalanceList opsv1beta1.PodRebalanceList
	if err := h.client.List(ctx, &podRebalanceList); err != nil {
		log.Error(err, "Failed to list PodRebalances")
		http.Error(w, "Failed to list PodRebalances", http.StatusInternalServerError)
		return
	}

	podRebalances := make([]PodRebalanceInfo, 0, len(podRebalanceList.Items))
	for _, pr := range podRebalanceList.Items {
		info := PodRebalanceInfo{
			Name:         pr.Name,
			Namespace:    pr.Namespace,
			Status:       pr.Status.Status,
			AutoApproval: pr.Spec.AutoApproval,
			Strategy:     pr.Spec.Strategy,
			DryRun:       pr.Spec.DryRun,
			CreatedAt:    pr.CreationTimestamp.Time,
			Message:      pr.Status.Message,
		}

		if !pr.Status.RebalanceBeginTime.IsZero() {
			info.BeginTime = &pr.Status.RebalanceBeginTime.Time
		}
		if !pr.Status.RebalanceEndTime.IsZero() {
			info.EndTime = &pr.Status.RebalanceEndTime.Time
		}

		podRebalances = append(podRebalances, info)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"podRebalances": podRebalances,
		"total":         len(podRebalances),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleGet gets a specific PodRebalance resource
func (h *PodRebalanceHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logf.FromContext(ctx).WithName("podrebalance-get")
	vars := mux.Vars(r)
	name := vars["name"]
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}

	var podRebalance opsv1beta1.PodRebalance
	if err := h.client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &podRebalance); err != nil {
		if client.IgnoreNotFound(err) == nil {
			http.Error(w, "PodRebalance not found", http.StatusNotFound)
			return
		}
		log.Error(err, "Failed to get PodRebalance", "namespace", namespace, "name", name)
		http.Error(w, "Failed to get PodRebalance", http.StatusInternalServerError)
		return
	}

	info := PodRebalanceInfo{
		Name:         podRebalance.Name,
		Namespace:    podRebalance.Namespace,
		Status:       podRebalance.Status.Status,
		AutoApproval: podRebalance.Spec.AutoApproval,
		Strategy:     podRebalance.Spec.Strategy,
		DryRun:       podRebalance.Spec.DryRun,
		CreatedAt:    podRebalance.CreationTimestamp.Time,
		Message:      podRebalance.Status.Message,
	}

	if !podRebalance.Status.RebalanceBeginTime.IsZero() {
		info.BeginTime = &podRebalance.Status.RebalanceBeginTime.Time
	}
	if !podRebalance.Status.RebalanceEndTime.IsZero() {
		info.EndTime = &podRebalance.Status.RebalanceEndTime.Time
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleCreate creates a new PodRebalance resource
func (h *PodRebalanceHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logf.FromContext(ctx).WithName("podrebalance-create")

	var spec opsv1beta1.PodRebalanceSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if spec.Namespace == "" || spec.Strategy.Type == "" {
		http.Error(w, "namespace and strategy.type are required", http.StatusBadRequest)
		return
	}

	podRebalance := &opsv1beta1.PodRebalance{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "podrebalance-",
			Namespace:    spec.Namespace,
		},
		Spec: spec,
		Status: opsv1beta1.PodRebalanceStatus{
			Status: types.RebalanceStatusPending,
		},
	}

	if err := h.client.Create(ctx, podRebalance); err != nil {
		log.Error(err, "Failed to create PodRebalance")
		http.Error(w, "Failed to create PodRebalance", http.StatusInternalServerError)
		return
	}

	log.Info("Created PodRebalance", "namespace", podRebalance.Namespace, "name", podRebalance.Name)

	info := PodRebalanceInfo{
		Name:         podRebalance.Name,
		Namespace:    podRebalance.Namespace,
		Status:       podRebalance.Status.Status,
		AutoApproval: podRebalance.Spec.AutoApproval,
		Strategy:     podRebalance.Spec.Strategy,
		DryRun:       podRebalance.Spec.DryRun,
		CreatedAt:    podRebalance.CreationTimestamp.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleUpdate updates a PodRebalance resource
func (h *PodRebalanceHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Update operation not supported for PodRebalance", http.StatusMethodNotAllowed)
}

// handleDelete deletes a PodRebalance resource
func (h *PodRebalanceHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logf.FromContext(ctx).WithName("podrebalance-delete")
	vars := mux.Vars(r)
	name := vars["name"]
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}

	var podRebalance opsv1beta1.PodRebalance
	if err := h.client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &podRebalance); err != nil {
		if client.IgnoreNotFound(err) == nil {
			http.Error(w, "PodRebalance not found", http.StatusNotFound)
			return
		}
		log.Error(err, "Failed to get PodRebalance for deletion", "namespace", namespace, "name", name)
		http.Error(w, "Failed to get PodRebalance", http.StatusInternalServerError)
		return
	}

	if err := h.client.Delete(ctx, &podRebalance); err != nil {
		log.Error(err, "Failed to delete PodRebalance", "namespace", namespace, "name", name)
		http.Error(w, "Failed to delete PodRebalance", http.StatusInternalServerError)
		return
	}

	log.Info("Deleted PodRebalance", "namespace", namespace, "name", name)
	w.WriteHeader(http.StatusNoContent)
}

// handleApprove approves a PodRebalance resource
func (h *PodRebalanceHandler) handleApprove(w http.ResponseWriter, r *http.Request) {
	h.handleApprovalAction(w, r, "approve")
}

// handleReject rejects a PodRebalance resource
func (h *PodRebalanceHandler) handleReject(w http.ResponseWriter, r *http.Request) {
	h.handleApprovalAction(w, r, "reject")
}

// PodRebalanceApprovalRequest represents an approval/rejection request for PodRebalance
type PodRebalanceApprovalRequest struct {
	Approver string `json:"approver"`
	Reason   string `json:"reason"`
	Comment  string `json:"comment,omitempty"`
}

// handleApprovalAction handles approval/rejection actions
func (h *PodRebalanceHandler) handleApprovalAction(w http.ResponseWriter, r *http.Request, action string) {
	ctx := r.Context()
	log := logf.FromContext(ctx).WithName("podrebalance-approval")
	vars := mux.Vars(r)
	name := vars["name"]
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}

	var req PodRebalanceApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Approver == "" || req.Reason == "" {
		http.Error(w, "approver and reason are required", http.StatusBadRequest)
		return
	}

	var podRebalance opsv1beta1.PodRebalance
	if err := h.client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &podRebalance); err != nil {
		if client.IgnoreNotFound(err) == nil {
			http.Error(w, "PodRebalance not found", http.StatusNotFound)
			return
		}
		log.Error(err, "Failed to get PodRebalance for approval", "namespace", namespace, "name", name)
		http.Error(w, "Failed to get PodRebalance", http.StatusInternalServerError)
		return
	}

	if podRebalance.Status.Status != types.RebalanceStatusApprovaling {
		http.Error(w, "PodRebalance is not in approvaling state", http.StatusBadRequest)
		return
	}

	timestamp := time.Now().Format(time.RFC3339)

	// 设置审批注解 - 控制器会检测并处理
	if podRebalance.Annotations == nil {
		podRebalance.Annotations = make(map[string]string)
	}
	podRebalance.Annotations[constants.ApprovalDecisionAnnotation] = action
	podRebalance.Annotations[constants.ApprovalTimestampAnnotation] = timestamp
	podRebalance.Annotations[constants.ApprovalOperatorAnnotation] = req.Approver
	podRebalance.Annotations[constants.ApprovalReasonAnnotation] = req.Reason
	if req.Comment != "" {
		podRebalance.Annotations[constants.ApprovalCommentAnnotation] = req.Comment
	}
	podRebalance.Annotations[constants.ApprovalProcessingAnnotation] = ApprovalProcessingPending

	if err := h.client.Update(ctx, &podRebalance); err != nil {
		log.Error(err, "Failed to update PodRebalance with approval decision", "action", action)
		http.Error(w, "Failed to process approval", http.StatusInternalServerError)
		return
	}

	log.Info("PodRebalance approval processed", "namespace", namespace, "name", name, "action", action, "approver", req.Approver)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"action":    action,
		"approver":  req.Approver,
		"timestamp": timestamp,
		"message":   "Approval processed successfully",
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
