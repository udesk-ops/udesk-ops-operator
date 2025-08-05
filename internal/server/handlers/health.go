package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// init registers the Health handler automatically
func init() {
	RegisterHandler("health", func(k8sClient client.Client) Handler {
		return NewHealthHandler()
	})
}

// HealthHandler handles health check endpoints
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// RegisterRoutes registers health routes to the router
func (h *HealthHandler) RegisterRoutes(router *mux.Router, responseWriter ResponseWriter) {
	api := GetAPIRouter(router)
	api.HandleFunc("/health", h.withResponseWriter(responseWriter, h.healthCheck)).Methods("GET")
}

// HandlerFuncWithWriter with ResponseWriter support
type HealthHandlerFuncWithWriter func(ResponseWriter, http.ResponseWriter, *http.Request)

// withResponseWriter wraps handler functions with ResponseWriter
func (h *HealthHandler) withResponseWriter(rw ResponseWriter, handler HealthHandlerFuncWithWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(rw, w, r)
	}
}

// healthCheck handles GET /api/v1/health
func (h *HealthHandler) healthCheck(responseWriter ResponseWriter, w http.ResponseWriter, r *http.Request) {
	healthData := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "v1.0.0",
		"server":    "udesk-ops-operator-api-server",
	}

	responseWriter.WriteSuccess(w, "API server is healthy", healthData)
}
