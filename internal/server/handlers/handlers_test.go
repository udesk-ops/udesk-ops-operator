package handlers

import (
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewAlertScaleHandler(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	handler := NewAlertScaleHandler(client)

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}

	if handler.client != client {
		t.Error("Expected handler to store client")
	}
}

func TestNewHealthHandler(t *testing.T) {
	handler := NewHealthHandler()

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
}

func TestNewDefaultResponseWriter(t *testing.T) {
	rw := NewDefaultResponseWriter()

	if rw == nil {
		t.Fatal("Expected non-nil response writer")
	}
}
