package service

import (
	"careerprogression/internal/apperr"
	"careerprogression/internal/auth"
	"careerprogression/internal/clock"
	"careerprogression/internal/domain"
	"careerprogression/internal/id"
	"careerprogression/internal/repository"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type AuthService struct {
	DB   *store.DB
	Repo *repository.SQLite
	TTL  time.Duration
}

func (s *AuthService) Login(ctx context.Context, email, password string) (domain.Session, domain.User, error) {
	u, e := s.Repo.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if e != nil {
		return domain.Session{}, domain.User{}, apperr.Wrap(apperr.Unauthorized, "invalid credentials", e)
	}
	if !u.Active || !auth.CheckPassword(u.PasswordHash, password) {
		return domain.Session{}, domain.User{}, apperr.New(apperr.Unauthorized, "invalid credentials")
	}
	now := clock.From(ctx).Now()
	sess := domain.Session{ID: id.New("sess"), UserID: u.ID, ExpiresAt: now.Add(s.TTL), CreatedAt: now}
	e = s.DB.Tx(ctx, func(tx *sql.Tx) error { return s.Repo.CreateSession(ctx, tx, sess) })
	return sess, u, e
}
func (s *AuthService) Authenticate(ctx context.Context, sessionID string) (domain.User, error) {
	if sessionID == "" {
		return domain.User{}, apperr.New(apperr.Unauthorized, "session required")
	}
	sess, e := s.Repo.GetSession(ctx, sessionID)
	if e != nil {
		return domain.User{}, apperr.Wrap(apperr.Unauthorized, "session invalid", e)
	}
	now := clock.From(ctx).Now()
	if sess.RevokedAt != nil || !sess.ExpiresAt.After(now) {
		return domain.User{}, apperr.New(apperr.Unauthorized, "session expired or revoked")
	}
	u, e := s.Repo.ByUserID(ctx, sess.UserID)
	if e != nil || !u.Active {
		return domain.User{}, apperr.New(apperr.Unauthorized, "user unavailable")
	}
	return u, nil
}
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	sess, e := s.Repo.GetSession(ctx, sessionID)
	if e != nil {
		return errors.New("session not found")
	}
	return s.DB.Tx(ctx, func(tx *sql.Tx) error { return s.Repo.RevokeSession(ctx, tx, sess.ID) })
}
