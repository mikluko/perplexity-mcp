package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds the application configuration.
type Config struct {
	Mode       string // "stdio" or "http"
	ListenAddr string // for HTTP mode
	APIKey     string // from PERPLEXITY_API_KEY
}

// LoadConfig parses command-line flags and environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Mode, "mode", "stdio", "Server mode: stdio or http")
	flag.StringVar(&cfg.ListenAddr, "listen", ":8080", "HTTP listen address (http mode only)")
	flag.Parse()

	// Load API key from environment
	cfg.APIKey = os.Getenv("PERPLEXITY_API_KEY")
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("PERPLEXITY_API_KEY environment variable is required")
	}

	// Validate mode
	if cfg.Mode != "stdio" && cfg.Mode != "http" {
		return nil, fmt.Errorf("mode must be 'stdio' or 'http', got: %s", cfg.Mode)
	}

	return cfg, nil
}
