package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	scaletypes "udesk.cn/ops/internal/types"
)

// APIServer provides REST API endpoints for external access
type APIServer struct {
	client client.Client
	server *http.Server
	router *mux.Router
	addr   string
}

// NewAPIServer creates a new API server instance
func NewAPIServer(k8sClient client.Client, addr string) *APIServer {
	s := &APIServer{
		client: k8sClient,
		addr:   addr,
		router: mux.NewRouter(),
	}

	s.setupRoutes()
	s.setupMiddleware()

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start implements manager.Runnable interface
func (s *APIServer) Start(ctx context.Context) error {
	log := logf.FromContext(ctx)
	log.Info("Starting API server", "address", s.addr)

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err, "API server failed to start")
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	log.Info("Shutting down API server")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Error(err, "Failed to gracefully shutdown API server")
		return err
	}

	log.Info("API server stopped")
	return nil
}

// setupRoutes configures all API routes
func (s *APIServer) setupRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// AlertScale endpoints
	api.HandleFunc("/alertscales", s.listAlertScales).Methods("GET")
	api.HandleFunc("/alertscales/{namespace}/{name}", s.getAlertScale).Methods("GET")
	api.HandleFunc("/alertscales/{namespace}/{name}/approve", s.approveAlertScale).Methods("POST")
	api.HandleFunc("/alertscales/{namespace}/{name}/reject", s.rejectAlertScale).Methods("POST")

	// Health endpoint
	api.HandleFunc("/health", s.healthCheck).Methods("GET")
}

// setupMiddleware configures middleware for the API server
func (s *APIServer) setupMiddleware() {
	// CORS middleware
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Logging middleware
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logf.Log.WithName("api-server").Info("API request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start),
			)
		})
	})
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// Helper methods for writing responses
func (s *APIServer) writeResponse(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}, errorMsg string) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		Error:     errorMsg,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logf.Log.WithName("api-server").Error(err, "Failed to encode JSON response")
	}
}

func (s *APIServer) writeSuccess(w http.ResponseWriter, message string, data interface{}) {
	s.writeResponse(w, http.StatusOK, true, message, data, "")
}

func (s *APIServer) writeError(w http.ResponseWriter, statusCode int, message string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	s.writeResponse(w, statusCode, false, message, nil, errorMsg)
}

// listAlertScales handles GET /api/v1/alertscales
func (s *APIServer) listAlertScales(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logf.FromContext(ctx)

	// Query AlertScale list
	alertScaleList := &opsv1beta1.AlertScaleList{}
	if err := s.client.List(ctx, alertScaleList); err != nil {
		log.Error(err, "Failed to list AlertScales")
		s.writeError(w, http.StatusInternalServerError, "Failed to list AlertScales", err)
		return
	}

	// Convert to response format
	responses := make([]map[string]interface{}, 0, len(alertScaleList.Items))
	for _, item := range alertScaleList.Items {
		responses = append(responses, map[string]interface{}{
			"name":      item.Name,
			"namespace": item.Namespace,
			"reason":    item.Spec.ScaleReason,
			"status":    item.Status.ScaleStatus.Status,
			"duration":  item.Spec.ScaleDuration,
			"template":  item.Spec.ScaleNotifyMsgTemplate,
		})
	}

	s.writeSuccess(w, "AlertScales retrieved successfully", map[string]interface{}{
		"items": responses,
		"count": len(responses),
	})
}

// getAlertScale handles GET /api/v1/alertscales/{namespace}/{name}
func (s *APIServer) getAlertScale(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logf.FromContext(ctx)

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	if namespace == "" || name == "" {
		s.writeError(w, http.StatusBadRequest, "Missing namespace or name parameter", nil)
		return
	}

	// Get AlertScale
	alertScale := &opsv1beta1.AlertScale{}
	key := types.NamespacedName{Name: name, Namespace: namespace}

	if err := s.client.Get(ctx, key, alertScale); err != nil {
		if client.IgnoreNotFound(err) == nil {
			s.writeError(w, http.StatusNotFound, "AlertScale not found", nil)
			return
		}
		log.Error(err, "Failed to get AlertScale", "name", name, "namespace", namespace)
		s.writeError(w, http.StatusInternalServerError, "Failed to get AlertScale", err)
		return
	}

	response := map[string]interface{}{
		"name":         alertScale.Name,
		"namespace":    alertScale.Namespace,
		"reason":       alertScale.Spec.ScaleReason,
		"status":       alertScale.Status.ScaleStatus.Status,
		"duration":     alertScale.Spec.ScaleDuration,
		"template":     alertScale.Spec.ScaleNotifyMsgTemplate,
		"autoApproval": alertScale.Spec.ScaleAutoApproval,
		"createdAt":    alertScale.CreationTimestamp.Format(time.RFC3339),
	}

	s.writeSuccess(w, "AlertScale retrieved successfully", response)
}

