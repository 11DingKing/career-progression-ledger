package handler

import (
	"careerprogression/internal/domain"
	"context"
	"net/http"
)

type key string

const userKey key = "user"

func withUser(r *http.Request, u domain.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userKey, u))
}
func current(r *http.Request) (domain.User, bool) {
	u, ok := r.Context().Value(userKey).(domain.User)
	return u, ok
}
