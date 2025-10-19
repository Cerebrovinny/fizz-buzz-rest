package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Cerebrovinny/fizz-buzz-rest/internal/statistics"
)

func TestHandler_Health_ReturnsOK(t *testing.T) {
	h := NewHandler(statistics.NewStore(), nil)
	rec := callHealthHandler(t, h)

	res := rec.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	if contentType := res.Header.Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", contentType)
	}

	if cacheControl := res.Header.Get("Cache-Control"); cacheControl != "no-store" {
		t.Fatalf("expected Cache-Control no-store, got %s", cacheControl)
	}

	body := rec.Body.Bytes()
	assertHealthResponse(t, body)
}

func TestHandler_Health_CacheControl(t *testing.T) {
	h := NewHandler(statistics.NewStore(), nil)
	rec := callHealthHandler(t, h)

	if cacheControl := rec.Result().Header.Get("Cache-Control"); cacheControl != "no-store" {
		t.Fatalf("expected Cache-Control no-store, got %s", cacheControl)
	}
}

func TestHandler_Health_JSONFormat(t *testing.T) {
	h := NewHandler(statistics.NewStore(), nil)
	rec := callHealthHandler(t, h)

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(payload) != 2 {
		t.Fatalf("expected 2 fields in health response, got %d", len(payload))
	}

	if status, ok := payload["status"].(string); !ok || status != "ok" {
		t.Fatalf("expected status 'ok', got %v", payload["status"])
	}

	if service, ok := payload["service"].(string); !ok || service != "fizzbuzz-api" {
		t.Fatalf("expected service 'fizzbuzz-api', got %v", payload["service"])
	}
}

func TestHandler_Health_ThroughRouter(t *testing.T) {
	h := NewHandler(statistics.NewStore(), nil)
	router := chi.NewRouter()
	router.Get("/health", h.Health)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	assertHealthResponse(t, body)
}

func TestHandler_Health_MultipleRequests(t *testing.T) {
	h := NewHandler(statistics.NewStore(), nil)

	for i := 0; i < 100; i++ {
		rec := callHealthHandler(t, h)
		if rec.Code != http.StatusOK {
			t.Fatalf("iteration %d: expected status %d, got %d", i, http.StatusOK, rec.Code)
		}
		assertHealthResponse(t, rec.Body.Bytes())
	}
}

func callHealthHandler(t *testing.T, h *Handler) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Health(rec, req)
	return rec
}

func assertHealthResponse(t *testing.T, body []byte) {
	t.Helper()
	var resp HealthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal health response: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("expected status 'ok', got %q", resp.Status)
	}
	if resp.Service != "fizzbuzz-api" {
		t.Fatalf("expected service 'fizzbuzz-api', got %q", resp.Service)
	}
}
