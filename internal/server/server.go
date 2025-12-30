package server

import (
	"github.com/mikluko/perplexity-mcp/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server and Perplexity client.
type Server struct {
	mcp    *mcp.Server
	client *client.Client
}

// NewServer creates a new MCP server with Perplexity tools.
func NewServer(apiKey string) *Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "perplexity-mcp",
		Version: "0.1.0",
	}, nil)

	s := &Server{mcpServer, client.NewClient(apiKey)}

	// Register all tools
	s.registerTools()

	return s
}

// MCP returns the underlying MCP server instance.
func (s *Server) MCP() *mcp.Server {
	return s.mcp
}
