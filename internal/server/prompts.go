package server

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const systemPrompt = `Concise, factual answers. Cite sources inline.
State explicitly when information is uncertain or unavailable.`

// Helper to format date for defaults (30/365 days ago)
func daysAgo(days int) string {
	return time.Now().AddDate(0, 0, -days).Format("2006-01-02")
}

// Research prompts


var researchPrompt = &mcp.Prompt{
	Name:        "research",
	Description: "Comprehensive research with current state and key developments",
	Arguments: []*mcp.PromptArgument{
		{Name: "topic", Description: "Topic to research", Required: true},
		{Name: "timeframe", Description: "Time period (e.g., 'last 12 months')", Required: false},
	},
}

func (s *Server) handleResearch(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	topic := req.Params.Arguments["topic"]
	timeframe := req.Params.Arguments["timeframe"]
	if timeframe == "" {
		timeframe = "last 12 months"
	}

	userPrompt := fmt.Sprintf("%s: current state, key developments since %s, main players, open questions. If information is incomplete, indicate gaps.", topic, timeframe)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

var comparePrompt = &mcp.Prompt{
	Name:        "compare",
	Description: "Compare two items with features and tradeoffs",
	Arguments: []*mcp.PromptArgument{
		{Name: "item_a", Description: "First item", Required: true},
		{Name: "item_b", Description: "Second item", Required: true},
		{Name: "use_case", Description: "Specific use case", Required: false},
	},
}

func (s *Server) handleCompare(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	itemA := req.Params.Arguments["item_a"]
	itemB := req.Params.Arguments["item_b"]
	useCase := req.Params.Arguments["use_case"]

	useCaseText := ""
	if useCase != "" {
		useCaseText = " for " + useCase
	}

	userPrompt := fmt.Sprintf("%s vs %s%s: feature comparison, tradeoffs, current recommendations", itemA, itemB, useCaseText)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

var verifyPrompt = &mcp.Prompt{
	Name:        "verify",
	Description: "Verify a claim with supporting and contradicting evidence",
	Arguments: []*mcp.PromptArgument{
		{Name: "claim", Description: "Claim to verify", Required: true},
	},
}

func (s *Server) handleVerify(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	claim := req.Params.Arguments["claim"]

	userPrompt := fmt.Sprintf(`Verify claim: "%s" — find supporting AND contradicting evidence. If sources conflict, note discrepancies.`, claim)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

var statusPrompt = &mcp.Prompt{
	Name:        "status",
	Description: "Check current status and recent developments",
	Arguments: []*mcp.PromptArgument{
		{Name: "entity", Description: "Entity to check", Required: true},
		{Name: "date", Description: "Start date (YYYY-MM-DD)", Required: false},
	},
}

func (s *Server) handleStatus(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	entity := req.Params.Arguments["entity"]
	date := req.Params.Arguments["date"]
	if date == "" {
		date = daysAgo(30)
	}

	userPrompt := fmt.Sprintf("%s current status: latest developments, changes, announcements since %s", entity, date)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

// Technical prompts

var docsPrompt = &mcp.Prompt{
	Name:        "docs",
	Description: "Find API reference and examples",
	Arguments: []*mcp.PromptArgument{
		{Name: "library", Description: "Library name", Required: true},
		{Name: "version", Description: "Version", Required: false},
		{Name: "topic", Description: "Specific topic", Required: false},
	},
}

func (s *Server) handleDocs(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	library := req.Params.Arguments["library"]
	version, _ := req.Params.Arguments["version"]
	topic, _ := req.Params.Arguments["topic"]

	versionText := ""
	if version != "" {
		versionText = " " + version
	}
	topicText := ""
	if topic != "" {
		topicText = " " + topic
	}

	userPrompt := fmt.Sprintf("%s%s%s: API reference, examples, common patterns", library, versionText, topicText)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

var errorPrompt = &mcp.Prompt{
	Name:        "error",
	Description: "Find causes and solutions for an error",
	Arguments: []*mcp.PromptArgument{
		{Name: "error_message", Description: "Error message", Required: true},
		{Name: "technology", Description: "Technology/framework", Required: true},
	},
}

func (s *Server) handleError(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	errorMsg := req.Params.Arguments["error_message"]
	technology := req.Params.Arguments["technology"]

	userPrompt := fmt.Sprintf(`"%s" in %s: causes, verified solutions, workarounds`, errorMsg, technology)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

var securityPrompt = &mcp.Prompt{
	Name:        "security",
	Description: "Check for security vulnerabilities",
	Arguments: []*mcp.PromptArgument{
		{Name: "package", Description: "Package name", Required: true},
		{Name: "date", Description: "Start date (YYYY-MM-DD)", Required: false},
	},
}

func (s *Server) handleSecurity(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	pkg := req.Params.Arguments["package"]
	date := req.Params.Arguments["date"]
	if date == "" {
		date = daysAgo(365)
	}

	userPrompt := fmt.Sprintf("%s security vulnerabilities since %s", pkg, date)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

// Practical prompts

var howtoPrompt = &mcp.Prompt{
	Name:        "howto",
	Description: "Step-by-step guide for a task",
	Arguments: []*mcp.PromptArgument{
		{Name: "task", Description: "Task to accomplish", Required: true},
		{Name: "context", Description: "Context or environment", Required: false},
	},
}

func (s *Server) handleHowto(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	task := req.Params.Arguments["task"]
	contextArg, _ := req.Params.Arguments["context"]

	contextText := ""
	if contextArg != "" {
		contextText = " for " + contextArg
	}

	userPrompt := fmt.Sprintf("%s step-by-step%s: current best practice, required tools, common pitfalls", task, contextText)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

var newsPrompt = &mcp.Prompt{
	Name:        "news",
	Description: "Find recent news and developments",
	Arguments: []*mcp.PromptArgument{
		{Name: "topic", Description: "News topic", Required: true},
	},
}

func (s *Server) handleNews(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	topic := req.Params.Arguments["topic"]

	userPrompt := fmt.Sprintf("%s significant developments", topic)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

// Academic prompt

var academicPrompt = &mcp.Prompt{
	Name:        "academic",
	Description: "Find peer-reviewed research",
	Arguments: []*mcp.PromptArgument{
		{Name: "topic", Description: "Research topic", Required: true},
	},
}

func (s *Server) handleAcademic(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	topic := req.Params.Arguments["topic"]

	userPrompt := fmt.Sprintf("%s: peer-reviewed findings, methodology, key papers", topic)

	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: systemPrompt}},
			{Role: "user", Content: &mcp.TextContent{Text: userPrompt}},
		},
	}, nil
}

// registerPrompts registers all prompts with the MCP server
func (s *Server) registerPrompts() {
	s.mcp.AddPrompt(researchPrompt, s.handleResearch)
	s.mcp.AddPrompt(comparePrompt, s.handleCompare)
	s.mcp.AddPrompt(verifyPrompt, s.handleVerify)
	s.mcp.AddPrompt(statusPrompt, s.handleStatus)
	s.mcp.AddPrompt(docsPrompt, s.handleDocs)
	s.mcp.AddPrompt(errorPrompt, s.handleError)
	s.mcp.AddPrompt(securityPrompt, s.handleSecurity)
	s.mcp.AddPrompt(howtoPrompt, s.handleHowto)
	s.mcp.AddPrompt(newsPrompt, s.handleNews)
	s.mcp.AddPrompt(academicPrompt, s.handleAcademic)
}
