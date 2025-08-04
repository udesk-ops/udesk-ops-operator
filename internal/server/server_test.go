package server

import (
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestNewAPIServer tests API server creation
func TestNewAPIServer(t *testing.T) {
	// Create a fake Kubernetes client
	k8sClient := fake.NewClientBuilder().Build()

	// Create API server
	server := NewAPIServer(k8sClient, ":8088")

	// Verify server is created
	if server == nil {
		t.Fatal("Expected API server to be created, got nil")
	}

	if server.addr != ":8088" {
		t.Errorf("Expected address :8088, got %s", server.addr)
	}

	if server.client != k8sClient {
		t.Error("Expected client to be set correctly")
	}
}

// TestAPIServerAddr tests different address formats
func TestAPIServerAddr(t *testing.T) {
	k8sClient := fake.NewClientBuilder().Build()

	testCases := []struct {
		name string
		addr string
	}{
		{"localhost", "localhost:8088"},
		{"ip address", "127.0.0.1:8088"},
		{"port only", ":9000"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := NewAPIServer(k8sClient, tc.addr)
			if server.addr != tc.addr {
				t.Errorf("Expected address %s, got %s", tc.addr, server.addr)
			}
		})
	}
}
