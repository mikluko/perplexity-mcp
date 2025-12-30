package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	url := flag.String("url", "http://localhost:8080", "MCP server URL")
	flag.StringVar(url, "u", "http://localhost:8080", "MCP server URL (shorthand)")

	question := flag.String("question", "", "Question to ask")
	flag.StringVar(question, "q", "", "Question to ask (shorthand)")

	flag.Parse()

	if *question == "" {
		fmt.Fprintf(os.Stderr, "Error: -question/-q is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -url <url> -question <question>\n", os.Args[0])
		os.Exit(1)
	}

	ctx := context.Background()

	// Create client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "perplexity-ask-cli",
		Version: "1.0.0",
	}, nil)

	// Connect via HTTP
	transport := &mcp.StreamableClientTransport{
		Endpoint: *url,
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", *url, err)
		os.Exit(1)
	}
	defer session.Close()

	// Call perplexity_ask tool
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "perplexity_ask",
		Arguments: map[string]any{
			"query": *question,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Tool call failed: %v\n", err)
		os.Exit(1)
	}

	// Print result
	if result.IsError {
		fmt.Fprintf(os.Stderr, "Error: ")
		for _, content := range result.Content {
			if textContent, ok := content.(*mcp.TextContent); ok {
				fmt.Fprintf(os.Stderr, "%s\n", textContent.Text)
			}
		}
		os.Exit(1)
	}

	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			fmt.Println(textContent.Text)
		}
	}
}
