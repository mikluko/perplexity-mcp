package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mikluko/perplexity-mcp/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	modelAsk      = "sonar-pro"
	modelResearch = "sonar-deep-research"
	modelReason   = "sonar-reasoning-pro"
)

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}
}

// ask tool

type AskInput struct {
	Query string `json:"query" jsonschema:"Question to answer using web search"`
}

var askTool = &mcp.Tool{
	Name:        "perplexity_ask",
	Description: "Answer a question using Perplexity AI web search. Best for niche/specialized queries where general web search drowns in noise (specific regulations, financial data, product specs, configuration details). Also use as a fast sanity check to verify that facts you know from training (software versions, API status, pricing, regulations, company details) are still current. For broad factual lookups on well-documented topics, prefer your built-in web search instead. For multi-step analysis or debugging use perplexity_reason. For comprehensive research on broad topics use perplexity_research_start.",
}

func (s *Server) handleAsk(ctx context.Context, _ *mcp.CallToolRequest, in AskInput) (*mcp.CallToolResult, any, error) {
	result, err := s.client.Query(ctx, modelAsk,
		"You are a helpful search assistant. Provide accurate, concise answers with relevant information from the web.",
		in.Query)
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(formatResultWithCitations(result)), nil, nil
}

func formatResultWithCitations(result *client.QueryResult) string {
	if len(result.Citations) == 0 {
		return result.Content
	}

	var sb strings.Builder
	sb.WriteString(result.Content)
	sb.WriteString("\n\n---\n**Sources:**\n")

	for i, c := range result.Citations {
		if c.Title != "" {
			fmt.Fprintf(&sb, "%d. [%s](%s)", i+1, c.Title, c.URL)
		} else {
			fmt.Fprintf(&sb, "%d. %s", i+1, c.URL)
		}
		if c.Date != "" {
			fmt.Fprintf(&sb, " (%s)", c.Date)
		}
		sb.WriteString("\n")
		if c.Snippet != "" {
			fmt.Fprintf(&sb, "   > %s\n", c.Snippet)
		}
	}

	return sb.String()
}

// research_start tool (async)

type ResearchStartInput struct {
	Query string `json:"query" jsonschema:"Topic to research in depth"`
}

var researchStartTool = &mcp.Tool{
	Name:        "perplexity_research_start",
	Description: "Start deep research on a topic (async). Returns request_id to check results later with perplexity_research_result or perplexity_research_wait. Use this instead of perplexity_ask for broad or complex topics that benefit from comprehensive investigation: literature reviews, technology evaluations, market analysis, understanding a domain in depth, or any question where thoroughness matters more than speed. Takes longer but produces significantly more detailed and well-sourced results.",
}

func (s *Server) handleResearchStart(ctx context.Context, _ *mcp.CallToolRequest, in ResearchStartInput) (*mcp.CallToolResult, any, error) {
	requestID, err := s.client.StartResearch(ctx, modelResearch, in.Query)
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult("Research started. Request ID: " + requestID), nil, nil
}

// research_result tool (async)

type ResearchResultInput struct {
	RequestID string `json:"request_id" jsonschema:"Request ID from perplexity_research_start"`
}

var researchResultTool = &mcp.Tool{
	Name:        "perplexity_research_result",
	Description: "Get results of async deep research by request_id",
}

func (s *Server) handleResearchResult(ctx context.Context, _ *mcp.CallToolRequest, in ResearchResultInput) (*mcp.CallToolResult, any, error) {
	status, err := s.client.GetResearchResult(ctx, in.RequestID)
	if err != nil {
		return errorResult(err), nil, nil
	}

	switch status.Status() {
	case "completed":
		if len(status.Response.Choices) > 0 {
			result := &client.QueryResult{
				Content: status.Response.Choices[0].Message.Content,
			}
			if len(status.Response.SearchResults) > 0 {
				result.Citations = status.Response.SearchResults
			} else if len(status.Response.Citations) > 0 {
				for _, url := range status.Response.Citations {
					result.Citations = append(result.Citations, client.Citation{URL: url})
				}
			}
			return textResult(formatResultWithCitations(result)), nil, nil
		}
		return textResult("Research completed but no content returned"), nil, nil
	case "failed":
		return errorResult(fmt.Errorf("research failed: %s", status.Error)), nil, nil
	default:
		return textResult("Status: " + status.Status() + ". Research is still in progress. Check again later."), nil, nil
	}
}

