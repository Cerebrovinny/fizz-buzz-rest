package handler

import "net/http"

// StatisticsParams describes the request parameters in the statistics response.
type StatisticsParams struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
}

// StatisticsResponse represents the payload returned by the statistics endpoint.
type StatisticsResponse struct {
	Params StatisticsParams `json:"params"`
	Hits   int              `json:"hits"`
}

// Statistics returns the most frequent FizzBuzz request observed so far.
func (h *Handler) Statistics(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.store == nil {
		respondError(w, http.StatusNotFound, "no statistics available")
		return
	}

	stats, ok := h.store.GetMostFrequent()
	if !ok {
		respondError(w, http.StatusNotFound, "no statistics available")
		return
	}

	response := StatisticsResponse{
		Params: StatisticsParams{
			Int1:  stats.Params.Int1,
			Int2:  stats.Params.Int2,
			Limit: stats.Params.Limit,
			Str1:  stats.Params.Str1,
			Str2:  stats.Params.Str2,
		},
		Hits: stats.Hits,
	}

	respondJSON(w, http.StatusOK, response)
}
