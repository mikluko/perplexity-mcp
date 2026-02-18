package server

import (
	"context"
	"log/slog"
	"os"

	"github.com/mikluko/perplexity-mcp/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverInstructions = `This server provides three tiers of Perplexity AI tools. Choose the right tier based on query complexity:

1. perplexity_ask — Quick factual lookups and simple questions. Use when you need a fast, direct answer.
2. perplexity_reason — Analytical and multi-step problems: comparisons, debugging, tradeoff analysis, mathematical reasoning. Use when the question requires thinking through a problem, not just retrieving facts.
3. perplexity_research_start/result/wait — Deep, comprehensive research on broad topics. Use for literature reviews, technology evaluations, domain deep-dives, or any question where thoroughness matters more than speed. Returns significantly more detailed results but takes longer.

Default to perplexity_ask only for simple lookups. When in doubt between ask and reason, prefer reason. When the user needs comprehensive coverage of a topic, use research.`

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
