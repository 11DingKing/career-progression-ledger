package domain

import (
	"testing"
	"time"
)

func TestEmploymentValidSalaryBoundaries(t *testing.T) {
	now := time.Now()
	for _, c := range []struct {
		min, max int64
		valid    bool
	}{{0, 0, true}, {0, 1, true}, {100, 100, true}, {100, 99, false}, {-1, 1, false}} {
		e := EmploymentRecord{GraduateID: "g", EmployerID: "e", Title: "t", SalaryMin: c.min, SalaryMax: c.max, StartDate: now}
		if (ValidateEmployment(e) == nil) != c.valid {
			t.Fatalf("%+v", c)
		}
	}
}
func TestEmploymentDateBoundaries(t *testing.T) {
	now := time.Now()
	same := now
	e := EmploymentRecord{GraduateID: "g", EmployerID: "e", Title: "t", SalaryMin: 1, SalaryMax: 1, StartDate: now, EndDate: &same}
	if ValidateEmployment(e) != nil {
		t.Fatal("same day invalid")
	}
	before := now.Add(-time.Nanosecond)
	e.EndDate = &before
	if ValidateEmployment(e) == nil {
		t.Fatal("before accepted")
	}
}
func TestEventKinds(t *testing.T) {
	now := time.Now()
	for _, kind := range []string{"hire", "promotion", "transfer", "qualification", "training", "salary_review", "leave", "return"} {
		e := CareerEvent{EmploymentID: "e", Kind: kind, OccurredAt: now}
		if ValidateEvent(e) != nil {
			t.Fatal(kind)
		}
	}
}
func TestEventRequiredFields(t *testing.T) {
	now := time.Now()
	cases := []CareerEvent{{Kind: "x", OccurredAt: now}, {EmploymentID: "e", OccurredAt: now}, {EmploymentID: "e", Kind: "x"}}
	for _, e := range cases {
		if ValidateEvent(e) == nil {
			t.Fatalf("accepted %+v", e)
		}
	}
}
func TestTransitionExhaustive(t *testing.T) {
	statuses := []EmploymentStatus{EmploymentDraft, EmploymentActive, EmploymentEnded, EmploymentDisputed}
	for _, from := range statuses {
		for _, to := range statuses {
			got := from.CanTransition(to)
			if from == EmploymentEnded && got {
				t.Fatal("ended transition")
			}
			if from == EmploymentDraft && to == EmploymentEnded && got {
				t.Fatal("draft ended")
			}
		}
	}
}
func TestAppealTerminalStates(t *testing.T) {
	for _, terminal := range []AppealStatus{AppealResolved, AppealRejected} {
		for _, to := range []AppealStatus{AppealOpen, AppealReview, AppealResolved, AppealRejected} {
			if terminal.CanTransition(to) {
				t.Fatalf("terminal %s to %s", terminal, to)
			}
		}
	}
}
func TestNormalizeEmailVariants(t *testing.T) {
	cases := map[string]string{" A@B.COM ": "a@b.com", "x@y": "x@y", "Mixed@Example.Org": "mixed@example.org", "  ": ""}
	for in, want := range cases {
		if got := NormalizeEmail(in); got != want {
			t.Fatalf("%q=%q", in, got)
		}
	}
}
func TestTimeWindowExactLimit(t *testing.T) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if ValidateTimeWindow(start, start.Add(10*365*24*time.Hour)) != nil {
		t.Fatal("exact limit")
	}
	if ValidateTimeWindow(start, start.Add(10*365*24*time.Hour+time.Nanosecond)) == nil {
		t.Fatal("over limit")
	}
}
func TestTimeWindowZero(t *testing.T) {
	now := time.Now()
	if ValidateTimeWindow(now, now) != nil {
		t.Fatal("zero rejected")
	}
}
func TestStatusValuesStable(t *testing.T) {
	if EmploymentDraft != "draft" || EmploymentActive != "active" || EmploymentEnded != "ended" || EmploymentDisputed != "disputed" {
		t.Fatal("status values changed")
	}
}
func TestAppealValuesStable(t *testing.T) {
	if AppealOpen != "open" || AppealReview != "in_review" || AppealResolved != "resolved" || AppealRejected != "rejected" {
		t.Fatal("appeal values changed")
	}
}
func TestRoleValuesStable(t *testing.T) {
	if RoleGraduate != "graduate" || RoleEmployer != "employer" || RoleCounselor != "counselor" || RoleMajorAdmin != "major_admin" {
		t.Fatal("role values changed")
	}
}
func TestGraduateDefaults(t *testing.T) {
	g := Graduate{}
	if g.ID != "" || g.GraduationYear != 0 {
		t.Fatal("unexpected defaults")
	}
}
func TestEmployerDefaults(t *testing.T) {
	e := Employer{}
	if e.Verified {
		t.Fatal("verified default")
	}
}
func TestConsentDefaults(t *testing.T) {
	c := Consent{}
	if c.Granted || c.Version != 0 {
		t.Fatal("consent defaults")
	}
}
func TestAppealDefaults(t *testing.T) {
	a := Appeal{}
	if a.Status != "" || a.Reason != "" {
		t.Fatal("appeal defaults")
	}
}
func TestJobDefaults(t *testing.T) {
	j := FollowupJob{}
	if j.Attempts != 0 || j.Status != "" {
		t.Fatal("job defaults")
	}
}
