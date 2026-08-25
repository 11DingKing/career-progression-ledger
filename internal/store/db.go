package store

import (
	"context"
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

type DB struct {
	SQL  *sql.DB
	path string
}

func Open(ctx context.Context, path string) (*DB, error) {
	db, e := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if e != nil {
		return nil, e
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	if e = db.PingContext(ctx); e != nil {
		db.Close()
		return nil, e
	}
	return &DB{SQL: db, path: path}, nil
}
func (d *DB) Close() error { return d.SQL.Close() }
func (d *DB) Tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, e := d.SQL.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	if e = fn(tx); e != nil {
		_ = tx.Rollback()
		return e
	}
	if e = tx.Commit(); e != nil {
		return fmt.Errorf("commit: %w", e)
	}
	return nil
}
