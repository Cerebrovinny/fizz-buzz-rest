package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestLogger_LogsRequest(t *testing.T) {
	logger, buf := createTestLogger(t)
	mw := RequestLogger(logger)

	wrapped := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "true")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test response")); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)
	req.Header.Set("User-Agent", "test-agent")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	entry := parseLogEntry(t, buf)
	assertLogString(t, entry, "msg", "http request")
	assertLogString(t, entry, "method", http.MethodGet)
	assertLogString(t, entry, "path", "/test")
	assertLogNumberGreater(t, entry, "duration_ms", 0)
	assertLogNumberEqual(t, entry, "status", 200)
	assertLogNumberEqual(t, entry, "bytes", 13)
	assertLogString(t, entry, "user_agent", "test-agent")
	assertLogStringNotEmpty(t, entry, "remote_addr")
	assertLogString(t, entry, "level", "INFO")
}

func TestRequestLogger_LogLevelByStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		level  string
	}{
		{"status 200", http.StatusOK, "INFO"},
		{"status 301", http.StatusMovedPermanently, "INFO"},
		{"status 404", http.StatusNotFound, "WARN"},
		{"status 503", http.StatusServiceUnavailable, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := createTestLogger(t)
			mw := RequestLogger(logger)
			h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			entry := parseLogEntry(t, buf)
			assertLogString(t, entry, "level", tt.level)
		})
	}
}

func TestRequestLogger_CapturesResponseSize(t *testing.T) {
	tests := []struct {
		name      string
		writeFunc func(http.ResponseWriter) error
		expected  float64
	}{
		{
			name:      "empty response",
			writeFunc: func(w http.ResponseWriter) error { return nil },
			expected:  0,
		},
		{
			name: "small response",
			writeFunc: func(w http.ResponseWriter) error {
				_, err := w.Write([]byte("hello"))
				return err
			},
			expected: 5,
		},
		{
			name: "multiple writes",
			writeFunc: func(w http.ResponseWriter) error {
				if _, err := w.Write([]byte("foo")); err != nil {
					return err
				}
				_, err := w.Write([]byte("bar"))
				return err
			},
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := createTestLogger(t)
			mw := RequestLogger(logger)
			h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := tt.writeFunc(w); err != nil {
					t.Fatalf("writeFunc error = %v", err)
				}
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			entry := parseLogEntry(t, buf)
			assertLogNumberEqual(t, entry, "bytes", tt.expected)
		})
	}
}

func TestRequestLogger_MeasuresDuration(t *testing.T) {
	logger, buf := createTestLogger(t)
	mw := RequestLogger(logger)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	entry := parseLogEntry(t, buf)
	duration := getNumber(t, entry, "duration_ms")
	if duration < 50 {
		t.Fatalf("duration_ms = %v, want >= 50", duration)
	}
	if duration > 150 {
		t.Fatalf("duration_ms = %v, want reasonable upper bound", duration)
	}
}

func TestRequestLogger_HandlerPanics(t *testing.T) {
	logger, buf := createTestLogger(t)
	mw := RequestLogger(logger)

	h := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	defer func() {
		if rec := recover(); rec == nil {
			t.Fatalf("expected panic to propagate")
		}

		entry := parseLogEntry(t, buf)
		assertLogString(t, entry, "path", "/panic")
		assertLogString(t, entry, "level", "ERROR")
		assertLogNumberEqual(t, entry, "status", http.StatusInternalServerError)
		if _, ok := entry["panic"]; !ok {
			t.Fatalf("expected panic field in log entry")
		}
	}()

	h.ServeHTTP(rec, req)
}

func TestRequestLogger_DifferentMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodHead,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			logger, buf := createTestLogger(t)
			mw := RequestLogger(logger)
			h := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

			req := httptest.NewRequest(method, "/method", nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			entry := parseLogEntry(t, buf)
			assertLogString(t, entry, "method", method)
		})
	}
}

func TestRequestLogger_PreservesResponseWriter(t *testing.T) {
	logger, buf := createTestLogger(t)
	mw := RequestLogger(logger)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte("test body")); err != nil {
			panic(err)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	if got := resp.StatusCode; got != http.StatusCreated {
		t.Fatalf("StatusCode = %d, want %d", got, http.StatusCreated)
	}
	if got := resp.Header.Get("X-Custom"); got != "value" {
		t.Fatalf("X-Custom header = %s, want value", got)
	}
	body := rec.Body.String()
	if body != "test body" {
		t.Fatalf("body = %q, want 'test body'", body)
	}

	entry := parseLogEntry(t, buf)
	assertLogNumberEqual(t, entry, "bytes", 9)
	assertLogNumberEqual(t, entry, "status", http.StatusCreated)
}

func createTestLogger(t *testing.T) (*slog.Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), &buf
}

func parseLogEntry(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	decoder := json.NewDecoder(buf)
	entry := make(map[string]any)
	if err := decoder.Decode(&entry); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}
	return entry
}

func assertLogString(t *testing.T, entry map[string]any, key, want string) {
	t.Helper()
	got, ok := entry[key]
	if !ok {
		t.Fatalf("log entry missing key %q", key)
	}
	if gotStr, ok := got.(string); !ok || gotStr != want {
		t.Fatalf("log[%s] = %v, want %s", key, got, want)
	}
}

func assertLogStringNotEmpty(t *testing.T, entry map[string]any, key string) {
	t.Helper()
	got, ok := entry[key]
	if !ok {
		t.Fatalf("log entry missing key %q", key)
	}
	gotStr, ok := got.(string)
	if !ok || gotStr == "" {
		t.Fatalf("log[%s] empty, want non-empty", key)
	}
}

func assertLogNumberEqual(t *testing.T, entry map[string]any, key string, want float64) {
	t.Helper()
	got := getNumber(t, entry, key)
	if got != want {
		t.Fatalf("log[%s] = %v, want %v", key, got, want)
	}
}

func assertLogNumberGreater(t *testing.T, entry map[string]any, key string, threshold float64) {
	t.Helper()
	got := getNumber(t, entry, key)
	if got <= threshold {
		t.Fatalf("log[%s] = %v, want > %v", key, got, threshold)
	}
}

func getNumber(t *testing.T, entry map[string]any, key string) float64 {
	t.Helper()
	got, ok := entry[key]
	if !ok {
		t.Fatalf("log entry missing key %q", key)
	}
	switch v := got.(type) {
	case float64:
		return v
	case json.Number:
		val, err := v.Float64()
		if err != nil {
			t.Fatalf("failed to convert number for key %s: %v", key, err)
		}
		return val
	default:
		t.Fatalf("log[%s] unexpected type %T", key, got)
	}
	return 0
}
