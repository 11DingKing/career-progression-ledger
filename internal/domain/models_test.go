package domain

import (
	"errors"
	"testing"
	"time"
)

func TestEmploymentTransitions(t *testing.T) {
	cases := []struct {
		from, to EmploymentStatus
		ok       bool
	}{{EmploymentDraft, EmploymentActive, true}, {EmploymentDraft, EmploymentDisputed, true}, {EmploymentActive, EmploymentEnded, true}, {EmploymentActive, EmploymentDisputed, true}, {EmploymentDisputed, EmploymentActive, true}, {EmploymentDisputed, EmploymentEnded, true}, {EmploymentEnded, EmploymentActive, false}, {EmploymentEnded, EmploymentEnded, false}}
	for _, c := range cases {
		if got := c.from.CanTransition(c.to); got != c.ok {
			t.Fatalf("%s to %s=%v", c.from, c.to, got)
		}
	}
}
func TestAppealTransitions(t *testing.T) {
	cases := []struct {
		from, to AppealStatus
		ok       bool
	}{{AppealOpen, AppealReview, true}, {AppealOpen, AppealRejected, true}, {AppealReview, AppealResolved, true}, {AppealReview, AppealRejected, true}, {AppealResolved, AppealOpen, false}, {AppealRejected, AppealReview, false}}
	for _, c := range cases {
		if c.from.CanTransition(c.to) != c.ok {
			t.Fatalf("transition %v %v", c.from, c.to)
		}
	}
}
func TestValidateEmployment(t *testing.T) {
	now := time.Now()
	valid := EmploymentRecord{GraduateID: "g", EmployerID: "e", Title: "developer", SalaryMin: 1, SalaryMax: 2, StartDate: now}
	if err := ValidateEmployment(valid); err != nil {
		t.Fatal(err)
	}
	bad := []EmploymentRecord{{}, {GraduateID: "g", EmployerID: "e", Title: "x", SalaryMin: -1, SalaryMax: 2, StartDate: now}, {GraduateID: "g", EmployerID: "e", Title: "x", SalaryMin: 3, SalaryMax: 2, StartDate: now}, {GraduateID: "g", EmployerID: "e", Title: "x", StartDate: now, EndDate: ptr(now.Add(-time.Hour))}}
	for _, v := range bad {
		if ValidateEmployment(v) == nil {
			t.Fatalf("expected invalid %+v", v)
		}
	}
}
func TestValidateEvent(t *testing.T) {
	if ValidateEvent(CareerEvent{}) == nil {
		t.Fatal("empty event accepted")
	}
	if ValidateEvent(CareerEvent{EmploymentID: "e", Kind: "promotion", OccurredAt: time.Now().Add(48 * time.Hour)}) == nil {
		t.Fatal("future event accepted")
	}
	if ValidateEvent(CareerEvent{EmploymentID: "e", Kind: "promotion", OccurredAt: time.Now()}) != nil {
		t.Fatal("valid event rejected")
	}
}
func TestNormalizeEmail(t *testing.T) {
	if NormalizeEmail("  A@EXAMPLE.COM ") != "a@example.com" {
		t.Fatal("normalization")
	}
}
func TestValidateTimeWindow(t *testing.T) {
	now := time.Now()
	if ValidateTimeWindow(now, now.Add(time.Hour)) != nil {
		t.Fatal("window")
	}
	if ValidateTimeWindow(now, now.Add(-time.Hour)) == nil {
		t.Fatal("reverse")
	}
	if ValidateTimeWindow(now, now.Add(11*365*24*time.Hour)) == nil {
		t.Fatal("long")
	}
}
func ptr(t time.Time) *time.Time { return &t }

var _ = errors.New
