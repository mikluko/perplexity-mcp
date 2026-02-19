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

1. perplexity_ask — Niche/specialized queries and fast fact verification. Excels where general web search drowns in noise: specific regulations, financial data, product specs, configuration details. Also use as a fast sanity check to verify that facts from your training data are still current.
2. perplexity_reason — Analytical queries that require both reasoning AND current web data: debugging with current docs, configuration troubleshooting, specific tradeoff analysis between named alternatives. Also a powerful fallback when your built-in web search returns inconclusive or contradictory results on broad queries. More expensive in time and tokens — do not use as a default.
3. perplexity_research_start/result/wait — Comprehensive deep research on broad topics. Use for literature reviews, technology evaluations, market analysis, understanding a domain in depth, or any question where thoroughness and extensive sourcing matter more than speed.

Query routing — perplexity_ask vs built-in web search:

perplexity_ask excels at narrow, focused queries ("How do I fix X?", "What config for Y?", specific bugs, niche regulations, product recommendations). It synthesizes scattered information into actionable answers and filters noise effectively.

For broad ecosystem exploration ("Compare all X solutions", "What's the landscape of Y?") and factual lookups on well-documented topics (release notes, official docs, multi-vendor ecosystems), start with your built-in web search — it returns more sources, surfaces primary/official documentation, and discovers entire categories that perplexity_ask tends to miss. If results are inconclusive or contradictory, escalate to perplexity_reason for deeper analytical synthesis.

Factual accuracy: Perplexity occasionally makes subtle factual errors in synthesized answers (e.g., claiming a feature doesn't exist when it does). When a Perplexity answer will directly influence implementation decisions, cross-check specific technical claims against primary documentation.

Use these tools when the user needs current facts, recent developments, live documentation, or cited sources. Do NOT use them for questions you can confidently answer from your own knowledge — you are the primary assistant, not a proxy for Perplexity.

However, be aware that your training data has a cutoff date. When you are about to state fast-moving facts — software versions, API signatures, pricing and plan tiers, regulatory or compliance rules, company leadership, current market status, and similar — use perplexity_ask to verify your knowledge is still current. This is a quick, cheap check that prevents confidently stating outdated information. This list is not exhaustive; use your judgement to recognize when a fact is likely to have changed since your training.`

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
