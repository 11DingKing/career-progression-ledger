package repository

import (
	"careerprogression/internal/domain"
	"database/sql"
	"testing"
	"time"
)

func TestStaleEmploymentVersionCannotOverwriteNewState(t *testing.T) {
	ctx, db, repo := repoFixture(t)
	now := time.Now().UTC()
	_, err := db.SQL.ExecContext(ctx, "INSERT INTO users(id,email,name,role,password_hash,active,created_at) VALUES('u','u@e','N','graduate','h',1,?)", now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.SQL.ExecContext(ctx, "INSERT INTO graduates(id,user_id,student_no,name,major,graduation_year,contact,created_at,updated_at) VALUES('g','u','s','N','m',2024,'c',?,?)", now, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.SQL.ExecContext(ctx, "INSERT INTO employers(id,name,registration_no,contact,created_at,updated_at) VALUES('e','E','r','c',?,?)", now, now)
	if err != nil {
		t.Fatal(err)
	}
	employment := domain.EmploymentRecord{ID: "emp-private", GraduateID: "g", EmployerID: "e", Title: "developer", SalaryMin: 1, SalaryMax: 2, StartDate: now, Status: domain.EmploymentActive, Version: 1, Source: "employer", CreatedAt: now, UpdatedAt: now}
	if err = db.Tx(ctx, func(tx *sql.Tx) error { return repo.CreateEmployment(ctx, tx, employment) }); err != nil {
		t.Fatal(err)
	}
	if err = db.Tx(ctx, func(tx *sql.Tx) error {
		return repo.UpdateEmploymentStatus(ctx, tx, employment.ID, domain.EmploymentDisputed, 1)
	}); err != nil {
		t.Fatal(err)
	}
	if err = db.Tx(ctx, func(tx *sql.Tx) error {
		return repo.UpdateEmploymentStatus(ctx, tx, employment.ID, domain.EmploymentEnded, 1)
	}); err == nil {
		t.Fatal("stale version overwrote a newer employment state")
	}
	got, err := repo.Employment(ctx, employment.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.EmploymentDisputed || got.Version != 2 {
		t.Fatalf("state changed through stale version: %+v", got)
	}
}
