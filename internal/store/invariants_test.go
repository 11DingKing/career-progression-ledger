package store

import (
	"careerprogression/internal/migrations"
	"context"
	"database/sql"
	"os"
	"testing"
)

func dbFixture(t *testing.T) (context.Context, *DB) {
	f, e := os.CreateTemp("", "db-*.sqlite")
	if e != nil {
		t.Fatal(e)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	ctx := context.Background()
	db, e := Open(ctx, f.Name())
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	if e = migrations.Apply(ctx, db.SQL); e != nil {
		t.Fatal(e)
	}
	return ctx, db
}
func TestForeignKeysRejectOrphans(t *testing.T) {
	ctx, db := dbFixture(t)
	checks := []string{"INSERT INTO sessions(id,user_id,expires_at,created_at) VALUES('s','missing','x','x')", "INSERT INTO graduates(id,user_id,student_no,name,major,graduation_year,contact,created_at,updated_at) VALUES('g','missing','s','n','m',2024,'c','x','x')", "INSERT INTO employment_records(id,graduate_id,employer_id,title,salary_min,salary_max,start_date,status,version,source,created_at,updated_at) VALUES('e','missing','missing','t',1,2,'x','draft',1,'s','x','x')"}
	for _, q := range checks {
		if _, e := db.SQL.ExecContext(ctx, q); e == nil {
			t.Fatalf("orphan accepted: %s", q)
		}
	}
}
func TestUniqueConstraints(t *testing.T) {
	ctx, db := dbFixture(t)
	_, e := db.SQL.ExecContext(ctx, "INSERT INTO users(id,email,name,role,password_hash,active,created_at) VALUES('u','u@e','n','graduate','h',1,'x')")
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.ExecContext(ctx, "INSERT INTO users(id,email,name,role,password_hash,active,created_at) VALUES('u2','u@e','n','graduate','h',1,'x')")
	if e == nil {
		t.Fatal("duplicate email")
	}
	_, e = db.SQL.ExecContext(ctx, "INSERT INTO employers(id,name,registration_no,contact,created_at,updated_at) VALUES('e','E','r','c','x','x')")
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.ExecContext(ctx, "INSERT INTO employers(id,name,registration_no,contact,created_at,updated_at) VALUES('e2','E2','r','c','x','x')")
	if e == nil {
		t.Fatal("duplicate registration")
	}
}
func TestTransactionRollback(t *testing.T) {
	ctx, db := dbFixture(t)
	e := db.Tx(ctx, func(tx *sql.Tx) error {
		if _, e := tx.ExecContext(ctx, "INSERT INTO users(id,email,name,role,password_hash,active,created_at) VALUES('u','u@e','n','graduate','h',1,'x')"); e != nil {
			return e
		}
		return sql.ErrTxDone
	})
	if e == nil {
		t.Fatal("expected error")
	}
	var n int
	if e = db.SQL.QueryRow("SELECT COUNT(*) FROM users").Scan(&n); e != nil || n != 0 {
		t.Fatalf("rollback count=%d err=%v", n, e)
	}
}
func TestBusyTimeoutConfigured(t *testing.T) {
	_, db := dbFixture(t)
	var v string
	if e := db.SQL.QueryRow("PRAGMA busy_timeout").Scan(&v); e != nil {
		t.Fatal(e)
	}
}
func TestCloseIdempotent(t *testing.T) {
	_, db := dbFixture(t)
	if e := db.Close(); e != nil {
		t.Fatal(e)
	}
}
func TestSchemaIndexes(t *testing.T) {
	_, db := dbFixture(t)
	rows, e := db.SQL.Query("SELECT name FROM sqlite_master WHERE type='index'")
	if e != nil {
		t.Fatal(e)
	}
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var n string
		_ = rows.Scan(&n)
		seen[n] = true
	}
	for _, n := range []string{"idx_sessions_user", "idx_employment_graduate", "idx_jobs_due"} {
		if !seen[n] {
			t.Fatalf("missing index %s", n)
		}
	}
}
func TestContextCancellation(t *testing.T) {
	ctx, db := dbFixture(t)
	cctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, e := db.SQL.ExecContext(cctx, "SELECT 1"); e == nil {
		t.Fatal("cancelled query succeeded")
	}
}
func TestMigrationVersion(t *testing.T) {
	_, db := dbFixture(t)
	var v int
	if e := db.SQL.QueryRow("SELECT version FROM schema_migrations").Scan(&v); e != nil || v != 1 {
		t.Fatalf("version %d %v", v, e)
	}
}
func TestTableColumns(t *testing.T) {
	_, db := dbFixture(t)
	tables := map[string]int{"users": 7, "sessions": 5, "graduates": 9, "employers": 7, "employment_records": 13, "career_events": 8, "skills": 6, "training_records": 7, "consents": 7, "appeals": 10, "audit_events": 9, "idempotency_keys": 5, "followup_jobs": 10, "statistic_snapshots": 7}
	for table, min := range tables {
		rows, e := db.SQL.Query("PRAGMA table_info(" + table + ")")
		if e != nil {
			t.Fatal(e)
		}
		count := 0
		for rows.Next() {
			count++
		}
		rows.Close()
		if count < min {
			t.Fatalf("%s columns %d", table, count)
		}
	}
}
