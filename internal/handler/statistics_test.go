package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"testing/synctest"

	"github.com/go-chi/chi/v5"

	"github.com/Cerebrovinny/fizz-buzz-rest/internal/statistics"
)

func TestHandler_Statistics_NoData(t *testing.T) {
	store := statistics.NewStore()
	h := NewHandler(store)

	rec := callStatisticsHandler(t, h)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	assertErrorResponse(t, rec.Body.Bytes(), "no statistics available")
}

func TestHandler_Statistics_SingleRequest(t *testing.T) {
	store := statistics.NewStore()
	params := statistics.RequestParams{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"}
	store.Record(params)

	h := NewHandler(store)

	rec := callStatisticsHandler(t, h)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	assertStatisticsResponse(t, rec.Body.Bytes(), params, 1)
}

func TestHandler_Statistics_MultipleRequests(t *testing.T) {
	store := statistics.NewStore()
	mostFrequent := statistics.RequestParams{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"}
	lessFrequent := statistics.RequestParams{Int1: 2, Int2: 3, Limit: 10, Str1: "foo", Str2: "bar"}
	rare := statistics.RequestParams{Int1: 7, Int2: 11, Limit: 20, Str1: "seven", Str2: "eleven"}

	recordRequest(store, mostFrequent, 10)
	recordRequest(store, lessFrequent, 5)
	recordRequest(store, rare, 3)

	h := NewHandler(store)
	rec := callStatisticsHandler(t, h)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	assertStatisticsResponse(t, rec.Body.Bytes(), mostFrequent, 10)
}

func TestHandler_Statistics_UpdatesOverTime(t *testing.T) {
	store := statistics.NewStore()
	early := statistics.RequestParams{Int1: 1, Int2: 2, Limit: 10, Str1: "foo", Str2: "bar"}
	later := statistics.RequestParams{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"}

	recordRequest(store, early, 5)

	h := NewHandler(store)

	rec := callStatisticsHandler(t, h)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	assertStatisticsResponse(t, rec.Body.Bytes(), early, 5)

	recordRequest(store, later, 10)

	rec = callStatisticsHandler(t, h)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	assertStatisticsResponse(t, rec.Body.Bytes(), later, 10)
}

func TestHandler_Statistics_JSONFormat(t *testing.T) {
	store := statistics.NewStore()
	params := statistics.RequestParams{Int1: 8, Int2: 9, Limit: 30, Str1: "eight", Str2: "nine"}
	recordRequest(store, params, 4)

	h := NewHandler(store)
	rec := callStatisticsHandler(t, h)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	paramsValue, ok := payload["params"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected params object in response")
	}

	if len(paramsValue) != 5 {
		t.Fatalf("expected 5 parameters, got %d", len(paramsValue))
	}

	if payload["hits"] != float64(4) {
		t.Fatalf("expected hits 4, got %v", payload["hits"])
	}
}

func TestHandler_Statistics_ThroughRouter(t *testing.T) {
	store := statistics.NewStore()
	params := statistics.RequestParams{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"}
	recordRequest(store, params, 7)

	h := NewHandler(store)

	router := chi.NewRouter()
	router.Get("/statistics", h.Statistics)

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/statistics")
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

	body := readBody(t, resp)
	assertStatisticsResponse(t, body, params, 7)
}

func TestHandler_Statistics_ConcurrentReads(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store := statistics.NewStore()
		params := statistics.RequestParams{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"}
		recordRequest(store, params, 12)

		h := NewHandler(store)

		var wg sync.WaitGroup
		for range 50 {
			wg.Go(func() {
				rec := callStatisticsHandler(t, h)
				if rec.Code != http.StatusOK {
					t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
					return
				}
				assertStatisticsResponse(t, rec.Body.Bytes(), params, 12)
			})
		}

		wg.Wait()
	})
}

func recordRequest(store *statistics.Store, params statistics.RequestParams, times int) {
	for range times {
		store.Record(params)
	}
}

func callStatisticsHandler(t *testing.T, h *Handler) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/statistics", nil)
	rec := httptest.NewRecorder()
	h.Statistics(rec, req)

	return rec
}

func assertStatisticsResponse(t *testing.T, body []byte, expectedParams statistics.RequestParams, expectedHits int) {
	t.Helper()

	var resp StatisticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Hits != expectedHits {
		t.Fatalf("expected hits %d, got %d", expectedHits, resp.Hits)
	}

	if resp.Params.Int1 != expectedParams.Int1 ||
		resp.Params.Int2 != expectedParams.Int2 ||
		resp.Params.Limit != expectedParams.Limit ||
		resp.Params.Str1 != expectedParams.Str1 ||
		resp.Params.Str2 != expectedParams.Str2 {
		t.Fatalf("expected params %+v, got %+v", expectedParams, resp.Params)
	}
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	return body
}
