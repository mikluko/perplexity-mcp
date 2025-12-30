package server

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	apiKey := "test-api-key"
	srv := NewServer(apiKey)

	if srv == nil {
		t.Fatal("NewServer() returned nil")
	}

	if srv.mcp == nil {
		t.Error("NewServer() mcp field is nil")
	}

	if srv.client == nil {
		t.Error("NewServer() client field is nil")
	}

	if srv.MCP() == nil {
		t.Error("NewServer() MCP() returns nil")
	}
}

func TestServerMCP(t *testing.T) {
	srv := NewServer("test-key")

	mcpServer := srv.MCP()
	if mcpServer == nil {
		t.Error("MCP() returned nil")
	}

	// Verify it returns the same instance
	if srv.MCP() != mcpServer {
		t.Error("MCP() does not return consistent instance")
	}
}
