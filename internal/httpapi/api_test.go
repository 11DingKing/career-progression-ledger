package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestJSONEnvelope(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("X-Request-ID", "r")
	JSON(w, 200, map[string]any{"ok": true})
	if w.Code != 200 || !strings.Contains(w.Body.String(), "request_id") {
		t.Fatal(w.Body.String())
	}
}
func TestJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("X-Request-ID", "r")
	Error(w, 409, "conflict", "already changed")
	if w.Code != 409 || !strings.Contains(w.Body.String(), "already changed") {
		t.Fatal(w.Body.String())
	}
}
func TestDeadline(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	next, cancel := Deadline(r, time.Second)
	defer cancel()
	if next.Context() == r.Context() {
		t.Fatal("context not replaced")
	}
}
