package service

import (
	"careerprogression/internal/domain"
	"careerprogression/internal/migrations"
	"careerprogression/internal/repository"
	"careerprogression/internal/store"
	"context"
	"os"
	"testing"
	"time"
)

func advancedDB(t *testing.T) (context.Context, *store.DB, *repository.SQLite) {
	f, e := os.CreateTemp("", "adv-*.db")
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
	_, e = db.SQL.Exec("INSERT INTO users(id,email,name,role,password_hash,active,created_at) VALUES('u','u@e','N','graduate','h',1,datetime('now'))")
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.Exec("INSERT INTO graduates(id,user_id,student_no,name,major,graduation_year,contact,created_at,updated_at) VALUES('g','u','s','N','m',2024,'c',datetime('now'),datetime('now'))")
	if e != nil {
		t.Fatal(e)
	}
	return ctx, db, repository.New(db.SQL)
}
func TestConsentLifecycle(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	s := &ConsentService{DB: db}
	if e := s.Grant(ctx, "g", "salary"); e != nil {
		t.Fatal(e)
	}
	ok, e := s.Allowed(ctx, "g", "salary")
	if e != nil || !ok {
		t.Fatalf("%v %v", ok, e)
	}
	if e = s.Revoke(ctx, "g", "salary"); e != nil {
		t.Fatal(e)
	}
	ok, e = s.Allowed(ctx, "g", "salary")
	if e != nil || ok {
		t.Fatalf("%v %v", ok, e)
	}
	list, e := s.Snapshot(ctx, "g")
	if e != nil || len(list) != 1 || list[0].RevokedAt == nil {
		t.Fatalf("%+v %v", list, e)
	}
}
func TestConsentScopesAreIndependent(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	s := &ConsentService{DB: db}
	for _, scope := range []string{"salary", "skills", "contact"} {
		if e := s.Grant(ctx, "g", scope); e != nil {
			t.Fatal(e)
		}
	}
	if e := s.Revoke(ctx, "g", "skills"); e != nil {
		t.Fatal(e)
	}
	for _, c := range []struct {
		scope string
		want  bool
	}{{"salary", true}, {"skills", false}, {"contact", true}} {
		got, _ := s.Allowed(ctx, "g", c.scope)
		if got != c.want {
			t.Fatalf("%s %v", c.scope, got)
		}
	}
}
func TestAppealLifecycle(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	_, e := db.SQL.Exec("INSERT INTO employers(id,name,registration_no,contact,created_at,updated_at) VALUES('e','E','r','c',datetime('now'),datetime('now'))")
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.Exec("INSERT INTO employment_records(id,graduate_id,employer_id,title,salary_min,salary_max,start_date,status,version,source,created_at,updated_at) VALUES('emp','g','e','dev',1,2,datetime('now'),'active',1,'source',datetime('now'),datetime('now'))")
	if e != nil {
		t.Fatal(e)
	}
	s := &AppealService{DB: db}
	a, e := s.Open(ctx, "g", "emp", "salary correction", "u")
	if e != nil {
		t.Fatal(e)
	}
	if a.Status != domain.AppealOpen {
		t.Fatal("open")
	}
	if e = s.Transition(ctx, a.ID, domain.AppealReview, "counselor", ""); e != nil {
		t.Fatal(e)
	}
	if e = s.Transition(ctx, a.ID, domain.AppealResolved, "admin", "updated"); e != nil {
		t.Fatal(e)
	}
	list, e := s.List(ctx, "g")
	if e != nil || len(list) != 1 || list[0].Status != domain.AppealResolved {
		t.Fatalf("%+v %v", list, e)
	}
}
func TestAppealRejectsInvalidTransition(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	_, _ = db.SQL.Exec("INSERT INTO employers(id,name,registration_no,contact,created_at,updated_at) VALUES('e','E','r','c',datetime('now'),datetime('now'))")
	_, _ = db.SQL.Exec("INSERT INTO employment_records(id,graduate_id,employer_id,title,salary_min,salary_max,start_date,status,version,source,created_at,updated_at) VALUES('emp','g','e','dev',1,2,datetime('now'),'active',1,'source',datetime('now'),datetime('now'))")
	s := &AppealService{DB: db}
	a, e := s.Open(ctx, "g", "emp", "reason", "u")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Transition(ctx, a.ID, domain.AppealResolved, "admin", ""); e == nil {
		t.Fatal("invalid transition")
	}
}
func TestIdempotencyReplaysValue(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	s := &IdempotencyService{DB: db}
	calls := 0
	fn := func() any { calls++; return map[string]any{"id": "x", "ok": true} }
	first, e := s.Execute(ctx, "key", "u", "create", fn)
	if e != nil {
		t.Fatal(e)
	}
	second, e := s.Execute(ctx, "key", "u", "create", fn)
	if e != nil {
		t.Fatal(e)
	}
	if calls != 1 {
		t.Fatalf("calls %d", calls)
	}
	if first.(map[string]any)["id"] != second.(map[string]any)["id"] {
		t.Fatal("replay mismatch")
	}
}
func TestIdempotencyActorAndOperationScope(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	s := &IdempotencyService{DB: db}
	n := 0
	fn := func() any { n++; return n }
	if _, e := s.Execute(ctx, "k", "u", "a", fn); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Execute(ctx, "k", "other", "a", fn); e == nil {
		t.Fatal("duplicate key unexpectedly accepted")
	}
	if _, e := s.Execute(ctx, "k2", "u", "b", fn); e != nil {
		t.Fatal(e)
	}
	if n != 3 {
		t.Fatalf("%d", n)
	}
}
func TestMergeTimeline(t *testing.T) {
	now := time.Now().UTC()
	a := domain.EmploymentRecord{ID: "a", StartDate: now.Add(-48 * time.Hour), EndDate: ptrTime(now.Add(-24 * time.Hour)), Status: domain.EmploymentEnded}
	b := domain.EmploymentRecord{ID: "b", StartDate: now.Add(-12 * time.Hour), Status: domain.EmploymentActive}
	c := domain.EmploymentRecord{ID: "c", StartDate: now.Add(2 * time.Hour), Status: domain.EmploymentDraft}
	result := MergeEmployment([]domain.EmploymentRecord{c, b, a})
	if len(result.Records) != 3 || len(result.Conflicts) != 0 {
		t.Fatalf("%+v", result)
	}
	timeline := Timeline([]domain.EmploymentRecord{c, a, b})
	if timeline[0] != "a:ended" || timeline[2] != "c:draft" {
		t.Fatalf("%v", timeline)
	}
}
func TestMergeDetectsOverlap(t *testing.T) {
	now := time.Now().UTC()
	a := domain.EmploymentRecord{ID: "a", StartDate: now.Add(-2 * time.Hour), EndDate: ptrTime(now.Add(time.Hour))}
	b := domain.EmploymentRecord{ID: "b", StartDate: now.Add(-time.Hour), EndDate: ptrTime(now.Add(2 * time.Hour))}
	result := MergeEmployment([]domain.EmploymentRecord{a, b})
	if len(result.Conflicts) != 1 || result.Conflicts[0] != "b" {
		t.Fatalf("%+v", result)
	}
}
func TestPrivacyRules(t *testing.T) {
	cases := []struct {
		role    domain.Role
		consent bool
		want    Visibility
	}{{domain.RoleGraduate, true, VisibilityPrivate}, {domain.RoleGraduate, false, VisibilityPrivate}, {domain.RoleEmployer, true, VisibilityPublic}, {domain.RoleCounselor, true, VisibilityRestricted}, {domain.RoleMajorAdmin, true, VisibilityRestricted}}
	for _, c := range cases {
		if got := VisibleRole(c.role, c.consent); got != c.want {
			t.Fatalf("%v %v", c, got)
		}
	}
	g := domain.Graduate{Contact: "secret", StudentNo: "123"}
	if RedactGraduate(g, VisibilityPrivate).Contact != "" {
		t.Fatal("private leak")
	}
	if RedactGraduate(g, VisibilityPublic).StudentNo != "masked" {
		t.Fatal("public id leak")
	}
}
func TestCanEditAndAccess(t *testing.T) {
	if !CanEdit(domain.RoleGraduate, true, "contact") {
		t.Fatal("owner cannot edit")
	}
	if CanEdit(domain.RoleEmployer, false, "salary") {
		t.Fatal("employer salary edit")
	}
	if !CanEdit(domain.RoleEmployer, false, "feedback") {
		t.Fatal("feedback denied")
	}
	if CheckAccess(domain.RoleEmployer, "salary") == nil {
		t.Fatal("salary access")
	}
	if CheckAccess(domain.RoleGraduate, "skills") != nil {
		t.Fatal("skills denied")
	}
}
func TestReportSummary(t *testing.T) {
	r := CareerReport{Graduate: domain.Graduate{Name: "N", Major: "M"}, Employment: []domain.EmploymentRecord{{Title: "dev", Status: domain.EmploymentActive, SalaryMin: 1, SalaryMax: 3}}, Events: map[string][]domain.CareerEvent{"e": {{Kind: "promotion"}}}}
	if r.Summary() != "N | M | dev | active" {
		t.Fatal(r.Summary())
	}
	if !r.HasPromotion() || r.ActiveEmployment() != 1 || r.SalaryBand() != "1-3" {
		t.Fatal("report")
	}
}
func ptrTime(t time.Time) *time.Time { return &t }
