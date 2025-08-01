package server

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

func TestNewAPIServer(t *testing.T) {
	// Create a fake client
	scheme := runtime.NewScheme()
	if err := opsv1beta1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add scheme: %v", err)
	}

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Test server creation
	server := NewAPIServer(client, ":8088")

	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.addr != ":8088" {
		t.Errorf("Expected addr to be ':8088', got %s", server.addr)
	}

	if server.client == nil {
		t.Error("Expected client to be set")
	}

	if server.router == nil {
		t.Error("Expected router to be set")
	}

	if server.server == nil {
		t.Error("Expected HTTP server to be set")
	}
}
