package service

import (
	"careerprogression/internal/clock"
	"careerprogression/internal/domain"
	"careerprogression/internal/id"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"fmt"
)

type StatisticsService struct{ DB *store.DB }

func (s *StatisticsService) Freeze(ctx context.Context, major string, year int, actor string) (domain.StatisticSnapshot, error) {
	if major == "" || year < 2000 || actor == "" {
		return domain.StatisticSnapshot{}, fmt.Errorf("invalid freeze request")
	}
	var total, promoted int
	if e := s.DB.SQL.QueryRowContext(ctx, "SELECT COUNT(*),COALESCE(SUM(CASE WHEN status='ended' THEN 1 ELSE 0 END),0) FROM employment_records e JOIN graduates g ON g.id=e.graduate_id WHERE g.major=? AND g.graduation_year=?", major, year).Scan(&total, &promoted); e != nil {
		return domain.StatisticSnapshot{}, e
	}
	now := clock.From(ctx).Now()
	snap := domain.StatisticSnapshot{ID: id.New("stat"), Major: major, GraduationYear: year, Total: total, Promoted: promoted, FrozenAt: now, FrozenBy: actor}
	e := s.DB.Tx(ctx, func(tx *sql.Tx) error {
		_, e := tx.ExecContext(ctx, "INSERT INTO statistic_snapshots(id,major,graduation_year,total,promoted,frozen_at,frozen_by) VALUES(?,?,?,?,?,?,?)", snap.ID, snap.Major, snap.GraduationYear, snap.Total, snap.Promoted, snap.FrozenAt, snap.FrozenBy)
		return e
	})
	return snap, e
}
func (s *StatisticsService) List(ctx context.Context, major string) ([]domain.StatisticSnapshot, error) {
	rows, e := s.DB.SQL.QueryContext(ctx, "SELECT id,major,graduation_year,total,promoted,frozen_at,frozen_by FROM statistic_snapshots WHERE (?='' OR major=?) ORDER BY frozen_at DESC", major, major)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.StatisticSnapshot{}
	for rows.Next() {
		var x domain.StatisticSnapshot
		var t string
		if e := rows.Scan(&x.ID, &x.Major, &x.GraduationYear, &x.Total, &x.Promoted, &t, &x.FrozenBy); e != nil {
			return nil, e
		}
		x.FrozenAt = parseService(t)
		out = append(out, x)
	}
	return out, rows.Err()
}
