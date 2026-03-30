package main

import (
	"os"
	"testing"
)

func TestDefaultPort(t *testing.T) {
	// Ensure GRPC_PORT is unset for this test.
	t.Setenv("GRPC_PORT", "")
	os.Unsetenv("GRPC_PORT")

	port := os.Getenv("GRPC_PORT")
	if port != "" {
		t.Fatalf("expected empty GRPC_PORT, got %q", port)
	}
}

func TestCustomPort(t *testing.T) {
	t.Setenv("GRPC_PORT", "8080")

	port := os.Getenv("GRPC_PORT")
	if port != "8080" {
		t.Fatalf("expected GRPC_PORT=8080, got %q", port)
	}
}
