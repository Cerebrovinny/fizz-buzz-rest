package statistics

import "sync"

// RequestParams represents the parameters of a FizzBuzz request.
type RequestParams struct {
	Int1  int
	Int2  int
	Limit int
	Str1  string
	Str2  string
}

// Stats describes how often a specific request was made.
type Stats struct {
	Params RequestParams
	Hits   int
}

// Store tracks request statistics with concurrency safety.
type Store struct {
	mu       sync.RWMutex
	requests map[RequestParams]int
}

// NewStore returns an initialized Store instance.
func NewStore() *Store {
	return &Store{
		requests: make(map[RequestParams]int),
	}
}

// Record increments the hit counter for the provided parameters.
func (s *Store) Record(params RequestParams) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests[params]++
}

// GetMostFrequent returns the most frequent request, if any exist.
func (s *Store) GetMostFrequent() (*Stats, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var (
		maxParams RequestParams
		maxHits   int
		found     bool
	)

	for params, hits := range s.requests {
		if !found || hits > maxHits {
			maxParams = params
			maxHits = hits
			found = true
		}
	}

	if !found {
		return nil, false
	}

	result := Stats{
		Params: maxParams,
		Hits:   maxHits,
	}

	return &result, true
}
