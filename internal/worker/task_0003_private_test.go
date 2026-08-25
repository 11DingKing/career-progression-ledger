package worker

import (
	"context"
	"testing"
	"time"
)

func TestCancelledRetryCannotMutateFollowup(t *testing.T) {
	ctx, db := workerFixture(t)
	if err := Enqueue(ctx, db, "g", "survey", time.Now()); err != nil {
		t.Fatal(err)
	}
	var id string
	if err := db.SQL.QueryRow("SELECT id FROM followup_jobs LIMIT 1").Scan(&id); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := Retry(cancelled, db, id, errTest{}); err == nil {
		t.Fatal("cancelled retry unexpectedly succeeded")
	}
	var attempts int
	var status string
	if err := db.SQL.QueryRow("SELECT attempts,status FROM followup_jobs WHERE id=?", id).Scan(&attempts, &status); err != nil {
		t.Fatal(err)
	}
	if attempts != 0 || status != "pending" {
		t.Fatalf("cancelled retry mutated job: attempts=%d status=%s", attempts, status)
	}
}
