package config

import (
	"flag"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Save original values
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		name       string
		args       []string
		envKey     string
		wantMode   string
		wantListen string
		wantErr    bool
	}{
		{
			name:       "default stdio mode with API key",
			args:       []string{"cmd"},
			envKey:     "test-key",
			wantMode:   "stdio",
			wantListen: ":8080",
			wantErr:    false,
		},
		{
			name:       "http mode with custom listen",
			args:       []string{"cmd", "-mode", "http", "-listen", ":9000"},
			envKey:     "test-key",
			wantMode:   "http",
			wantListen: ":9000",
			wantErr:    false,
		},
		{
			name:    "missing API key",
			args:    []string{"cmd"},
			envKey:  "",
			wantErr: true,
		},
		{
			name:    "invalid mode",
			args:    []string{"cmd", "-mode", "invalid"},
			envKey:  "test-key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag state
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Set up test environment
			os.Args = tt.args
			if tt.envKey != "" {
				os.Setenv("PERPLEXITY_API_KEY", tt.envKey)
			} else {
				os.Unsetenv("PERPLEXITY_API_KEY")
			}
			defer os.Unsetenv("PERPLEXITY_API_KEY")

			cfg, err := LoadConfig()

			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadConfig() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("LoadConfig() unexpected error: %v", err)
				return
			}

			if cfg.Mode != tt.wantMode {
				t.Errorf("LoadConfig() Mode = %v, want %v", cfg.Mode, tt.wantMode)
			}

			if cfg.ListenAddr != tt.wantListen {
				t.Errorf("LoadConfig() ListenAddr = %v, want %v", cfg.ListenAddr, tt.wantListen)
			}

			if cfg.APIKey != tt.envKey {
				t.Errorf("LoadConfig() APIKey = %v, want %v", cfg.APIKey, tt.envKey)
			}
		})
	}
}
