package handlers

import (
	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// APIPrefix defines the common API prefix for all endpoints
	APIPrefix = "/api/v1"

	// API version constants
	APIVersion  = "v1"
	APIBasePath = "/api"
)

// GetAPIRouter returns a subrouter configured with the standard API prefix
// This ensures all handlers use the same API prefix consistently
func GetAPIRouter(router *mux.Router) *mux.Router {
	return router.PathPrefix(APIPrefix).Subrouter()
}

// HandlerFactory is a function type that creates a handler instance
type HandlerFactory func(k8sClient client.Client) Handler

// HandlerRegistry holds all registered handler factories
var handlerRegistry = make(map[string]HandlerFactory)

// RegisterHandler registers a handler factory with a given name
func RegisterHandler(name string, factory HandlerFactory) {
	handlerRegistry[name] = factory
}

// GetAllHandlers returns all registered handlers initialized with the given client
func GetAllHandlers(k8sClient client.Client) []Handler {
	handlers := make([]Handler, 0, len(handlerRegistry))
	for _, factory := range handlerRegistry {
		handlers = append(handlers, factory(k8sClient))
	}
	return handlers
}

// GetHandlerNames returns all registered handler names
func GetHandlerNames() []string {
	names := make([]string, 0, len(handlerRegistry))
	for name := range handlerRegistry {
		names = append(names, name)
	}
	return names
}
