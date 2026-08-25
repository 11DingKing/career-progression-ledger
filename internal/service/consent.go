package service

import (
	"careerprogression/internal/clock"
	"careerprogression/internal/domain"
	"careerprogression/internal/id"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ConsentService struct{ DB *store.DB }

func (s *ConsentService) Grant(ctx context.Context, gid, scope string) error {
	return s.set(ctx, gid, scope, true)
}
func (s *ConsentService) Revoke(ctx context.Context, gid, scope string) error {
	return s.set(ctx, gid, scope, false)
}
func (s *ConsentService) set(ctx context.Context, gid, scope string, grant bool) error {
	if gid == "" || scope == "" {
		return fmt.Errorf("consent fields required")
	}
	now := clock.From(ctx).Now()
	return s.DB.Tx(ctx, func(tx *sql.Tx) error {
		var idv string
		var version int
		err := tx.QueryRowContext(ctx, "SELECT id,version FROM consents WHERE graduate_id=? AND scope=?", gid, scope).Scan(&idv, &version)
		if err == sql.ErrNoRows {
			idv = id.New("consent")
			version = 1
			_, err = tx.ExecContext(ctx, "INSERT INTO consents(id,graduate_id,scope,granted,granted_at,version) VALUES(?,?,?,?,?,?)", idv, gid, scope, consentInt(grant), now, version)
			return err
		}
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "UPDATE consents SET granted=?,granted_at=?,revoked_at=?,version=version+1 WHERE id=? AND version=?", consentInt(grant), now, consentTime(grant, now), idv, version)
		return err
	})
}
func (s *ConsentService) Allowed(ctx context.Context, gid, scope string) (bool, error) {
	var granted int
	err := s.DB.SQL.QueryRowContext(ctx, "SELECT granted FROM consents WHERE graduate_id=? AND scope=?", gid, scope).Scan(&granted)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return granted == 1, err
}
func (s *ConsentService) Snapshot(ctx context.Context, gid string) ([]domain.Consent, error) {
	rows, e := s.DB.SQL.QueryContext(ctx, "SELECT id,graduate_id,scope,granted,granted_at,revoked_at,version FROM consents WHERE graduate_id=? ORDER BY scope", gid)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Consent{}
	for rows.Next() {
		var c domain.Consent
		var g, gr string
		var granted int
		if e := rows.Scan(&c.ID, &c.GraduateID, &c.Scope, &granted, &g, &gr, &c.Version); e != nil {
			return nil, e
		}
		c.Granted = granted == 1
		c.GrantedAt, _ = time.Parse(time.RFC3339Nano, g)
		if gr != "" {
			t, _ := time.Parse(time.RFC3339Nano, gr)
			c.RevokedAt = &t
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func consentInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func consentTime(grant bool, now time.Time) any {
	if grant {
		return nil
	}
	return now
}
