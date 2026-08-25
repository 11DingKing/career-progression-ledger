package service

import (
	"careerprogression/internal/clock"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type IdempotencyService struct{ DB *store.DB }

func (s *IdempotencyService) Execute(ctx context.Context, key, actor, operation string, fn func() any) (any, error) {
	if key == "" || actor == "" || operation == "" {
		return nil, fmt.Errorf("idempotency fields required")
	}
	var raw string
	err := s.DB.SQL.QueryRowContext(ctx, "SELECT response FROM idempotency_keys WHERE key=? AND actor_id=? AND operation=?", key, actor, operation).Scan(&raw)
	if err == nil {
		var v any
		if e := json.Unmarshal([]byte(raw), &v); e != nil {
			return nil, e
		}
		return v, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	value := fn()
	encoded, e := json.Marshal(value)
	if e != nil {
		return nil, e
	}
	e = s.DB.Tx(ctx, func(tx *sql.Tx) error {
		_, e := tx.ExecContext(ctx, "INSERT INTO idempotency_keys(key,actor_id,operation,response,created_at) VALUES(?,?,?,?,?)", key, actor, operation, string(encoded), clock.From(ctx).Now())
		return e
	})
	return value, e
}
