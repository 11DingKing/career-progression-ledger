package audit

import (
	"careerprogression/internal/domain"
	"context"
	"database/sql"
	"time"
)

func Query(ctx context.Context, db *sql.DB, objectID string, limit int) ([]domain.AuditEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	rows, e := db.QueryContext(ctx, "SELECT id,actor_id,object_type,object_id,action,result,request_id,details,created_at FROM audit_events WHERE object_id=? ORDER BY created_at DESC LIMIT ?", objectID, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.AuditEvent{}
	for rows.Next() {
		var a domain.AuditEvent
		var created string
		if e := rows.Scan(&a.ID, &a.ActorID, &a.ObjectType, &a.ObjectID, &a.Action, &a.Result, &a.RequestID, &a.Details, &created); e != nil {
			return nil, e
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, a)
	}
	return out, rows.Err()
}
