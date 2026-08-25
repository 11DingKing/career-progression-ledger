package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Envelope struct {
	Data      any      `json:"data,omitempty"`
	Error     *Problem `json:"error,omitempty"`
	RequestID string   `json:"request_id"`
}
type Problem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: data, RequestID: w.Header().Get("X-Request-ID")})
}
func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Error: &Problem{Code: code, Message: message}, RequestID: w.Header().Get("X-Request-ID")})
}
func Deadline(r *http.Request, d time.Duration) (*http.Request, func()) {
	return contextWithTimeout(r, d)
}
func contextWithTimeout(r *http.Request, d time.Duration) (*http.Request, func()) {
	ctx, cancel := context.WithTimeout(r.Context(), d)
	return r.WithContext(ctx), cancel
}
