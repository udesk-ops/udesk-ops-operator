package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// ResponseWriter interface for handling HTTP responses
type ResponseWriter interface {
	WriteSuccess(w http.ResponseWriter, message string, data interface{})
	WriteError(w http.ResponseWriter, statusCode int, message string, err error)
	WriteResponse(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}, errorMsg string)
}

// DefaultResponseWriter implements ResponseWriter interface
type DefaultResponseWriter struct{}

// NewDefaultResponseWriter creates a new default response writer
func NewDefaultResponseWriter() ResponseWriter {
	return &DefaultResponseWriter{}
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// WriteSuccess writes a successful response
func (rw *DefaultResponseWriter) WriteSuccess(w http.ResponseWriter, message string, data interface{}) {
	rw.WriteResponse(w, http.StatusOK, true, message, data, "")
}

// WriteError writes an error response
func (rw *DefaultResponseWriter) WriteError(w http.ResponseWriter, statusCode int, message string, err error) {
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}
	rw.WriteResponse(w, statusCode, false, message, nil, errorMsg)
}

// WriteResponse writes a standard API response
func (rw *DefaultResponseWriter) WriteResponse(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}, errorMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		Error:     errorMsg,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback if JSON encoding fails
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
