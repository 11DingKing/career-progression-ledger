package handler

import (
	"careerprogression/internal/apperr"
	"careerprogression/internal/auth"
	"careerprogression/internal/domain"
	"careerprogression/internal/pagination"
	"careerprogression/internal/service"
	"careerprogression/internal/store"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	Auth   *service.AuthService
	Career *service.CareerService
	DB     *store.DB
	Log    *slog.Logger
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.health)
	mux.HandleFunc("/readyz", h.ready)
	mux.HandleFunc("/v1/auth/login", h.login)
	mux.HandleFunc("/v1/auth/logout", h.logout)
	mux.Handle("/v1/graduates", h.authn(http.HandlerFunc(h.graduates)))
	mux.Handle("/v1/employment", h.authn(http.HandlerFunc(h.employment)))
	mux.Handle("/v1/events", h.authn(http.HandlerFunc(h.events)))
	return requestID(recoverer(mux))
}
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	write(w, 200, map[string]any{"status": "ok"})
}
func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if e := h.DB.SQL.PingContext(r.Context()); e != nil {
		writeErr(w, apperr.Wrap(apperr.Unavailable, "database unavailable", e))
		return
	}
	write(w, 200, map[string]any{"status": "ready"})
}
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if !decode(r, &in) {
		writeErr(w, apperr.New(apperr.Invalid, "invalid json"))
		return
	}
	s, u, e := h.Auth.Login(r.Context(), in.Email, in.Password)
	if e != nil {
		writeErr(w, e)
		return
	}
	write(w, 200, map[string]any{"session_id": s.ID, "expires_at": s.ExpiresAt, "user": u})
}
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if e := h.Auth.Logout(r.Context(), token(r)); e != nil {
		writeErr(w, apperr.Wrap(apperr.Unauthorized, "logout failed", e))
		return
	}
	write(w, 204, nil)
}
func (h *Handler) authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, e := h.Auth.Authenticate(r.Context(), token(r))
		if e != nil {
			writeErr(w, e)
			return
		}
		next.ServeHTTP(w, withUser(r, u))
	})
}
func (h *Handler) graduates(w http.ResponseWriter, r *http.Request) {
	u, _ := current(r)
	if u.Role != domain.RoleCounselor && u.Role != domain.RoleMajorAdmin && u.Role != domain.RoleGraduate {
		writeErr(w, apperr.New(apperr.Forbidden, "role cannot list graduates"))
		return
	}
	p := pagination.Parse(atoi(r.URL.Query().Get("limit")), atoi(r.URL.Query().Get("offset")))
	gs, e := h.Career.ListGraduates(r.Context(), r.URL.Query().Get("major"), p.Limit, p.Offset)
	if e != nil {
		writeErr(w, e)
		return
	}
	write(w, 200, map[string]any{"items": gs, "limit": p.Limit, "offset": p.Offset})
}
func (h *Handler) employment(w http.ResponseWriter, r *http.Request) {
	u, _ := current(r)
	switch r.Method {
	case http.MethodPost:
		var in domain.EmploymentRecord
		if !decode(r, &in) {
			writeErr(w, apperr.New(apperr.Invalid, "invalid json"))
			return
		}
		if u.Role == domain.RoleGraduate {
			g, e := h.Career.Repo.Graduate(r.Context(), in.GraduateID)
			if e != nil || g.UserID != u.ID {
				writeErr(w, apperr.New(apperr.Forbidden, "graduate ownership required"))
				return
			}
		}
		if e := h.Career.AddEmployment(r.Context(), in); e != nil {
			writeErr(w, e)
			return
		}
		write(w, 201, in)
	case http.MethodPatch:
		var in struct {
			ID      string                  `json:"id"`
			Status  domain.EmploymentStatus `json:"status"`
			Version int                     `json:"version"`
		}
		if !decode(r, &in) {
			writeErr(w, apperr.New(apperr.Invalid, "invalid json"))
			return
		}
		if e := h.Career.TransitionEmployment(r.Context(), in.ID, in.Status, in.Version); e != nil {
			writeErr(w, e)
			return
		}
		write(w, 204, nil)
	default:
		writeErr(w, apperr.New(apperr.Invalid, "method not supported"))
	}
}
func (h *Handler) events(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		id := r.URL.Query().Get("employment_id")
		es, e := h.Career.Events(r.Context(), id)
		if e != nil {
			writeErr(w, e)
			return
		}
		write(w, 200, map[string]any{"items": es})
		return
	}
	var in domain.CareerEvent
	if !decode(r, &in) {
		writeErr(w, apperr.New(apperr.Invalid, "invalid json"))
		return
	}
	if e := h.Career.AddEvent(r.Context(), in); e != nil {
		writeErr(w, e)
		return
	}
	write(w, 201, in)
}
func token(r *http.Request) string {
	v := r.Header.Get("Authorization")
	return strings.TrimPrefix(v, "Bearer ")
}
func decode(r *http.Request, v any) bool {
	return json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(v) == nil
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
func writeErr(w http.ResponseWriter, e error) {
	status := http.StatusInternalServerError
	var ae *apperr.Error
	if errors.As(e, &ae) {
		switch ae.Code {
		case apperr.Invalid:
			status = 400
		case apperr.Unauthorized:
			status = 401
		case apperr.Forbidden:
			status = 403
		case apperr.NotFound:
			status = 404
		case apperr.Conflict:
			status = 409
		case apperr.Unavailable:
			status = 503
		}
		write(w, status, map[string]any{"error": map[string]any{"code": ae.Code, "message": ae.Message}})
		return
	}
	write(w, status, map[string]any{"error": map[string]string{"code": "internal", "message": "internal error"}})
}
func atoi(v string) int { n, _ := strconv.Atoi(v); return n }
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = auth.HashPassword(r.RemoteAddr)[:16]
		}
		w.Header().Set("X-Request-ID", rid)
		next.ServeHTTP(w, r)
	})
}
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				writeErr(w, apperr.New(apperr.Internal, "request failed"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
