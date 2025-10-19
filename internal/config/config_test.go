package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expected := &Config{
		Port:               "8080",
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		IdleTimeout:        60 * time.Second,
		RequestTimeout:     60 * time.Second,
		ShutdownTimeout:    30 * time.Second,
		LogLevel:           "info",
		LogFormat:          "json",
		CORSAllowedOrigins: []string{"*"},
	}

	assertConfig(t, cfg, expected)
}

func TestLoad_CustomValues(t *testing.T) {
	tests := []struct {
		name     string
		vars     map[string]string
		expected *Config
	}{
		{
			name: "all custom",
			vars: map[string]string{
				"PORT":                 "3000",
				"READ_TIMEOUT":         "5s",
				"WRITE_TIMEOUT":        "10s",
				"IDLE_TIMEOUT":         "2m",
				"REQUEST_TIMEOUT":      "90s",
				"SHUTDOWN_TIMEOUT":     "45s",
				"LOG_LEVEL":            "debug",
				"LOG_FORMAT":           "text",
				"CORS_ALLOWED_ORIGINS": "https://example.com,https://app.example.com",
			},
			expected: &Config{
				Port:               "3000",
				ReadTimeout:        5 * time.Second,
				WriteTimeout:       10 * time.Second,
				IdleTimeout:        2 * time.Minute,
				RequestTimeout:     90 * time.Second,
				ShutdownTimeout:    45 * time.Second,
				LogLevel:           "debug",
				LogFormat:          "text",
				CORSAllowedOrigins: []string{"https://example.com", "https://app.example.com"},
			},
		},
		{
			name: "mixed defaults",
			vars: map[string]string{
				"PORT":                 "",
				"READ_TIMEOUT":         "20s",
				"REQUEST_TIMEOUT":      "120s",
				"LOG_LEVEL":            "warn",
				"CORS_ALLOWED_ORIGINS": "https://example.com",
			},
			expected: &Config{
				Port:               "8080",
				ReadTimeout:        20 * time.Second,
				WriteTimeout:       15 * time.Second,
				IdleTimeout:        60 * time.Second,
				RequestTimeout:     120 * time.Second,
				ShutdownTimeout:    30 * time.Second,
				LogLevel:           "warn",
				LogFormat:          "json",
				CORSAllowedOrigins: []string{"https://example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			setEnvVars(t, tt.vars)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			assertConfig(t, cfg, tt.expected)
		})
	}
}

func TestLoad_InvalidDuration(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  string
	}{
		{"read timeout", "READ_TIMEOUT", "invalid"},
		{"write timeout", "WRITE_TIMEOUT", "5x"},
		{"idle timeout", "IDLE_TIMEOUT", "abc"},
		{"request timeout", "REQUEST_TIMEOUT", "ten"},
		{"shutdown timeout", "SHUTDOWN_TIMEOUT", "not-a-duration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			setEnvVars(t, map[string]string{tt.key: tt.val})

			if _, err := Load(); err == nil {
				t.Fatalf("Load() error = nil, want error")
			}
		})
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	tests := []string{"invalid", "INFO", "trace", ""}
	for _, val := range tests {
		t.Run(val, func(t *testing.T) {
			clearEnv(t)
			setEnvVars(t, map[string]string{"LOG_LEVEL": val})

			if _, err := Load(); err == nil {
				t.Fatalf("Load() error = nil, want error")
			}
		})
	}
}

func TestLoad_InvalidLogFormat(t *testing.T) {
	tests := []string{"xml", "yaml", "JSON", ""}
	for _, val := range tests {
		t.Run(val, func(t *testing.T) {
			clearEnv(t)
			setEnvVars(t, map[string]string{"LOG_FORMAT": val})

			if _, err := Load(); err == nil {
				t.Fatalf("Load() error = nil, want error")
			}
		})
	}
}

func TestLoad_ZeroTimeout(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  string
	}{
		{"read timeout zero", "READ_TIMEOUT", "0s"},
		{"write timeout negative", "WRITE_TIMEOUT", "-5s"},
		{"idle timeout zero", "IDLE_TIMEOUT", "0ms"},
		{"request timeout zero", "REQUEST_TIMEOUT", "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			setEnvVars(t, map[string]string{tt.key: tt.val})

			if _, err := Load(); err == nil {
				t.Fatalf("Load() error = nil, want error")
			}
		})
	}
}

func TestLoad_CORSOrigins(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{"single", "https://example.com", []string{"https://example.com"}},
		{"multiple", "https://example.com,https://app.example.com", []string{"https://example.com", "https://app.example.com"}},
		{"with spaces", "https://example.com, https://app.example.com", []string{"https://example.com", "https://app.example.com"}},
		{"wildcard", "*", []string{"*"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			setEnvVars(t, map[string]string{"CORS_ALLOWED_ORIGINS": tt.value})

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if got := cfg.CORSAllowedOrigins; !equalStringSlices(got, tt.expected) {
				t.Fatalf("CORSAllowedOrigins = %v, want %v", got, tt.expected)
			}
		})
	}
}

func setEnvVars(t *testing.T, vars map[string]string) {
	t.Helper()
	for key, value := range vars {
		t.Setenv(key, value)
	}
}

func assertConfig(t *testing.T, cfg *Config, expected *Config) {
	t.Helper()
	if cfg.Port != expected.Port {
		t.Fatalf("Port = %s, want %s", cfg.Port, expected.Port)
	}
	if cfg.ReadTimeout != expected.ReadTimeout {
		t.Fatalf("ReadTimeout = %s, want %s", cfg.ReadTimeout, expected.ReadTimeout)
	}
	if cfg.WriteTimeout != expected.WriteTimeout {
		t.Fatalf("WriteTimeout = %s, want %s", cfg.WriteTimeout, expected.WriteTimeout)
	}
	if cfg.IdleTimeout != expected.IdleTimeout {
		t.Fatalf("IdleTimeout = %s, want %s", cfg.IdleTimeout, expected.IdleTimeout)
	}
	if cfg.RequestTimeout != expected.RequestTimeout {
		t.Fatalf("RequestTimeout = %s, want %s", cfg.RequestTimeout, expected.RequestTimeout)
	}
	if cfg.ShutdownTimeout != expected.ShutdownTimeout {
		t.Fatalf("ShutdownTimeout = %s, want %s", cfg.ShutdownTimeout, expected.ShutdownTimeout)
	}
	if cfg.LogLevel != expected.LogLevel {
		t.Fatalf("LogLevel = %s, want %s", cfg.LogLevel, expected.LogLevel)
	}
	if cfg.LogFormat != expected.LogFormat {
		t.Fatalf("LogFormat = %s, want %s", cfg.LogFormat, expected.LogFormat)
	}
	if !equalStringSlices(cfg.CORSAllowedOrigins, expected.CORSAllowedOrigins) {
		t.Fatalf("CORSAllowedOrigins = %v, want %v", cfg.CORSAllowedOrigins, expected.CORSAllowedOrigins)
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"PORT",
		"READ_TIMEOUT",
		"WRITE_TIMEOUT",
		"IDLE_TIMEOUT",
		"REQUEST_TIMEOUT",
		"SHUTDOWN_TIMEOUT",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"CORS_ALLOWED_ORIGINS",
	}
	for _, key := range keys {
		unsetEnv(t, key)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset %s: %v", key, err)
	}
}