// ApprovalRequest represents approval/rejection request
type ApprovalRequest struct {
	Reason   string `json:"reason,omitempty"`
	Approver string `json:"approver,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

// approveAlertScale handles POST /api/v1/alertscales/{namespace}/{name}/approve
func (s *APIServer) approveAlertScale(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logf.FromContext(ctx)

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	if namespace == "" || name == "" {
		s.writeError(w, http.StatusBadRequest, "Missing namespace or name parameter", nil)
		return
	}

	// Parse approval request
	var req ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body for simple approval
		req = ApprovalRequest{}
	}

	// Get AlertScale
	alertScale := &opsv1beta1.AlertScale{}
	key := types.NamespacedName{Name: name, Namespace: namespace}

	if err := s.client.Get(ctx, key, alertScale); err != nil {
		if client.IgnoreNotFound(err) == nil {
			s.writeError(w, http.StatusNotFound, "AlertScale not found", nil)
			return
		}
		log.Error(err, "Failed to get AlertScale for approval", "name", name, "namespace", namespace)
		s.writeError(w, http.StatusInternalServerError, "Failed to get AlertScale", err)
		return
	}

	// Check current status
	currentStatus := alertScale.Status.ScaleStatus.Status
	if currentStatus != scaletypes.ScaleStatusApprovaling && currentStatus != scaletypes.ScaleStatusPending {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("AlertScale is not in approvaling or pending state, current status: %s", currentStatus), nil)
		return
	}

	// Update status to approved
	alertScale.Status.ScaleStatus.Status = scaletypes.ScaleStatusApproved
	alertScale.Status.ScaleStatus.ScaleEndTime = metav1.NewTime(time.Now())

	// Add approval information to annotations
	if alertScale.Annotations == nil {
		alertScale.Annotations = make(map[string]string)
	}
	alertScale.Annotations["ops.udesk.cn/approved-by"] = req.Approver
	alertScale.Annotations["ops.udesk.cn/approved-at"] = time.Now().UTC().Format(time.RFC3339)
	alertScale.Annotations["ops.udesk.cn/approval-reason"] = req.Reason
	alertScale.Annotations["ops.udesk.cn/approval-comment"] = req.Comment

	if err := s.client.Status().Update(ctx, alertScale); err != nil {
		log.Error(err, "Failed to approve AlertScale", "name", name, "namespace", namespace)
		s.writeError(w, http.StatusInternalServerError, "Failed to approve AlertScale", err)
		return
	}

	// Also update the main object to persist annotations
	if err := s.client.Update(ctx, alertScale); err != nil {
		log.Error(err, "Failed to update AlertScale annotations", "name", name, "namespace", namespace)
		// Don't fail the approval if annotation update fails
	}

	log.Info("AlertScale approved successfully", "name", name, "namespace", namespace, "approver", req.Approver)
	s.writeSuccess(w, "AlertScale approved successfully", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"status":    scaletypes.ScaleStatusApproved,
		"approver":  req.Approver,
	})
}

// rejectAlertScale handles POST /api/v1/alertscales/{namespace}/{name}/reject
func (s *APIServer) rejectAlertScale(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logf.FromContext(ctx)

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	if namespace == "" || name == "" {
		s.writeError(w, http.StatusBadRequest, "Missing namespace or name parameter", nil)
		return
	}

	// Parse rejection request
	var req ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body for simple rejection
		req = ApprovalRequest{}
	}

	// Get AlertScale
	alertScale := &opsv1beta1.AlertScale{}
	key := types.NamespacedName{Name: name, Namespace: namespace}

	if err := s.client.Get(ctx, key, alertScale); err != nil {
		if client.IgnoreNotFound(err) == nil {
			s.writeError(w, http.StatusNotFound, "AlertScale not found", nil)
			return
		}
		log.Error(err, "Failed to get AlertScale for rejection", "name", name, "namespace", namespace)
		s.writeError(w, http.StatusInternalServerError, "Failed to get AlertScale", err)
		return
	}

	// Check current status
	currentStatus := alertScale.Status.ScaleStatus.Status
	if currentStatus != scaletypes.ScaleStatusApprovaling && currentStatus != scaletypes.ScaleStatusPending {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("AlertScale is not in approvaling or pending state, current status: %s", currentStatus), nil)
		return
	}

	// Update status to rejected
	alertScale.Status.ScaleStatus.Status = scaletypes.ScaleStatusRejected
	alertScale.Status.ScaleStatus.ScaleEndTime = metav1.NewTime(time.Now())

	// Add rejection information to annotations
	if alertScale.Annotations == nil {
		alertScale.Annotations = make(map[string]string)
	}
	alertScale.Annotations["ops.udesk.cn/rejected-by"] = req.Approver
	alertScale.Annotations["ops.udesk.cn/rejected-at"] = time.Now().UTC().Format(time.RFC3339)
	alertScale.Annotations["ops.udesk.cn/rejection-reason"] = req.Reason
	alertScale.Annotations["ops.udesk.cn/rejection-comment"] = req.Comment

	if err := s.client.Status().Update(ctx, alertScale); err != nil {
		log.Error(err, "Failed to reject AlertScale", "name", name, "namespace", namespace)
		s.writeError(w, http.StatusInternalServerError, "Failed to reject AlertScale", err)
		return
	}

	// Also update the main object to persist annotations
	if err := s.client.Update(ctx, alertScale); err != nil {
		log.Error(err, "Failed to update AlertScale annotations", "name", name, "namespace", namespace)
		// Don't fail the rejection if annotation update fails
	}

	log.Info("AlertScale rejected successfully", "name", name, "namespace", namespace, "rejector", req.Approver)
	s.writeSuccess(w, "AlertScale rejected successfully", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"status":    scaletypes.ScaleStatusRejected,
		"rejector":  req.Approver,
	})
}

// healthCheck handles GET /api/v1/health
func (s *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.writeSuccess(w, "API server is healthy", map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "v1.0.0",
		"server":    "udesk-ops-operator-api-server",
	})
}
