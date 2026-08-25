package migrations

import (
	"careerprogression/internal/store"
	"context"
	"os"
	"testing"
)

func TestApplyIsIdempotent(t *testing.T) {
	f, e := os.CreateTemp("", "mig-*.db")
	if e != nil {
		t.Fatal(e)
	}
	f.Close()
	defer os.Remove(f.Name())
	ctx := context.Background()
	db, e := store.Open(ctx, f.Name())
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = Apply(ctx, db.SQL); e != nil {
		t.Fatal(e)
	}
	if e = Apply(ctx, db.SQL); e != nil {
		t.Fatal(e)
	}
	var n int
	if e = db.SQL.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&n); e != nil || n != 1 {
		t.Fatalf("%d %v", n, e)
	}
	tables := []string{"users", "sessions", "graduates", "employers", "employment_records", "career_events", "skills", "training_records", "consents", "appeals", "audit_events", "idempotency_keys", "followup_jobs", "statistic_snapshots"}
	for _, name := range tables {
		var count int
		if e = db.SQL.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&count); e != nil || count != 1 {
			t.Fatalf("table %s %v", name, e)
		}
	}
}
