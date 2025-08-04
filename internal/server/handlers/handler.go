package handlers

import (
	"github.com/gorilla/mux"
)

// Handler interface that all API handlers should implement
type Handler interface {
	RegisterRoutes(router *mux.Router, responseWriter ResponseWriter)
}