// research_wait tool (blocking)

type ResearchWaitInput struct {
	RequestID string `json:"request_id" jsonschema:"Request ID from perplexity_research_start"`
	Timeout   int    `json:"timeout,omitempty" jsonschema:"Timeout in seconds (default 300, max 600)"`
}

var researchWaitTool = &mcp.Tool{
	Name:        "perplexity_research_wait",
	Description: "Wait for async deep research to complete (blocking). Returns result when done or times out.",
}

func (s *Server) handleResearchWait(ctx context.Context, _ *mcp.CallToolRequest, in ResearchWaitInput) (*mcp.CallToolResult, any, error) {
	timeout := in.Timeout
	if timeout <= 0 {
		timeout = 300 // 5 min default
	}
	if timeout > 600 {
		timeout = 600 // 10 min max
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	status, err := s.client.WaitForResearch(ctx, in.RequestID, 5*time.Second)
	if err != nil {
		return errorResult(err), nil, nil
	}

	switch status.Status() {
	case "completed":
		if len(status.Response.Choices) > 0 {
			result := &client.QueryResult{
				Content: status.Response.Choices[0].Message.Content,
			}
			if len(status.Response.SearchResults) > 0 {
				result.Citations = status.Response.SearchResults
			} else if len(status.Response.Citations) > 0 {
				for _, url := range status.Response.Citations {
					result.Citations = append(result.Citations, client.Citation{URL: url})
				}
			}
			return textResult(formatResultWithCitations(result)), nil, nil
		}
		return textResult("Research completed but no content returned"), nil, nil
	case "failed":
		return errorResult(fmt.Errorf("research failed: %s", status.Error)), nil, nil
	default:
		return textResult("Unexpected status: " + status.Status()), nil, nil
	}
}

// reason tool

type ReasonInput struct {
	Query string `json:"query" jsonschema:"Problem or question requiring step-by-step reasoning"`
}

var reasonTool = &mcp.Tool{
	Name:        "perplexity_reason",
	Description: "Solve problems using step-by-step reasoning grounded in current web data. Provides cited sources and up-to-date information that your training data may lack. Use when the query requires analytical reasoning AND current data: debugging with current docs, configuration troubleshooting, or specific tradeoff analysis between named alternatives. Also a powerful fallback when your built-in web search returns inconclusive or contradictory results. More expensive in time and tokens than perplexity_ask — do not use as a default. Do not use for analytical questions you can answer confidently from your own knowledge.",
}

func (s *Server) handleReason(ctx context.Context, _ *mcp.CallToolRequest, in ReasonInput) (*mcp.CallToolResult, any, error) {
	result, err := s.client.Query(ctx, modelReason,
		"You are a reasoning assistant. Think step by step, show your work, and provide well-reasoned conclusions.",
		in.Query)
	if err != nil {
		return errorResult(err), nil, nil
	}
	return textResult(formatResultWithCitations(result)), nil, nil
}

// registerTools registers all Perplexity tools with the MCP server.
func (s *Server) registerTools() {
	mcp.AddTool(s.mcp, askTool, withLogging(s, "perplexity_ask", s.handleAsk))
	mcp.AddTool(s.mcp, researchStartTool, withLogging(s, "perplexity_research_start", s.handleResearchStart))
	mcp.AddTool(s.mcp, researchResultTool, withLogging(s, "perplexity_research_result", s.handleResearchResult))
	mcp.AddTool(s.mcp, researchWaitTool, withLogging(s, "perplexity_research_wait", s.handleResearchWait))
	mcp.AddTool(s.mcp, reasonTool, withLogging(s, "perplexity_reason", s.handleReason))
}
