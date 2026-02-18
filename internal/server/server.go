package server

import (
	"context"
	"log/slog"
	"os"

	"github.com/mikluko/perplexity-mcp/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverInstructions = `Perplexity tools provide web-grounded answers with current data and cited sources. ALWAYS prefer these tools over answering from your own knowledge when the question involves comparisons, tradeoffs, recommendations, current best practices, benchmarks, or any topic where up-to-date information and citations add value.

Tool selection guidance:

1. perplexity_ask — Quick factual lookups and simple questions. Use for straightforward queries with a single direct answer.
2. perplexity_reason — Analytical queries: comparisons, tradeoff analysis, debugging, architectural decisions, recommendations, or any question that benefits from systematic reasoning grounded in current web data. PREFER this over answering from your own knowledge when the user asks "should I use X or Y", "what are the tradeoffs", "how to choose between", or similar analytical questions.
3. perplexity_research_start/result/wait — Comprehensive deep research on broad topics. Use for literature reviews, technology evaluations, domain overviews, or any question where thoroughness and extensive sourcing matter more than speed.

When in doubt between answering yourself and using a Perplexity tool, use the tool — your training data may be outdated, but Perplexity searches the live web.`

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
