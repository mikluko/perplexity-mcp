package client

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.apiKey != apiKey {
		t.Errorf("NewClient() apiKey = %v, want %v", client.apiKey, apiKey)
	}

	if client.httpClient == nil {
		t.Error("NewClient() httpClient is nil")
	}

	if client.httpClient.Timeout != 120*time.Second {
		t.Errorf("NewClient() httpClient.Timeout = %v, want %v", client.httpClient.Timeout, 120*time.Second)
	}
}

func TestAsyncStatusResponse_Status(t *testing.T) {
	tests := []struct {
		name       string
		response   AsyncStatusResponse
		wantStatus string
	}{
		{
			name:       "pending status",
			response:   AsyncStatusResponse{},
			wantStatus: "pending",
		},
		{
			name:       "in_progress status",
			response:   AsyncStatusResponse{StartedAt: 123456},
			wantStatus: "in_progress",
		},
		{
			name:       "completed status",
			response:   AsyncStatusResponse{StartedAt: 123456, CompletedAt: 123460},
			wantStatus: "completed",
		},
		{
			name:       "failed status",
			response:   AsyncStatusResponse{StartedAt: 123456, FailedAt: 123460},
			wantStatus: "failed",
		},
		{
			name:       "failed overrides completed",
			response:   AsyncStatusResponse{StartedAt: 123456, CompletedAt: 123460, FailedAt: 123460},
			wantStatus: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.response.Status()
			if got != tt.wantStatus {
				t.Errorf("Status() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}
