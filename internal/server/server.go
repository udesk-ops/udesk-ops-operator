package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	handlers "udesk.cn/ops/internal/server/handlers"
)

// APIServer represents the REST API server
type APIServer struct {
	client client.Client
	addr   string
	server *http.Server
	router *mux.Router
}

// NewAPIServer creates a new API server instance
func NewAPIServer(k8sClient client.Client, addr string) *APIServer {
	router := mux.NewRouter()

	return &APIServer{
		client: k8sClient,
		addr:   addr,
		router: router,
		server: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start starts the API server
func (s *APIServer) Start(ctx context.Context) error {
	logger := log.FromContext(ctx)

	// Set up routes
	s.setupRoutes()

	logger.Info("Starting API server", "address", s.addr)

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "API server failed to start")
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	logger.Info("Shutting down API server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(shutdownCtx)
}

// setupRoutes configures all API routes
func (s *APIServer) setupRoutes() {
	logger := log.Log.WithName("api-server")

	// Create response writer
	responseWriter := handlers.NewDefaultResponseWriter()

	// Add CORS middleware
	s.router.Use(corsMiddleware)

	// Add logging middleware
	s.router.Use(loggingMiddleware(logger))

	// Register all handlers automatically
	s.registerHandlers(responseWriter)

	logger.Info("API routes configured successfully")
}

// registerHandlers registers all handlers using auto-discovery
func (s *APIServer) registerHandlers(responseWriter handlers.ResponseWriter) {
	logger := log.Log.WithName("api-server")

	// Get all registered handlers
	handlerList := handlers.GetAllHandlers(s.client)

	logger.Info("Auto-registered handlers", "count", len(handlerList))

	// Register routes for each handler
	for _, handler := range handlerList {
		handler.RegisterRoutes(s.router, responseWriter)
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(logger logr.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			duration := time.Since(start)
			logger.Info("API request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", fmt.Sprintf("%.1fms", float64(duration.Nanoseconds())/1e6))
		})
	}
}
