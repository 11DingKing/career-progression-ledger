package audit

import (
	"careerprogression/internal/clock"
	"careerprogression/internal/domain"
	"careerprogression/internal/id"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"time"
)

type Logger struct{ DB *store.DB }

func (l *Logger) Record(ctx context.Context, tx *sql.Tx, a domain.AuditEvent) error {
	if a.ID == "" {
		a.ID = id.New("audit")
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = clock.From(ctx).Now()
	}
	_, e := tx.ExecContext(ctx, "INSERT INTO audit_events(id,actor_id,object_type,object_id,action,result,request_id,details,created_at) VALUES(?,?,?,?,?,?,?,?,?)", a.ID, a.ActorID, a.ObjectType, a.ObjectID, a.Action, a.Result, a.RequestID, a.Details, a.CreatedAt.UTC().Format(time.RFC3339Nano))
	return e
}
