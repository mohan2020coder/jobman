package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// APIError represents an error that maps to a consistent JSON error body.
type APIError struct {
	Status  int
	Code    string
	Message string
	Err     error // optional underlying cause, logged server-side
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func NewAPIError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("httpapi: failed to encode response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}

func WriteAPIError(w http.ResponseWriter, err error) {
	apiErr, ok := err.(*APIError)
	if !ok {
		log.Printf("httpapi: unhandled error: %v", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong.")
		return
	}
	if apiErr.Err != nil {
		log.Printf("httpapi: %s (%s): %v", apiErr.Code, apiErr.Message, apiErr.Err)
	}
	if apiErr.Code == "" {
		apiErr.Code = "ERROR"
	}
	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(apiErr.Status)
	}
	WriteError(w, apiErr.Status, apiErr.Code, apiErr.Message)
}

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

func ParsePageLimit(r *http.Request, defaultLimit int) (int, int) {
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	limit := parsePositiveInt(r.URL.Query().Get("limit"), defaultLimit)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = defaultLimit
	}
	return page, limit
}

func parsePositiveInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return def
	}
	return n
}

func OffsetFrom(page, limit int) int {
	return (page - 1) * limit
}