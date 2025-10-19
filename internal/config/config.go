package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config contains all runtime configuration derived from environment variables.
// Environment variables:
// - PORT: HTTP server port (default: 8080)
// - READ_TIMEOUT: HTTP read timeout, e.g. "15s" (default: 15s)
// - WRITE_TIMEOUT: HTTP write timeout, e.g. "15s" (default: 15s)
// - IDLE_TIMEOUT: HTTP idle timeout, e.g. "60s" (default: 60s)
// - REQUEST_TIMEOUT: Per-request timeout, e.g. "60s" (default: 60s)
// - SHUTDOWN_TIMEOUT: Graceful shutdown timeout, e.g. "30s" (default: 30s)
// - LOG_LEVEL: Log level - debug, info, warn, error (default: info)
// - LOG_FORMAT: Log format - json, text (default: json)
// - CORS_ALLOWED_ORIGINS: Comma-separated CORS origins, e.g. "https://example.com,https://app.example.com" (default: *)
type Config struct {
	Port               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	RequestTimeout     time.Duration
	ShutdownTimeout    time.Duration
	LogLevel           string
	LogFormat          string
	CORSAllowedOrigins []string
}

var (
	allowedLogLevels = map[string]struct{}{
		"debug": {},
		"info":  {},
		"warn":  {},
		"error": {},
	}
	allowedLogFormats = map[string]struct{}{
		"json": {},
		"text": {},
	}
)

// Load populates the Config struct with environment variables and validates the result.
func Load() (*Config, error) {
	cfg := &Config{}

	var err error

	cfg.Port = getEnv("PORT", "8080")
	if cfg.Port == "" {
		return nil, errors.New("port must not be empty")
	}

	if cfg.ReadTimeout, err = parseDuration("READ_TIMEOUT", "15s"); err != nil {
		return nil, err
	}
	if cfg.WriteTimeout, err = parseDuration("WRITE_TIMEOUT", "15s"); err != nil {
		return nil, err
	}
	if cfg.IdleTimeout, err = parseDuration("IDLE_TIMEOUT", "60s"); err != nil {
		return nil, err
	}
	if cfg.RequestTimeout, err = parseDuration("REQUEST_TIMEOUT", "60s"); err != nil {
		return nil, err
	}
	if cfg.ShutdownTimeout, err = parseDuration("SHUTDOWN_TIMEOUT", "30s"); err != nil {
		return nil, err
	}

	if err = validatePositiveDuration("READ_TIMEOUT", cfg.ReadTimeout); err != nil {
		return nil, err
	}
	if err = validatePositiveDuration("WRITE_TIMEOUT", cfg.WriteTimeout); err != nil {
		return nil, err
	}
	if err = validatePositiveDuration("IDLE_TIMEOUT", cfg.IdleTimeout); err != nil {
		return nil, err
	}
	if err = validatePositiveDuration("REQUEST_TIMEOUT", cfg.RequestTimeout); err != nil {
		return nil, err
	}
	if err = validatePositiveDuration("SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout); err != nil {
		return nil, err
	}

	cfg.LogLevel = getEnv("LOG_LEVEL", "info")
	if value, ok := os.LookupEnv("LOG_LEVEL"); ok && strings.TrimSpace(value) == "" {
		return nil, errors.New("invalid log level: value cannot be empty")
	}
	if _, ok := allowedLogLevels[cfg.LogLevel]; !ok {
		return nil, fmt.Errorf("invalid log level: %s", cfg.LogLevel)
	}

	cfg.LogFormat = getEnv("LOG_FORMAT", "json")
	if value, ok := os.LookupEnv("LOG_FORMAT"); ok && strings.TrimSpace(value) == "" {
		return nil, errors.New("invalid log format: value cannot be empty")
	}
	if _, ok := allowedLogFormats[cfg.LogFormat]; !ok {
		return nil, fmt.Errorf("invalid log format: %s", cfg.LogFormat)
	}

	cfg.CORSAllowedOrigins = parseStringSlice("CORS_ALLOWED_ORIGINS", "*")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return defaultValue
}

func parseDuration(key, defaultValue string) (time.Duration, error) {
	value := getEnv(key, defaultValue)
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}
	return d, nil
}

func parseStringSlice(key, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{defaultValue}
	}
	return result
}

func validatePositiveDuration(name string, d time.Duration) error {
	if d <= 0 {
		return fmt.Errorf("%s must be greater than zero", strings.ToLower(name))
	}
	return nil
}
