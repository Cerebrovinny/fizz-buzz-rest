package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Cerebrovinny/fizz-buzz-rest/internal/fizzbuzz"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type FizzBuzzResponse struct {
	Result []string `json:"result"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type fizzBuzzParams struct {
	int1  int
	int2  int
	limit int
	str1  string
	str2  string
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(payload); err != nil {
		log.Printf("json response write error: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

func (h *Handler) FizzBuzz(w http.ResponseWriter, r *http.Request) {
	params, err := parseFizzBuzzParams(r.URL.Query())
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result := fizzbuzz.Generate(params.int1, params.int2, params.limit, params.str1, params.str2)

	respondJSON(w, http.StatusOK, FizzBuzzResponse{Result: result})
}

func parseFizzBuzzParams(values url.Values) (fizzBuzzParams, error) {
	const missingParamsMessage = "missing required parameters: int1, int2, limit, str1, str2"

	requiredParams := []string{"int1", "int2", "limit", "str1", "str2"}
	for _, param := range requiredParams {
		if _, exists := values[param]; !exists || len(values[param]) == 0 {
			return fizzBuzzParams{}, errors.New(missingParamsMessage)
		}
	}

	str1 := values.Get("str1")
	if str1 == "" {
		return fizzBuzzParams{}, fmt.Errorf("str1 cannot be empty")
	}

	str2 := values.Get("str2")
	if str2 == "" {
		return fizzBuzzParams{}, fmt.Errorf("str2 cannot be empty")
	}

	int1, err := parsePositiveInt(values.Get("int1"), "int1")
	if err != nil {
		return fizzBuzzParams{}, err
	}

	int2, err := parsePositiveInt(values.Get("int2"), "int2")
	if err != nil {
		return fizzBuzzParams{}, err
	}

	limit, err := parsePositiveInt(values.Get("limit"), "limit")
	if err != nil {
		return fizzBuzzParams{}, err
	}

	return fizzBuzzParams{
		int1:  int1,
		int2:  int2,
		limit: limit,
		str1:  str1,
		str2:  str2,
	}, nil
}

func parsePositiveInt(value string, name string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", name)
	}

	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", name)
	}

	return parsed, nil
}
