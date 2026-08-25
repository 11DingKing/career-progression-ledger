package service

import (
	"careerprogression/internal/domain"
	"testing"
	"time"
)

func TestFilterEventsByKindAndText(t *testing.T) {
	now := time.Now()
	events := []domain.CareerEvent{{ID: "b", Kind: "promotion", OccurredAt: now.Add(time.Hour), Summary: "Lead architect"}, {ID: "a", Kind: "training", OccurredAt: now, Summary: "Go course"}, {ID: "c", Kind: "promotion", OccurredAt: now.Add(2 * time.Hour), Summary: "Manager"}}
	got := FilterEvents(events, EventFilter{Kinds: []string{"promotion"}, Text: "lead"})
	if len(got) != 1 || got[0].ID != "b" {
		t.Fatalf("%v", got)
	}
}
func TestFilterEventsByWindow(t *testing.T) {
	now := time.Now()
	events := []domain.CareerEvent{{ID: "a", OccurredAt: now.Add(-time.Hour)}, {ID: "b", OccurredAt: now}, {ID: "c", OccurredAt: now.Add(time.Hour)}}
	from := now.Add(-time.Minute)
	to := now.Add(2 * time.Minute)
	got := FilterEvents(events, EventFilter{From: &from, To: &to})
	if len(got) != 1 || got[0].ID != "b" {
		t.Fatalf("%v", got)
	}
}
func TestFilterEventsCopiesAndSorts(t *testing.T) {
	now := time.Now()
	events := []domain.CareerEvent{{ID: "late", OccurredAt: now.Add(time.Hour)}, {ID: "early", OccurredAt: now}}
	got := FilterEvents(events, EventFilter{})
	if got[0].ID != "early" {
		t.Fatal("not sorted")
	}
	if &got[0] == &events[0] {
		t.Fatal("shared slice")
	}
}
func TestFilterEmployment(t *testing.T) {
	records := []domain.EmploymentRecord{{ID: "a", Status: domain.EmploymentActive, SalaryMax: 5}, {ID: "b", Status: domain.EmploymentEnded, SalaryMax: 10}, {ID: "c", Status: domain.EmploymentActive, SalaryMax: 20}}
	if got := FilterEmployment(records, domain.EmploymentActive, 10); len(got) != 1 || got[0].ID != "c" {
		t.Fatalf("%v", got)
	}
	if got := FilterEmployment(records, "", 0); len(got) != 3 {
		t.Fatal("all")
	}
}
func TestRedactEvents(t *testing.T) {
	events := []domain.CareerEvent{{ID: "a", Summary: "private", EvidenceURL: "http://e"}}
	public := RedactEvents(events, domain.RoleEmployer)
	if public[0].Summary == events[0].Summary || public[0].EvidenceURL != "" {
		t.Fatal("not redacted")
	}
	if events[0].Summary != "private" {
		t.Fatal("mutated source")
	}
	admin := RedactEvents(events, domain.RoleMajorAdmin)
	if admin[0].Summary != "private" {
		t.Fatal("admin redacted")
	}
}
func TestPermissionsMatrix(t *testing.T) {
	roles := []domain.Role{domain.RoleGraduate, domain.RoleEmployer, domain.RoleCounselor, domain.RoleMajorAdmin}
	perms := []Permission{ReadProfile, WriteProfile, VerifyEmployer, FreezeStats, ResolveAppeal}
	for _, r := range roles {
		for _, p := range perms {
			allowed := Allowed(r, p)
			if r == domain.RoleMajorAdmin && !allowed {
				t.Fatalf("admin denied %s", p)
			}
			if Require(r, p) == nil && !allowed {
				t.Fatalf("require mismatch %s %s", r, p)
			}
		}
	}
}
func TestPermissionExpected(t *testing.T) {
	if !Allowed(domain.RoleGraduate, ReadProfile) || !Allowed(domain.RoleGraduate, WriteProfile) {
		t.Fatal("graduate profile")
	}
	if Allowed(domain.RoleGraduate, FreezeStats) {
		t.Fatal("graduate freeze")
	}
	if Allowed(domain.RoleEmployer, WriteProfile) {
		t.Fatal("employer write")
	}
	if !Allowed(domain.RoleCounselor, VerifyEmployer) {
		t.Fatal("counselor verify")
	}
	if !Allowed(domain.RoleMajorAdmin, ResolveAppeal) {
		t.Fatal("admin appeal")
	}
}
