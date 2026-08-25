package handler

import (
	"bytes"
	"careerprogression/internal/auth"
	"careerprogression/internal/domain"
	"careerprogression/internal/migrations"
	"careerprogression/internal/repository"
	"careerprogression/internal/service"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func httpFixture(t *testing.T) (http.Handler, *store.DB) {
	f, e := os.CreateTemp("", "http-*.db")
	if e != nil {
		t.Fatal(e)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	ctx := context.Background()
	db, e := store.Open(ctx, f.Name())
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	if e = migrations.Apply(ctx, db.SQL); e != nil {
		t.Fatal(e)
	}
	r := repository.New(db.SQL)
	now := time.Now().UTC()
	if e := db.Tx(ctx, func(tx *sql.Tx) error {
		u := domain.User{ID: "u", Email: "u@e", Name: "U", Role: domain.RoleCounselor, PasswordHash: auth.HashPassword("pw"), Active: true, CreatedAt: now}
		return r.CreateUser(ctx, tx, u)
	}); e != nil {
		t.Fatal(e)
	}
	h := &Handler{Auth: &service.AuthService{DB: db, Repo: r, TTL: time.Hour}, Career: &service.CareerService{DB: db, Repo: r}, DB: db}
	return h.Routes(), db
}
func TestHealthReady(t *testing.T) {
	h, db := httpFixture(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Fatalf("%s %d", path, rec.Code)
		}
	}
	db.Close()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != 503 {
		t.Fatalf("ready after close %d", rec.Code)
	}
}
func TestLoginAndProtectedRoute(t *testing.T) {
	h, _ := httpFixture(t)
	body, _ := json.Marshal(map[string]string{"email": "u@e", "password": "pw"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body)))
	if rec.Code != 200 {
		t.Fatalf("login %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		SessionID string `json:"session_id"`
	}
	if json.Unmarshal(rec.Body.Bytes(), &out) != nil || out.SessionID == "" {
		t.Fatal("session missing")
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/graduates", nil)
	req.Header.Set("Authorization", "Bearer "+out.SessionID)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("protected %d", rec.Code)
	}
}
func TestUnauthorizedAndInvalidJSON(t *testing.T) {
	h, _ := httpFixture(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/graduates", nil))
	if rec.Code != 401 {
		t.Fatalf("unauth %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString("{")))
	if rec.Code != 400 {
		t.Fatalf("invalid %d", rec.Code)
	}
}
func TestRequestIDAndLogout(t *testing.T) {
	h, _ := httpFixture(t)
	body := bytes.NewBufferString(`{"email":"u@e","password":"pw"}`)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/auth/login", body))
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	sid := out["session_id"].(string)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+sid)
	req.Header.Set("X-Request-ID", "req-123")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 204 || rec.Header().Get("X-Request-ID") != "req-123" {
		t.Fatalf("logout %d", rec.Code)
	}
}
func TestEmploymentMethodAndEventRoute(t *testing.T) {
	h, _ := httpFixture(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/v1/employment", nil))
	if rec.Code != 401 {
		t.Fatalf("method %d", rec.Code)
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/events?employment_id=x", nil)
	req.Header.Set("Authorization", "Bearer missing")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("event auth %d", rec.Code)
	}
}
