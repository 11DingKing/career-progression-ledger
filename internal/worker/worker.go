package worker

import (
	"careerprogression/internal/clock"
	"careerprogression/internal/id"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"log/slog"
	"time"
)

type Worker struct {
	DB    *store.DB
	Every time.Duration
	Log   *slog.Logger
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.Every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			w.tick(ctx, now)
		}
	}
}
func (w *Worker) tick(ctx context.Context, now time.Time) {
	_ = w.DB.Tx(ctx, func(tx *sql.Tx) error {
		rows, e := tx.QueryContext(ctx, "SELECT id,attempts FROM followup_jobs WHERE status='pending' AND due_at<=? ORDER BY due_at LIMIT 10", now.UTC().Format(time.RFC3339Nano))
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var job string
			var attempts int
			if e := rows.Scan(&job, &attempts); e != nil {
				return e
			}
			if _, e = tx.ExecContext(ctx, "UPDATE followup_jobs SET status='running',attempts=attempts+1,locked_at=?,updated_at=? WHERE id=? AND status='pending'", now, now, job); e != nil {
				return e
			}
		}
		return rows.Err()
	})
}
func Enqueue(ctx context.Context, db *store.DB, graduate, kind string, due time.Time) error {
	return db.Tx(ctx, func(tx *sql.Tx) error {
		_, e := tx.ExecContext(ctx, "INSERT INTO followup_jobs(id,graduate_id,kind,due_at,attempts,status,last_error,created_at,updated_at) VALUES(?,?,?,?,0,'pending','',?,?)", id.New("job"), graduate, kind, due, due, due)
		return e
	})
}
func Retry(ctx context.Context, db *store.DB, id string, err error) error {
	return db.Tx(ctx, func(tx *sql.Tx) error {
		_, e := tx.ExecContext(ctx, "UPDATE followup_jobs SET attempts=attempts+1,status=CASE WHEN attempts+1>=3 THEN 'failed' ELSE 'pending' END,last_error=?,updated_at=? WHERE id=?", err.Error(), clock.From(ctx).Now(), id)
		return e
	})
}
