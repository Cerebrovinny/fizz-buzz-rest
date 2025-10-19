package middleware

import (
	"net/http"
	"strconv"

	"github.com/Cerebrovinny/fizz-buzz-rest/internal/statistics"
)

// Statistics returns middleware that records successful FizzBuzz requests.
func Statistics(store *statistics.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if store == nil {
				next.ServeHTTP(w, r)
				return
			}

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			if rec.status != http.StatusOK {
				return
			}

			query := r.URL.Query()

			int1Str := query.Get("int1")
			int2Str := query.Get("int2")
			limitStr := query.Get("limit")
			str1 := query.Get("str1")
			str2 := query.Get("str2")

			if int1Str == "" || int2Str == "" || limitStr == "" || str1 == "" || str2 == "" {
				return
			}

			int1, err := strconv.Atoi(int1Str)
			if err != nil {
				return
			}

			int2, err := strconv.Atoi(int2Str)
			if err != nil {
				return
			}

			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				return
			}

			store.Record(statistics.RequestParams{
				Int1:  int1,
				Int2:  int2,
				Limit: limit,
				Str1:  str1,
				Str2:  str2,
			})
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}
