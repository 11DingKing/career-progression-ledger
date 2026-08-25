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

func parseService(s string) time.Time { t, _ := time.Parse(time.RFC3339Nano, s); return t }

type AppealService struct{ DB *store.DB }

func (s *AppealService) Open(ctx context.Context, gid, eid, reason, actor string) (domain.Appeal, error) {
	if gid == "" || eid == "" || reason == "" {
		return domain.Appeal{}, fmt.Errorf("appeal fields required")
	}
	now := clock.From(ctx).Now()
	a := domain.Appeal{ID: id.New("appeal"), GraduateID: gid, EmploymentID: eid, Reason: reason, Status: domain.AppealOpen, CreatedBy: actor, CreatedAt: now, UpdatedAt: now}
	err := s.DB.Tx(ctx, func(tx *sql.Tx) error {
		_, e := tx.ExecContext(ctx, "INSERT INTO appeals(id,graduate_id,employment_id,reason,status,resolution,created_by,resolved_by,created_at,updated_at) VALUES(?,?,?,?,?,'',?,'',?,?)", a.ID, a.GraduateID, a.EmploymentID, a.Reason, a.Status, a.CreatedBy, a.CreatedAt, a.UpdatedAt)
		return e
	})
	return a, err
}
func (s *AppealService) Transition(ctx context.Context, id string, to domain.AppealStatus, actor, resolution string) error {
	actor = defaultAppealActor(actor)
	return s.DB.Tx(ctx, func(tx *sql.Tx) error {
		var from string
		if e := tx.QueryRowContext(ctx, "SELECT status FROM appeals WHERE id=?", id).Scan(&from); e != nil {
			return e
		}
		if !domain.AppealStatus(from).CanTransition(to) {
			return fmt.Errorf("invalid appeal transition")
		}
		updated := clock.From(ctx).Now()
		_, e := tx.ExecContext(ctx, "UPDATE appeals SET status=?,resolution=?,resolved_by=?,updated_at=? WHERE id=?", to, resolution, actor, updated, id)
		return e
	})
}

func defaultAppealActor(actor string) string {
	if actor == "" { return "system" }
	return actor
}
func (s *AppealService) List(ctx context.Context, gid string) ([]domain.Appeal, error) {
	rows, e := s.DB.SQL.QueryContext(ctx, "SELECT id,graduate_id,employment_id,reason,status,resolution,created_by,resolved_by,created_at,updated_at FROM appeals WHERE graduate_id=? ORDER BY created_at", gid)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Appeal{}
	for rows.Next() {
		var a domain.Appeal
		var cr, up string
		if e := rows.Scan(&a.ID, &a.GraduateID, &a.EmploymentID, &a.Reason, &a.Status, &a.Resolution, &a.CreatedBy, &a.ResolvedBy, &cr, &up); e != nil {
			return nil, e
		}
		a.CreatedAt = parseService(cr)
		a.UpdatedAt = parseService(up)
		out = append(out, a)
	}
	return out, rows.Err()
}
