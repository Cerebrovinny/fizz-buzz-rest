package statistics

import (
	"sync"
	"testing"
	"testing/synctest"
)

func TestStore_Record_Sequential(t *testing.T) {
	t.Helper()

	tests := []struct {
		name    string
		records []struct {
			params RequestParams
			count  int
		}
		wantParams RequestParams
		wantHits   int
		wantOK     bool
	}{
		{
			name: "single request recorded once",
			records: []struct {
				params RequestParams
				count  int
			}{
				{params: createParams(3, 5, 15, "fizz", "buzz"), count: 1},
			},
			wantParams: createParams(3, 5, 15, "fizz", "buzz"),
			wantHits:   1,
			wantOK:     true,
		},
		{
			name: "same request recorded multiple times",
			records: []struct {
				params RequestParams
				count  int
			}{
				{params: createParams(3, 5, 15, "fizz", "buzz"), count: 5},
			},
			wantParams: createParams(3, 5, 15, "fizz", "buzz"),
			wantHits:   5,
			wantOK:     true,
		},
		{
			name: "multiple different requests",
			records: []struct {
				params RequestParams
				count  int
			}{
				{params: createParams(3, 5, 15, "fizz", "buzz"), count: 3},
				{params: createParams(2, 4, 20, "foo", "bar"), count: 5},
				{params: createParams(7, 11, 50, "seven", "eleven"), count: 2},
			},
			wantParams: createParams(2, 4, 20, "foo", "bar"),
			wantHits:   5,
			wantOK:     true,
		},
		{
			name:     "empty store",
			wantOK:   false,
			wantHits: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()

			for _, record := range tt.records {
				for i := 0; i < record.count; i++ {
					store.Record(record.params)
				}
			}

			stats, ok := store.GetMostFrequent()

			if ok != tt.wantOK {
				t.Fatalf("expected ok %t, got %t", tt.wantOK, ok)
			}

			if !ok {
				if stats != nil {
					t.Fatalf("expected nil stats when ok is false")
				}
				return
			}

			assertStats(t, stats, tt.wantParams, tt.wantHits)
		})
	}
}

func TestStore_Record_Concurrent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store := NewStore()
		params := createParams(3, 5, 15, "fizz", "buzz")

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Go(func() {
				store.Record(params)
			})
		}

		wg.Wait()

		stats, ok := store.GetMostFrequent()
		if !ok {
			t.Fatal("expected statistics to be available")
		}

		assertStats(t, stats, params, 100)
	})
}

func TestStore_GetMostFrequent_Concurrent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store := NewStore()

		base := createParams(3, 5, 15, "fizz", "buzz")
		for i := 0; i < 10; i++ {
			store.Record(base)
		}

		var wg sync.WaitGroup

		for i := 0; i < 50; i++ {
			wg.Go(func() {
				if _, ok := store.GetMostFrequent(); !ok {
					t.Log("expected statistics to exist during concurrent reads")
				}
			})
		}

		for i := 0; i < 10; i++ {
			wg.Go(func() {
				store.Record(base)
			})
		}

		wg.Wait()

		stats, ok := store.GetMostFrequent()
		if !ok {
			t.Fatal("expected statistics to be available after concurrent operations")
		}

		assertStats(t, stats, base, 20)
	})
}

func TestStore_MultipleRequests_FindMax(t *testing.T) {
	tests := []struct {
		name    string
		records []struct {
			params RequestParams
			count  int
		}
		wantHits       int
		wantParams     RequestParams
		allowedOptions []RequestParams
	}{
		{
			name: "clear winner",
			records: []struct {
				params RequestParams
				count  int
			}{
				{params: createParams(1, 2, 30, "foo", "bar"), count: 100},
				{params: createParams(3, 4, 30, "baz", "qux"), count: 10},
				{params: createParams(5, 6, 30, "spam", "eggs"), count: 5},
			},
			wantParams: createParams(1, 2, 30, "foo", "bar"),
			wantHits:   100,
		},
		{
			name: "tie scenario",
			records: []struct {
				params RequestParams
				count  int
			}{
				{params: createParams(1, 2, 10, "a", "b"), count: 5},
				{params: createParams(3, 4, 10, "c", "d"), count: 5},
			},
			wantHits:       5,
			allowedOptions: []RequestParams{createParams(1, 2, 10, "a", "b"), createParams(3, 4, 10, "c", "d")},
		},
		{
			name: "all equal",
			records: []struct {
				params RequestParams
				count  int
			}{
				{params: createParams(1, 2, 10, "x", "y"), count: 1},
				{params: createParams(3, 4, 10, "m", "n"), count: 1},
				{params: createParams(5, 6, 10, "p", "q"), count: 1},
			},
			wantHits:       1,
			allowedOptions: []RequestParams{createParams(1, 2, 10, "x", "y"), createParams(3, 4, 10, "m", "n"), createParams(5, 6, 10, "p", "q")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()

			for _, record := range tt.records {
				for i := 0; i < record.count; i++ {
					store.Record(record.params)
				}
			}

			stats, ok := store.GetMostFrequent()
			if !ok {
				t.Fatalf("expected statistics for test %s", tt.name)
			}

			if stats.Hits != tt.wantHits {
				t.Fatalf("expected hits %d, got %d", tt.wantHits, stats.Hits)
			}

			if len(tt.allowedOptions) > 0 {
				match := false
				for _, option := range tt.allowedOptions {
					if stats.Params == option {
						match = true
						break
					}
				}

				if !match {
					t.Fatalf("returned params %+v not in allowed options %+v", stats.Params, tt.allowedOptions)
				}
				return
			}

			if stats.Params != tt.wantParams {
				t.Fatalf("expected params %+v, got %+v", tt.wantParams, stats.Params)
			}
		})
	}
}

func TestRequestParams_AsMapKey(t *testing.T) {
	paramsA := createParams(3, 5, 15, "fizz", "buzz")
	paramsB := createParams(3, 5, 15, "fizz", "buzz")
	paramsC := createParams(2, 4, 20, "foo", "bar")

	if paramsA != paramsB {
		t.Fatal("expected identical params to be equal")
	}

	requests := map[RequestParams]int{
		paramsA: 1,
	}

	requests[paramsB]++
	requests[paramsC] = 5

	if requests[paramsA] != 2 {
		t.Fatalf("expected combined hits for identical params to be 2, got %d", requests[paramsA])
	}

	if requests[paramsC] != 5 {
		t.Fatalf("expected hits for distinct params to remain unaffected, got %d", requests[paramsC])
	}
}

func assertStats(t *testing.T, got *Stats, wantParams RequestParams, wantHits int) {
	t.Helper()

	if got == nil {
		t.Fatal("expected non-nil stats result")
	}

	if got.Params != wantParams {
		t.Fatalf("expected params %+v, got %+v", wantParams, got.Params)
	}

	if got.Hits != wantHits {
		t.Fatalf("expected hits %d, got %d", wantHits, got.Hits)
	}
}

func createParams(int1, int2, limit int, str1, str2 string) RequestParams {
	return RequestParams{
		Int1:  int1,
		Int2:  int2,
		Limit: limit,
		Str1:  str1,
		Str2:  str2,
	}
}
