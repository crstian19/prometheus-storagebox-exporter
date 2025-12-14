package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		args           []string
		setupFiles     func() (string, func())
		wantErr        bool
		errContains    string
		expectedToken  string
		expectFileRead bool
	}{
		{
			name: "HETZNER_TOKEN environment variable",
			envVars: map[string]string{
				"HETZNER_TOKEN": "test-token-env",
			},
			expectedToken: "test-token-env",
		},
		{
			name: "HETZNER_TOKEN_FILE environment variable",
			envVars: map[string]string{
				"HETZNER_TOKEN_FILE": "/tmp/test-token-file",
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()
				tokenFile := filepath.Join(tmpDir, "test-token")
				if err := os.WriteFile(tokenFile, []byte("test-token-from-file"), 0600); err != nil {
					t.Fatalf("Failed to create test token file: %v", err)
				}
				return tokenFile, func() {}
			},
			expectedToken:  "test-token-from-file",
			expectFileRead: true,
		},
		{
			name:          "--hetzner-token flag",
			args:          []string{"--hetzner-token=test-token-flag"},
			expectedToken: "test-token-flag",
		},
		{
			name: "--hetzner-token-file flag",
			args: []string{"--hetzner-token-file=/tmp/test-token-file-flag"},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()
				tokenFile := filepath.Join(tmpDir, "test-token")
				if err := os.WriteFile(tokenFile, []byte("test-token-from-file-flag"), 0600); err != nil {
					t.Fatalf("Failed to create test token file: %v", err)
				}
				return tokenFile, func() {}
			},
			expectedToken:  "test-token-from-file-flag",
			expectFileRead: true,
		},
		{
			name: "both HETZNER_TOKEN and HETZNER_TOKEN_FILE environment variables should fail",
			envVars: map[string]string{
				"HETZNER_TOKEN":      "test-token-env",
				"HETZNER_TOKEN_FILE": "/tmp/test-token-file",
			},
			wantErr:     true,
			errContains: "cannot specify both HETZNER_TOKEN and HETZNER_TOKEN_FILE environment variables",
		},
		{
			name: "both --hetzner-token and --hetzner-token-file flags should fail",
			args: []string{
				"--hetzner-token=test-token-flag",
				"--hetzner-token-file=/tmp/test-token-file",
			},
			wantErr:     true,
			errContains: "cannot specify both --hetzner-token and --hetzner-token-file",
		},
		{
			name: "HETZNER_TOKEN env with --hetzner-token-file flag should fail",
			envVars: map[string]string{
				"HETZNER_TOKEN": "test-token-env",
			},
			args:        []string{"--hetzner-token-file=/tmp/test-token-file"},
			wantErr:     true,
			errContains: "cannot specify both --hetzner-token and --hetzner-token-file",
		},
		{
			name: "HETZNER_TOKEN_FILE env with --hetzner-token flag should fail",
			envVars: map[string]string{
				"HETZNER_TOKEN_FILE": "/tmp/test-token-file",
			},
			args:        []string{"--hetzner-token=test-token-flag"},
			wantErr:     true,
			errContains: "cannot specify both --hetzner-token and --hetzner-token-file",
		},
		{
			name:        "no token provided should fail",
			wantErr:     true,
			errContains: "HETZNER_TOKEN or HETZNER_TOKEN_FILE environment variable is required (or corresponding flags)",
		},
		{
			name: "empty token file should fail",
			envVars: map[string]string{
				"HETZNER_TOKEN_FILE": "/tmp/empty-token-file",
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()
				tokenFile := filepath.Join(tmpDir, "empty-token")
				if err := os.WriteFile(tokenFile, []byte(""), 0600); err != nil {
					t.Fatalf("Failed to create empty test token file: %v", err)
				}
				return tokenFile, func() {}
			},
			wantErr:     true,
			errContains: "failed to read token from file",
		},
		{
			name: "non-existent token file should fail",
			envVars: map[string]string{
				"HETZNER_TOKEN_FILE": "/tmp/non-existent-token-file",
			},
			wantErr:     true,
			errContains: "failed to read token from file",
		},
		{
			name: "token file with whitespace should be trimmed",
			envVars: map[string]string{
				"HETZNER_TOKEN_FILE": "/tmp/whitespace-token-file",
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()
				tokenFile := filepath.Join(tmpDir, "whitespace-token")
				if err := os.WriteFile(tokenFile, []byte("  test-token-whitespace  \n"), 0600); err != nil {
					t.Fatalf("Failed to create whitespace test token file: %v", err)
				}
				return tokenFile, func() {}
			},
			expectedToken:  "test-token-whitespace",
			expectFileRead: true,
		},
		{
			name:          "version flag should not require token",
			args:          []string{"--version"},
			wantErr:       false,
			expectedToken: "",
		},
		{
			name: "default configuration values",
			envVars: map[string]string{
				"HETZNER_TOKEN": "test-token",
			},
			expectedToken: "test-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset pflag for each test
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

			// Store original environment variables
			originalEnv := make(map[string]string)
			for k := range tt.envVars {
				originalEnv[k] = os.Getenv(k)
			}

			// Set up environment variables
			for k, v := range tt.envVars {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf("Failed to set environment variable %s: %v", k, err)
				}
			}

			// Clean up environment variables after test
			defer func() {
				for k := range tt.envVars {
					if original, exists := originalEnv[k]; exists {
						if err := os.Setenv(k, original); err != nil {
							t.Logf("Failed to restore environment variable %s: %v", k, err)
						}
					} else {
						if err := os.Unsetenv(k); err != nil {
							t.Logf("Failed to unset environment variable %s: %v", k, err)
						}
					}
				}
			}()

			// Set up test files if needed
			if tt.setupFiles != nil {
				tokenFile, cleanup := tt.setupFiles()
				defer cleanup()

				// Update environment variable or args with the actual temp file path
				if tt.envVars != nil {
					if _, exists := tt.envVars["HETZNER_TOKEN_FILE"]; exists {
						tt.envVars["HETZNER_TOKEN_FILE"] = tokenFile
						if err := os.Setenv("HETZNER_TOKEN_FILE", tokenFile); err != nil {
							t.Fatalf("Failed to set HETZNER_TOKEN_FILE environment variable: %v", err)
						}
					}
				}

				// Update args with the actual temp file path
				for i, arg := range tt.args {
					if arg == "--hetzner-token-file=/tmp/test-token-file" ||
						arg == "--hetzner-token-file=/tmp/test-token-file-flag" ||
						arg == "--hetzner-token-file=/tmp/non-existent-token-file" ||
						arg == "--hetzner-token-file=/tmp/empty-token-file" ||
						arg == "--hetzner-token-file=/tmp/whitespace-token-file" {
						tt.args[i] = "--hetzner-token-file=" + tokenFile
					}
				}
			}

			// Reset command line args
			os.Args = append([]string{"test"}, tt.args...)

			// Load configuration
			cfg, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, wantErrContains %v", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error = %v", err)
				return
			}

			if cfg.HetznerToken != tt.expectedToken {
				t.Errorf("Load() HetznerToken = %v, want %v", cfg.HetznerToken, tt.expectedToken)
			}

			// Check default values
			if cfg.ListenAddress != ":9509" {
				t.Errorf("Load() ListenAddress = %v, want :9509", cfg.ListenAddress)
			}
			if cfg.MetricsPath != "/metrics" {
				t.Errorf("Load() MetricsPath = %v, want /metrics", cfg.MetricsPath)
			}
			if cfg.LogLevel != "info" {
				t.Errorf("Load() LogLevel = %v, want info", cfg.LogLevel)
			}
		})
	}
}

func TestReadTokenFromFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		want        string
		wantErr     bool
	}{
		{
			name:        "valid token",
			fileContent: "test-token-123",
			want:        "test-token-123",
		},
		{
			name:        "token with whitespace and newline",
			fileContent: "  test-token-456  \n",
			want:        "test-token-456",
		},
		{
			name:        "empty file",
			fileContent: "",
			wantErr:     true,
		},
		{
			name:        "file with only whitespace",
			fileContent: "   \n  \t  ",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tokenFile := filepath.Join(tmpDir, "token")

			err := os.WriteFile(tokenFile, []byte(tt.fileContent), 0600)
			if err != nil {
				t.Fatalf("Failed to create test token file: %v", err)
			}

			got, err := readTokenFromFile(tokenFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("readTokenFromFile() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("readTokenFromFile() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("readTokenFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadCacheConfiguration(t *testing.T) {
	tests := []struct {
		name                string
		envVars             map[string]string
		args                []string
		expectedTTL         time.Duration
		expectedMaxSize     int64
		expectedCleanupInt  time.Duration
		expectedStorageType string
	}{
		{
			name: "default cache values",
			envVars: map[string]string{
				"HETZNER_TOKEN": "test-token",
			},
			expectedTTL:         0,
			expectedMaxSize:     0,
			expectedCleanupInt:  10 * time.Second,
			expectedStorageType: "memory",
		},
		{
			name: "cache configuration from environment",
			envVars: map[string]string{
				"HETZNER_TOKEN":          "test-token",
				"CACHE_TTL":              "60",
				"CACHE_MAX_SIZE":         "1048576",
				"CACHE_CLEANUP_INTERVAL": "30",
				"CACHE_STORAGE_TYPE":     "redis",
			},
			expectedTTL:         60 * time.Second,
			expectedMaxSize:     1048576,
			expectedCleanupInt:  30 * time.Second,
			expectedStorageType: "redis",
		},
		{
			name: "cache configuration from flags",
			args: []string{
				"--hetzner-token=test-token",
				"--cache-ttl=120",
				"--cache-max-size=2097152",
				"--cache-cleanup-interval=45",
				"--cache-storage-type=memory",
			},
			expectedTTL:         120 * time.Second,
			expectedMaxSize:     2097152,
			expectedCleanupInt:  45 * time.Second,
			expectedStorageType: "memory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset pflag for each test
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

			// Store original environment variables
			originalEnv := make(map[string]string)
			for k := range tt.envVars {
				originalEnv[k] = os.Getenv(k)
			}

			// Set up environment variables
			for k, v := range tt.envVars {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf("Failed to set environment variable %s: %v", k, err)
				}
			}

			// Clean up environment variables after test
			defer func() {
				for k := range tt.envVars {
					if original, exists := originalEnv[k]; exists {
						if err := os.Setenv(k, original); err != nil {
							t.Logf("Failed to restore environment variable %s: %v", k, err)
						}
					} else {
						if err := os.Unsetenv(k); err != nil {
							t.Logf("Failed to unset environment variable %s: %v", k, err)
						}
					}
				}
			}()

			// Reset command line args
			os.Args = append([]string{"test"}, tt.args...)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() unexpected error = %v", err)
			}

			if cfg.CacheTTL != tt.expectedTTL {
				t.Errorf("Load() CacheTTL = %v, want %v", cfg.CacheTTL, tt.expectedTTL)
			}
			if cfg.CacheMaxSize != tt.expectedMaxSize {
				t.Errorf("Load() CacheMaxSize = %v, want %v", cfg.CacheMaxSize, tt.expectedMaxSize)
			}
			if cfg.CacheCleanupInterval != tt.expectedCleanupInt {
				t.Errorf("Load() CacheCleanupInterval = %v, want %v", cfg.CacheCleanupInterval, tt.expectedCleanupInt)
			}
			if cfg.CacheStorageType != tt.expectedStorageType {
				t.Errorf("Load() CacheStorageType = %v, want %v", cfg.CacheStorageType, tt.expectedStorageType)
			}
		})
	}
}
