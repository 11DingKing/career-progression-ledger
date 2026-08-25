package service

import (
	"careerprogression/internal/store"
	"careerprogression/internal/worker"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type FollowupService struct{ DB *store.DB }

func (s *FollowupService) Schedule(ctx context.Context, gid string, due time.Time) error {
	if gid == "" || due.Before(time.Now().Add(-24*time.Hour)) {
		return fmt.Errorf("invalid followup")
	}
	return worker.Enqueue(ctx, s.DB, gid, "career_check", due)
}
func (s *FollowupService) Cancel(ctx context.Context, id string) error {
	return s.DB.Tx(ctx, func(tx *sql.Tx) error {
		_, e := tx.ExecContext(ctx, "UPDATE followup_jobs SET status='cancelled',updated_at=? WHERE id=? AND status IN ('pending','running')", time.Now().UTC(), id)
		return e
	})
}
