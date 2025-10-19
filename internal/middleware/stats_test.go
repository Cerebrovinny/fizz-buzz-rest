package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"testing/synctest"

	"github.com/go-chi/chi/v5"

	"github.com/Cerebrovinny/fizz-buzz-rest/internal/statistics"
)

func TestStatistics_RecordsValidRequest(t *testing.T) {
	store := statistics.NewStore()
	mw := Statistics(store)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := mw(handler)

	rec := makeRequest(t, wrapped, "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	assertRecorded(t, store, statistics.RequestParams{
		Int1:  3,
		Int2:  5,
		Limit: 15,
		Str1:  "fizz",
		Str2:  "buzz",
	}, 1)
}

func TestStatistics_RecordsMultipleRequests(t *testing.T) {
	store := statistics.NewStore()
	mw := Statistics(store)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := mw(handler)

	for i := 0; i < 5; i++ {
		makeRequest(t, wrapped, "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz")
	}

	for i := 0; i < 3; i++ {
		makeRequest(t, wrapped, "/fizzbuzz?int1=2&int2=3&limit=10&str1=foo&str2=bar")
	}

	assertRecorded(t, store, statistics.RequestParams{
		Int1:  3,
		Int2:  5,
		Limit: 15,
		Str1:  "fizz",
		Str2:  "buzz",
	}, 5)
}

func TestStatistics_IgnoresInvalidRequests(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "missing int1", query: "/fizzbuzz?int2=5&limit=15&str1=fizz&str2=buzz"},
		{name: "missing int2", query: "/fizzbuzz?int1=3&limit=15&str1=fizz&str2=buzz"},
		{name: "missing limit", query: "/fizzbuzz?int1=3&int2=5&str1=fizz&str2=buzz"},
		{name: "missing str1", query: "/fizzbuzz?int1=3&int2=5&limit=15&str2=buzz"},
		{name: "missing str2", query: "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz"},
		{name: "invalid int1", query: "/fizzbuzz?int1=abc&int2=5&limit=15&str1=fizz&str2=buzz"},
		{name: "invalid int2", query: "/fizzbuzz?int1=3&int2=xyz&limit=15&str1=fizz&str2=buzz"},
		{name: "invalid limit", query: "/fizzbuzz?int1=3&int2=5&limit=abc&str1=fizz&str2=buzz"},
		{name: "no query params", query: "/fizzbuzz"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store := statistics.NewStore()
			mw := Statistics(store)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrapped := mw(handler)

			makeRequest(t, wrapped, tt.query)

			assertNotRecorded(t, store)
		})
	}
}

func TestStatistics_HandlerStillExecutes(t *testing.T) {
	store := statistics.NewStore()
	mw := Statistics(store)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrapped := mw(handler)

	makeRequest(t, wrapped, "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz")
	if !handlerCalled {
		t.Fatal("expected handler to be called for valid request")
	}

	handlerCalled = false
	makeRequest(t, wrapped, "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz")
	if !handlerCalled {
		t.Fatal("expected handler to be called even when request is invalid")
	}
}

func TestStatistics_DoesNotRecordWhenStatusNotOK(t *testing.T) {
	tests := []struct {
		name   string
		status int
		query  string
	}{
		{
			name:   "bad request zero int1",
			status: http.StatusBadRequest,
			query:  "/fizzbuzz?int1=0&int2=5&limit=15&str1=fizz&str2=buzz",
		},
		{
			name:   "bad request negative int2",
			status: http.StatusBadRequest,
			query:  "/fizzbuzz?int1=3&int2=-5&limit=15&str1=fizz&str2=buzz",
		},
		{
			name:   "internal error",
			status: http.StatusInternalServerError,
			query:  "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store := statistics.NewStore()
			mw := Statistics(store)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			})

			wrapped := mw(handler)

			makeRequest(t, wrapped, tt.query)

			assertNotRecorded(t, store)
		})
	}
}

func TestStatistics_ConcurrentRequests(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store := statistics.NewStore()
		mw := Statistics(store)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := mw(handler)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Go(func() {
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz", nil)
				wrapped.ServeHTTP(rec, req)
			})
		}

		wg.Wait()

		assertRecorded(t, store, statistics.RequestParams{
			Int1:  3,
			Int2:  5,
			Limit: 15,
			Str1:  "fizz",
			Str2:  "buzz",
		}, 100)
	})
}

func TestStatistics_DifferentPaths(t *testing.T) {
	store := statistics.NewStore()

	router := chi.NewRouter()
	router.With(Statistics(store)).Get("/fizzbuzz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Get("/other", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz")
	if err != nil {
		t.Fatalf("failed to make fizzbuzz request: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("failed to close fizzbuzz response body: %v", err)
	}

	resp, err = http.Get(server.URL + "/other?int1=3&int2=5&limit=15&str1=fizz&str2=buzz")
	if err != nil {
		t.Fatalf("failed to make other request: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("failed to close other response body: %v", err)
	}

	stats, ok := store.GetMostFrequent()
	if !ok {
		t.Fatal("expected statistics after middleware handled requests")
	}

	if stats.Hits != 1 {
		t.Fatalf("expected hits recorded only for fizzbuzz path, got %d", stats.Hits)
	}
}

func makeRequest(t *testing.T, handler http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func assertRecorded(t *testing.T, store *statistics.Store, expectedParams statistics.RequestParams, expectedHits int) {
	t.Helper()

	stats, ok := store.GetMostFrequent()
	if !ok {
		t.Fatal("expected statistics to be recorded")
	}

	if stats.Params != expectedParams {
		t.Fatalf("expected params %+v, got %+v", expectedParams, stats.Params)
	}

	if stats.Hits != expectedHits {
		t.Fatalf("expected hits %d, got %d", expectedHits, stats.Hits)
	}
}

func assertNotRecorded(t *testing.T, store *statistics.Store) {
	t.Helper()

	stats, ok := store.GetMostFrequent()
	if ok || stats != nil {
		t.Fatal("expected no statistics to be recorded")
	}
}
