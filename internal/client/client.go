package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	apiURL      = "https://api.perplexity.ai/chat/completions"
	asyncAPIURL = "https://api.perplexity.ai/async/chat/completions"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Citation struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Date    string `json:"date,omitempty"`
	Snippet string `json:"snippet,omitempty"`
}

type Response struct {
	Choices       []Choice   `json:"choices"`
	Citations     []string   `json:"citations,omitempty"`      // URL strings (older format)
	SearchResults []Citation `json:"search_results,omitempty"` // Full citation objects
}

type QueryResult struct {
	Content   string
	Citations []Citation
}

func (c *Client) Query(ctx context.Context, model, systemPrompt, query string) (*QueryResult, error) {
	req := Request{
		Model: model,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: query,
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Perplexity API request failed: %v", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read Perplexity API response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Perplexity API error (status %d): %s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	result := &QueryResult{
		Content: response.Choices[0].Message.Content,
	}

	// Collect citations from search_results (preferred) or citations field
	if len(response.SearchResults) > 0 {
		result.Citations = response.SearchResults
	} else if len(response.Citations) > 0 {
		// Convert URL strings to Citation objects
		for _, url := range response.Citations {
			result.Citations = append(result.Citations, Citation{URL: url})
		}
	}

	return result, nil
}

// Async API types

type AsyncCreateResponse struct {
	ID string `json:"id"`
}

type AsyncStatusResponse struct {
	ID          string   `json:"id"`
	CreatedAt   int64    `json:"created_at,omitempty"`
	StartedAt   int64    `json:"started_at,omitempty"`
	CompletedAt int64    `json:"completed_at,omitempty"`
	FailedAt    int64    `json:"failed_at,omitempty"`
	Response    Response `json:"response,omitempty"`
	Error       string   `json:"error,omitempty"`
}

func (r *AsyncStatusResponse) Status() string {
	if r.FailedAt > 0 {
		return "failed"
	}
	if r.CompletedAt > 0 {
		return "completed"
	}
	if r.StartedAt > 0 {
		return "in_progress"
	}
	return "pending"
}

type AsyncRequest struct {
	Request Request `json:"request"`
}

func (c *Client) StartResearch(ctx context.Context, query string) (string, error) {
	req := AsyncRequest{
		Request: Request{
			Model: "sonar-deep-research",
			Messages: []Message{
				{
					Role:    "system",
					Content: "You are a research assistant. Provide comprehensive, well-structured analysis with citations and multiple perspectives.",
				},
				{
					Role:    "user",
					Content: query,
				},
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", asyncAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Perplexity async API request failed: %v", err)
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read Perplexity async API response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		log.Printf("Perplexity async API error (status %d): %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var asyncResp AsyncCreateResponse
	if err := json.Unmarshal(respBody, &asyncResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return asyncResp.ID, nil
}

func (c *Client) GetResearchResult(ctx context.Context, requestID string) (*AsyncStatusResponse, error) {
	url := asyncAPIURL + "/" + requestID

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Perplexity async status request failed: %v", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read Perplexity async status response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Perplexity async status error (status %d): %s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var statusResp AsyncStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &statusResp, nil
}

func (c *Client) WaitForResearch(ctx context.Context, requestID string, pollInterval time.Duration) (*AsyncStatusResponse, error) {
	for {
		status, err := c.GetResearchResult(ctx, requestID)
		if err != nil {
			return nil, err
		}

		switch status.Status() {
		case "completed", "failed":
			return status, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
			// continue polling
		}
	}
}
