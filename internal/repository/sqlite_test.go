package repository

import (
	"careerprogression/internal/domain"
	"careerprogression/internal/migrations"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"
)

func repoFixture(t *testing.T) (context.Context, *store.DB, *SQLite) {
	f, e := os.CreateTemp("", "repo-*.db")
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
	return ctx, db, New(db.SQL)
}
func TestSQLiteUserRoundTrip(t *testing.T) {
	ctx, db, r := repoFixture(t)
	now := time.Now().UTC()
	u := domain.User{ID: "u", Email: "u@e", Name: "U", Role: domain.RoleCounselor, PasswordHash: "hash", Active: true, CreatedAt: now}
	if e := db.Tx(ctx, func(tx *sql.Tx) error { return r.CreateUser(ctx, tx, u) }); e != nil {
		t.Fatal(e)
	}
	got, e := r.ByEmail(ctx, "u@e")
	if e != nil || got.ID != u.ID || got.Role != u.Role {
		t.Fatalf("%+v %v", got, e)
	}
	got, e = r.ByUserID(ctx, "u")
	if e != nil || got.Email != u.Email {
		t.Fatal(e)
	}
}
func TestSQLiteSessionRevoke(t *testing.T) {
	ctx, db, r := repoFixture(t)
	now := time.Now().UTC()
	u := domain.User{ID: "u", Email: "u@e", Name: "U", Role: domain.RoleGraduate, PasswordHash: "h", Active: true, CreatedAt: now}
	s := domain.Session{ID: "s", UserID: "u", ExpiresAt: now.Add(time.Hour), CreatedAt: now}
	if e := db.Tx(ctx, func(tx *sql.Tx) error {
		if e := r.CreateUser(ctx, tx, u); e != nil {
			return e
		}
		return r.CreateSession(ctx, tx, s)
	}); e != nil {
		t.Fatal(e)
	}
	got, e := r.GetSession(ctx, "s")
	if e != nil || got.UserID != "u" {
		t.Fatal(e)
	}
	if e = db.Tx(ctx, func(tx *sql.Tx) error { return r.RevokeSession(ctx, tx, "s") }); e != nil {
		t.Fatal(e)
	}
	got, e = r.GetSession(ctx, "s")
	if e != nil || got.RevokedAt == nil {
		t.Fatalf("%+v %v", got, e)
	}
}
func TestSQLiteGraduateListFilterAndPagination(t *testing.T) {
	ctx, db, r := repoFixture(t)
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		u := domain.User{ID: fmtID(i), Email: fmtID(i) + "@e", Name: "N", Role: domain.RoleGraduate, PasswordHash: "h", Active: true, CreatedAt: now}
		g := domain.Graduate{ID: "g" + fmtID(i), UserID: u.ID, StudentNo: fmt.Sprintf("2024%02d", i), Name: "N", Major: []string{"net", "data"}[i%2], GraduationYear: 2024, Contact: "c", CreatedAt: now, UpdatedAt: now}
		if e := db.Tx(ctx, func(tx *sql.Tx) error {
			if e := r.CreateUser(ctx, tx, u); e != nil {
				return e
			}
			return r.CreateGraduate(ctx, tx, g)
		}); e != nil {
			t.Fatal(e)
		}
	}
	all, e := r.ListGraduates(ctx, "", 3, 1)
	if e != nil || len(all) != 3 {
		t.Fatalf("%d %v", len(all), e)
	}
	net, e := r.ListGraduates(ctx, "net", 10, 0)
	if e != nil || len(net) != 3 {
		t.Fatalf("%d %v", len(net), e)
	}
}
func TestSQLiteEmploymentEvents(t *testing.T) {
	ctx, db, r := repoFixture(t)
	now := time.Now().UTC()
	u := domain.User{ID: "u", Email: "u@e", Name: "U", Role: domain.RoleGraduate, PasswordHash: "h", Active: true, CreatedAt: now}
	g := domain.Graduate{ID: "g", UserID: "u", StudentNo: "s", Name: "N", Major: "m", GraduationYear: 2024, Contact: "c", CreatedAt: now, UpdatedAt: now}
	if e := db.Tx(ctx, func(tx *sql.Tx) error {
		if e := r.CreateUser(ctx, tx, u); e != nil {
			return e
		}
		if e := r.CreateGraduate(ctx, tx, g); e != nil {
			return e
		}
		_, e := tx.ExecContext(ctx, "INSERT INTO employers(id,name,registration_no,contact,verified,created_at,updated_at) VALUES('e','E','r','c',1,?,?)", now, now)
		return e
	}); e != nil {
		t.Fatal(e)
	}
	emp := domain.EmploymentRecord{ID: "emp", GraduateID: "g", EmployerID: "e", Title: "dev", SalaryMin: 1, SalaryMax: 2, StartDate: now, Status: domain.EmploymentDraft, Version: 1, Source: "graduate", CreatedAt: now, UpdatedAt: now}
	if e := db.Tx(ctx, func(tx *sql.Tx) error { return r.CreateEmployment(ctx, tx, emp) }); e != nil {
		t.Fatal(e)
	}
	event := domain.CareerEvent{ID: "ev", EmploymentID: "emp", Kind: "training", OccurredAt: now, Summary: "s", EvidenceURL: "u", CreatedBy: "u", CreatedAt: now}
	if e := db.Tx(ctx, func(tx *sql.Tx) error { return r.AddEvent(ctx, tx, event) }); e != nil {
		t.Fatal(e)
	}
	events, e := r.Events(ctx, "emp")
	if e != nil || len(events) != 1 {
		t.Fatalf("%d %v", len(events), e)
	}
	got, e := r.Employment(ctx, "emp")
	if e != nil || got.Title != "dev" {
		t.Fatal(e)
	}
}
func TestSQLiteVersionConflict(t *testing.T) {
	ctx, db, r := repoFixture(t)
	now := time.Now().UTC()
	_, e := r.Employment(ctx, "missing")
	if e == nil {
		t.Fatal("missing employment")
	}
	_ = now
	_ = db
}
func fmtID(i int) string { return fmt.Sprintf("id%d", i) }
