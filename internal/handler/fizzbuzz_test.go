package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_FizzBuzz(t *testing.T) {
	h := NewHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedBody   interface{}
		checkBody      func(t *testing.T, body []byte)
	}{
		{
			name:           "classic fizzbuzz",
			queryParams:    "int1=3&int2=5&limit=15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusOK,
			expectedBody: FizzBuzzResponse{Result: []string{
				"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz",
			}},
		},
		{
			name:           "custom parameters",
			queryParams:    "int1=2&int2=3&limit=10&str1=foo&str2=bar",
			expectedStatus: http.StatusOK,
			expectedBody: FizzBuzzResponse{Result: []string{
				"1", "foo", "bar", "foo", "5", "foobar", "7", "foo", "bar", "foo",
			}},
		},
		{
			name:           "limit equals one",
			queryParams:    "int1=3&int2=5&limit=1&str1=fizz&str2=buzz",
			expectedStatus: http.StatusOK,
			expectedBody:   FizzBuzzResponse{Result: []string{"1"}},
		},
		{
			name:           "missing int1 parameter",
			queryParams:    "int2=5&limit=15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "missing required parameters: int1, int2, limit, str1, str2"},
		},
		{
			name:           "missing int2 parameter",
			queryParams:    "int1=3&limit=15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "missing required parameters: int1, int2, limit, str1, str2"},
		},
		{
			name:           "missing limit parameter",
			queryParams:    "int1=3&int2=5&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "missing required parameters: int1, int2, limit, str1, str2"},
		},
		{
			name:           "missing str1 parameter",
			queryParams:    "int1=3&int2=5&limit=15&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "missing required parameters: int1, int2, limit, str1, str2"},
		},
		{
			name:           "missing str2 parameter",
			queryParams:    "int1=3&int2=5&limit=15&str1=fizz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "missing required parameters: int1, int2, limit, str1, str2"},
		},
		{
			name:           "invalid int1 parameter",
			queryParams:    "int1=abc&int2=5&limit=15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "int1 must be a valid integer"},
		},
		{
			name:           "invalid int2 parameter",
			queryParams:    "int1=3&int2=xyz&limit=15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "int2 must be a valid integer"},
		},
		{
			name:           "invalid limit parameter",
			queryParams:    "int1=3&int2=5&limit=abc&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "limit must be a valid integer"},
		},
		{
			name:           "zero int1 parameter",
			queryParams:    "int1=0&int2=5&limit=15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "int1 must be greater than 0"},
		},
		{
			name:           "negative int1 parameter",
			queryParams:    "int1=-3&int2=5&limit=15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "int1 must be greater than 0"},
		},
		{
			name:           "zero int2 parameter",
			queryParams:    "int1=3&int2=0&limit=15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "int2 must be greater than 0"},
		},
		{
			name:           "negative int2 parameter",
			queryParams:    "int1=3&int2=-5&limit=15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "int2 must be greater than 0"},
		},
		{
			name:           "zero limit parameter",
			queryParams:    "int1=3&int2=5&limit=0&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "limit must be greater than 0"},
		},
		{
			name:           "negative limit parameter",
			queryParams:    "int1=3&int2=5&limit=-15&str1=fizz&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "limit must be greater than 0"},
		},
		{
			name:           "empty str1 parameter",
			queryParams:    "int1=3&int2=5&limit=15&str1=&str2=buzz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "str1 cannot be empty"},
		},
		{
			name:           "empty str2 parameter",
			queryParams:    "int1=3&int2=5&limit=15&str1=fizz&str2=",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrorResponse{Error: "str2 cannot be empty"},
		},
		{
			name:           "large limit request",
			queryParams:    "int1=7&int2=11&limit=1000&str1=seven&str2=eleven",
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				t.Helper()

				var resp FizzBuzzResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if len(resp.Result) != 1000 {
					t.Fatalf("expected 1000 results, got %d", len(resp.Result))
				}

				if resp.Result[6] != "seven" {
					t.Errorf("expected position 7 to be seven, got %s", resp.Result[6])
				}

				if resp.Result[10] != "eleven" {
					t.Errorf("expected position 11 to be eleven, got %s", resp.Result[10])
				}

				if resp.Result[76] != "seveneleven" {
					t.Errorf("expected position 77 to be seveneleven, got %s", resp.Result[76])
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/fizzbuzz?"+tc.queryParams, nil)
			rec := httptest.NewRecorder()

			h.FizzBuzz(rec, req)

			res := rec.Result()
			t.Cleanup(func() {
				if err := res.Body.Close(); err != nil {
					t.Fatalf("failed to close response body: %v", err)
				}
			})

			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, res.StatusCode)
			}

			if contentType := res.Header.Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("expected Content-Type application/json, got %s", contentType)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if tc.checkBody != nil {
				tc.checkBody(t, body)
				return
			}

			switch expected := tc.expectedBody.(type) {
			case FizzBuzzResponse:
				assertJSONResponse(t, body, expected)
			case ErrorResponse:
				assertErrorResponse(t, body, expected.Error)
			case nil:
			default:
				t.Fatalf("unsupported expected body type %T", expected)
			}
		})
	}
}

func TestHandler_FizzBuzz_ThroughRouter(t *testing.T) {
	h := NewHandler()
	router := chi.NewRouter()
	router.Get("/fizzbuzz", h.FizzBuzz)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	t.Cleanup(func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	assertJSONResponse(t, body, FizzBuzzResponse{Result: []string{
		"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz",
	}})
}

func assertJSONResponse(t *testing.T, body []byte, expected interface{}) {
	t.Helper()

	value := reflect.New(reflect.TypeOf(expected))
	if err := json.Unmarshal(body, value.Interface()); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	got := value.Elem().Interface()
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected body %+v, got %+v", expected, got)
	}
}

func assertErrorResponse(t *testing.T, body []byte, expectedMessage string) {
	t.Helper()

	var resp ErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if resp.Error != expectedMessage {
		t.Fatalf("expected error message %q, got %q", expectedMessage, resp.Error)
	}
}
