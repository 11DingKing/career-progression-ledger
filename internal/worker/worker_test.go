package worker

import (
	"careerprogression/internal/migrations"
	"careerprogression/internal/store"
	"context"
	"os"
	"testing"
	"time"
)

func workerFixture(t *testing.T) (context.Context, *store.DB) {
	f, e := os.CreateTemp("", "worker-*.db")
	if e != nil {
		t.Fatal(e)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	ctx := context.Background()
	db, e := store.Open(ctx, f.Name())
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	if e = migrations.Apply(ctx, db.SQL); e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.Exec("INSERT INTO users(id,email,name,role,password_hash,active,created_at) VALUES('u','u@e','U','graduate','h',1,datetime('now'))")
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.Exec("INSERT INTO graduates(id,user_id,student_no,name,major,graduation_year,contact,created_at,updated_at) VALUES('g','u','s','N','m',2024,'c',datetime('now'),datetime('now'))")
	if e != nil {
		t.Fatal(e)
	}
	return ctx, db
}
func TestEnqueueAndClaim(t *testing.T) {
	ctx, db := workerFixture(t)
	if e := Enqueue(ctx, db, "g", "survey", time.Now().Add(-time.Minute)); e != nil {
		t.Fatal(e)
	}
	w := &Worker{DB: db, Every: time.Millisecond}
	w.tick(ctx, time.Now())
	var status string
	if e := db.SQL.QueryRow("SELECT status FROM followup_jobs LIMIT 1").Scan(&status); e != nil || status != "running" {
		t.Fatalf("%s %v", status, e)
	}
}
func TestRetryEventuallyFails(t *testing.T) {
	ctx, db := workerFixture(t)
	if e := Enqueue(ctx, db, "g", "survey", time.Now()); e != nil {
		t.Fatal(e)
	}
	var id string
	if e := db.SQL.QueryRow("SELECT id FROM followup_jobs LIMIT 1").Scan(&id); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 3; i++ {
		if e := Retry(ctx, db, id, errTest{}); e != nil {
			t.Fatal(e)
		}
	}
	var status string
	if e := db.SQL.QueryRow("SELECT status FROM followup_jobs WHERE id=?", id).Scan(&status); e != nil || status != "failed" {
		t.Fatalf("%s %v", status, e)
	}
}
func TestWorkerStopsOnCancel(t *testing.T) {
	ctx, db := workerFixture(t)
	cctx, cancel := context.WithCancel(ctx)
	w := &Worker{DB: db, Every: time.Millisecond}
	done := make(chan struct{})
	go func() { w.Run(cctx); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
}

type errTest struct{}

func (errTest) Error() string { return "temporary" }
