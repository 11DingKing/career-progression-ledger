package migrations

import (
	"context"
	"database/sql"
	"fmt"
)

const schema = `CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS users(id TEXT PRIMARY KEY,email TEXT NOT NULL UNIQUE,name TEXT NOT NULL,role TEXT NOT NULL,password_hash TEXT NOT NULL,active INTEGER NOT NULL DEFAULT 1,created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS sessions(id TEXT PRIMARY KEY,user_id TEXT NOT NULL REFERENCES users(id),expires_at TEXT NOT NULL,revoked_at TEXT,created_at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE TABLE IF NOT EXISTS graduates(id TEXT PRIMARY KEY,user_id TEXT NOT NULL UNIQUE REFERENCES users(id),student_no TEXT NOT NULL UNIQUE,name TEXT NOT NULL,major TEXT NOT NULL,graduation_year INTEGER NOT NULL,contact TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS employers(id TEXT PRIMARY KEY,name TEXT NOT NULL,registration_no TEXT NOT NULL UNIQUE,contact TEXT NOT NULL,verified INTEGER NOT NULL DEFAULT 0,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS employment_records(id TEXT PRIMARY KEY,graduate_id TEXT NOT NULL REFERENCES graduates(id),employer_id TEXT NOT NULL REFERENCES employers(id),title TEXT NOT NULL,salary_min INTEGER NOT NULL,salary_max INTEGER NOT NULL,start_date TEXT NOT NULL,end_date TEXT,status TEXT NOT NULL,version INTEGER NOT NULL DEFAULT 1,source TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_employment_graduate ON employment_records(graduate_id,start_date);
CREATE TABLE IF NOT EXISTS career_events(id TEXT PRIMARY KEY,employment_id TEXT NOT NULL REFERENCES employment_records(id),kind TEXT NOT NULL,occurred_at TEXT NOT NULL,summary TEXT NOT NULL,evidence_url TEXT NOT NULL,created_by TEXT NOT NULL REFERENCES users(id),created_at TEXT NOT NULL,UNIQUE(employment_id,kind,occurred_at));
CREATE TABLE IF NOT EXISTS skills(id TEXT PRIMARY KEY,graduate_id TEXT NOT NULL REFERENCES graduates(id),name TEXT NOT NULL,level INTEGER NOT NULL,verified_at TEXT,created_at TEXT NOT NULL,UNIQUE(graduate_id,name));
CREATE TABLE IF NOT EXISTS training_records(id TEXT PRIMARY KEY,graduate_id TEXT NOT NULL REFERENCES graduates(id),provider TEXT NOT NULL,name TEXT NOT NULL,completed_at TEXT NOT NULL,certificate_no TEXT NOT NULL UNIQUE,created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS consents(id TEXT PRIMARY KEY,graduate_id TEXT NOT NULL REFERENCES graduates(id),scope TEXT NOT NULL,granted INTEGER NOT NULL,granted_at TEXT NOT NULL,revoked_at TEXT,version INTEGER NOT NULL DEFAULT 1,UNIQUE(graduate_id,scope));
CREATE TABLE IF NOT EXISTS appeals(id TEXT PRIMARY KEY,graduate_id TEXT NOT NULL REFERENCES graduates(id),employment_id TEXT NOT NULL REFERENCES employment_records(id),reason TEXT NOT NULL,status TEXT NOT NULL,resolution TEXT NOT NULL,created_by TEXT NOT NULL REFERENCES users(id),resolved_by TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS audit_events(id TEXT PRIMARY KEY,actor_id TEXT NOT NULL,object_type TEXT NOT NULL,object_id TEXT NOT NULL,action TEXT NOT NULL,result TEXT NOT NULL,request_id TEXT NOT NULL,details TEXT NOT NULL,created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS idempotency_keys(key TEXT PRIMARY KEY,actor_id TEXT NOT NULL,operation TEXT NOT NULL,response TEXT NOT NULL,created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS followup_jobs(id TEXT PRIMARY KEY,graduate_id TEXT NOT NULL REFERENCES graduates(id),kind TEXT NOT NULL,due_at TEXT NOT NULL,attempts INTEGER NOT NULL DEFAULT 0,status TEXT NOT NULL,last_error TEXT NOT NULL,locked_at TEXT,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_jobs_due ON followup_jobs(status,due_at);
CREATE TABLE IF NOT EXISTS statistic_snapshots(id TEXT PRIMARY KEY,major TEXT NOT NULL,graduation_year INTEGER NOT NULL,total INTEGER NOT NULL,promoted INTEGER NOT NULL,frozen_at TEXT NOT NULL,frozen_by TEXT NOT NULL);
`

func Apply(ctx context.Context, db *sql.DB) error {
	if _, e := db.ExecContext(ctx, schema); e != nil {
		return fmt.Errorf("schema: %w", e)
	}
	var n int
	if e := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&n); e != nil {
		return e
	}
	if n == 0 {
		_, e := db.ExecContext(ctx, "INSERT INTO schema_migrations(version,applied_at) VALUES(1,datetime('now'))")
		return e
	}
	return nil
}
