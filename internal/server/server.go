package server

import (
	"context"
	"log/slog"
	"os"

	"github.com/mikluko/perplexity-mcp/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverInstructions = `Perplexity tools provide web-grounded answers with current data and cited sources. Use them when the answer depends on up-to-date information that your training data may lack — NOT as a replacement for your own reasoning and knowledge.

Tool selection guidance:

1. perplexity_ask — Quick factual lookups and simple questions. Use for straightforward queries with a single direct answer.
2. perplexity_reason — Analytical queries that require current web data: comparisons using recent benchmarks, tradeoff analysis with up-to-date ecosystem information, debugging with current docs, or recommendations where the landscape has likely changed since your training cutoff.
3. perplexity_research_start/result/wait — Comprehensive deep research on broad topics. Use for literature reviews, technology evaluations, domain overviews, or any question where thoroughness and extensive sourcing matter more than speed.

Use these tools when the user needs current facts, recent developments, live documentation, or cited sources. Do NOT use them for questions you can confidently answer from your own knowledge — you are the primary assistant, not a proxy for Perplexity.`

// Server wraps the MCP server and Perplexity client.
type Server struct {
	mcp    *mcp.Server
	client *client.Client
}

// NewServer creates a new MCP server with Perplexity tools.
func NewServer(version, apiKey string) *Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "perplexity-mcp",
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: serverInstructions,
	})

	s := &Server{mcpServer, client.NewClient(apiKey)}

	// Register all tools
	s.registerTools()

	// Register all prompts
	s.registerPrompts()

	return s
}

// MCP returns the underlying MCP server instance.
func (s *Server) MCP() *mcp.Server {
	return s.mcp
}

// withLogging wraps a tool handler with automatic logging.
func withLogging[In, Out any](s *Server, toolName string, handler mcp.ToolHandlerFor[In, Out]) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		logger := s.getLogger()
		logger.Info(toolName+" called", "input", in)

		result, output, err := handler(ctx, req, in)

		if err != nil || (result != nil && result.IsError) {
			logger.Error(toolName+" failed", "error", err)
		} else {
			logger.Debug(toolName + " succeeded")
		}

		return result, output, err
	}
}

// getLogger returns an slog.Logger that sends logs to the MCP client.
// Falls back to stderr if no session is available.
func (s *Server) getLogger() *slog.Logger {
	for session := range s.mcp.Sessions() {
		return slog.New(mcp.NewLoggingHandler(session, &mcp.LoggingHandlerOptions{
			LoggerName: "perplexity-mcp",
		}))
	}
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}
