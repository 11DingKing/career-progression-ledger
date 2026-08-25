package service

import (
	"careerprogression/internal/domain"
	"testing"
	"time"
)

func TestRejectedOverlappingEmploymentDoesNotRemainInTimeline(t *testing.T) {
	db, repo, ctx := fixture(t)
	_, graduate, employer := seed(t, db, repo, ctx)
	svc := &CareerService{DB: db, Repo: repo}
	start := time.Now().UTC()
	first := domain.EmploymentRecord{ID: "overlap-first", GraduateID: graduate.ID, EmployerID: employer.ID, Title: "operator", SalaryMin: 1, SalaryMax: 2, StartDate: start, Status: domain.EmploymentActive}
	if err := svc.AddEmployment(ctx, first); err != nil {
		t.Fatal(err)
	}
	second := first
	second.ID = "overlap-second"
	second.Title = "supervisor"
	if err := svc.AddEmployment(ctx, second); err == nil {
		t.Fatal("overlapping employment was accepted")
	}
	rows, err := db.SQL.QueryContext(ctx, "SELECT id FROM employment_records WHERE graduate_id=? ORDER BY id", graduate.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if len(ids) != 1 || ids[0] != first.ID {
		t.Fatalf("rejected history leaked into timeline: %v", ids)
	}
}
