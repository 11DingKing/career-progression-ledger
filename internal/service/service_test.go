package service

import (
	"careerprogression/internal/auth"
	"careerprogression/internal/clock"
	"careerprogression/internal/domain"
	"careerprogression/internal/migrations"
	"careerprogression/internal/repository"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"
	"time"
)

func fixture(t *testing.T) (*store.DB, *repository.SQLite, context.Context) {
	t.Helper()
	f, e := os.CreateTemp("", "career-*.db")
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
	return db, repository.New(db.SQL), ctx
}
func seed(t *testing.T, db *store.DB, r *repository.SQLite, ctx context.Context) (domain.User, domain.Graduate, domain.Employer) {
	now := time.Now().UTC()
	u := domain.User{ID: "u1", Email: "grad@example.com", Name: "Lin", Role: domain.RoleGraduate, PasswordHash: auth.HashPassword("pw"), Active: true, CreatedAt: now}
	g := domain.Graduate{ID: "g1", UserID: u.ID, StudentNo: "2024001", Name: "Lin", Major: "network", GraduationYear: 2024, Contact: "masked", CreatedAt: now, UpdatedAt: now}
	e := domain.Employer{ID: "e1", Name: "Acme", RegistrationNo: "REG-1", Contact: "hr", CreatedAt: now, UpdatedAt: now}
	if err := db.Tx(ctx, func(tx *sql.Tx) error {
		if err := r.CreateUser(ctx, tx, u); err != nil {
			return err
		}
		if err := r.CreateGraduate(ctx, tx, g); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO employers(id,name,registration_no,contact,verified,created_at,updated_at) VALUES(?,?,?,?,1,?,?)", e.ID, e.Name, e.RegistrationNo, e.Contact, now, now)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return u, g, e
}
func TestAuthLoginLogoutAndExpiry(t *testing.T) {
	db, r, ctx := fixture(t)
	u, _, _ := seed(t, db, r, ctx)
	a := &AuthService{DB: db, Repo: r, TTL: time.Hour}
	sess, got, e := a.Login(ctx, u.Email, "pw")
	if e != nil || got.ID != u.ID {
		t.Fatalf("login %v", e)
	}
	if _, e = a.Authenticate(ctx, sess.ID); e != nil {
		t.Fatalf("auth %v", e)
	}
	if e = a.Logout(ctx, sess.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = a.Authenticate(ctx, sess.ID); e == nil {
		t.Fatal("revoked session accepted")
	}
	short := &AuthService{DB: db, Repo: r, TTL: time.Nanosecond}
	s, _, e := short.Login(ctx, u.Email, "pw")
	if e != nil {
		t.Fatal(e)
	}
	time.Sleep(time.Millisecond)
	if _, e = short.Authenticate(ctx, s.ID); e == nil {
		t.Fatal("expired accepted")
	}
}
func TestAuthWrongPassword(t *testing.T) {
	db, r, ctx := fixture(t)
	u, _, _ := seed(t, db, r, ctx)
	a := &AuthService{DB: db, Repo: r, TTL: time.Hour}
	if _, _, e := a.Login(ctx, u.Email, "bad"); e == nil {
		t.Fatal("wrong password")
	}
	if _, _, e := a.Login(ctx, "missing@example.com", "pw"); e == nil {
		t.Fatal("missing user")
	}
}
func TestGraduateAndEmploymentFlow(t *testing.T) {
	db, r, ctx := fixture(t)
	u, g, e := seed(t, db, r, ctx)
	_ = u
	cs := &CareerService{DB: db, Repo: r}
	start := time.Now().UTC()
	emp := domain.EmploymentRecord{ID: "emp-flow", GraduateID: g.ID, EmployerID: e.ID, Title: "analyst", SalaryMin: 5000, SalaryMax: 7000, StartDate: start, Status: domain.EmploymentActive, Source: "employer"}
	if err := cs.AddEmployment(ctx, emp); err != nil {
		t.Fatal(err)
	}
	list, err := cs.ListGraduates(ctx, "network", 25, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("list %v %d", err, len(list))
	}
	stored, err := r.Employment(ctx, emp.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.EmploymentActive {
		t.Fatal("status")
	}
	if err := cs.AddEvent(ctx, domain.CareerEvent{EmploymentID: emp.ID, Kind: "promotion", OccurredAt: start.Add(time.Hour), Summary: "lead", CreatedBy: u.ID}); err != nil {
		t.Fatal(err)
	}
	events, err := cs.Events(ctx, emp.ID)
	if err != nil || len(events) != 1 {
		t.Fatalf("events %v", err)
	}
}
func TestEmploymentConflictAndTransition(t *testing.T) {
	db, r, ctx := fixture(t)
	_, g, e := seed(t, db, r, ctx)
	cs := &CareerService{DB: db, Repo: r}
	now := time.Now().UTC()
	one := domain.EmploymentRecord{ID: "emp1", GraduateID: g.ID, EmployerID: e.ID, Title: "one", SalaryMin: 1, SalaryMax: 2, StartDate: now, Status: domain.EmploymentActive}
	if err := cs.AddEmployment(ctx, one); err != nil {
		t.Fatal(err)
	}
	two := one
	two.ID = "emp2"
	two.Title = "two"
	if err := cs.AddEmployment(ctx, two); err == nil {
		t.Fatal("overlap accepted")
	}
	if err := cs.TransitionEmployment(ctx, one.ID, domain.EmploymentEnded, 1); err != nil {
		t.Fatal(err)
	}
	if err := cs.TransitionEmployment(ctx, one.ID, domain.EmploymentActive, 2); err == nil {
		t.Fatal("ended reopened")
	}
}
func TestOptimisticVersion(t *testing.T) {
	db, r, ctx := fixture(t)
	_, g, e := seed(t, db, r, ctx)
	cs := &CareerService{DB: db, Repo: r}
	now := time.Now().UTC()
	emp := domain.EmploymentRecord{ID: "emp-v", GraduateID: g.ID, EmployerID: e.ID, Title: "v", SalaryMin: 1, SalaryMax: 2, StartDate: now}
	if err := cs.AddEmployment(ctx, emp); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- cs.TransitionEmployment(ctx, emp.ID, domain.EmploymentActive, 1) }()
	}
	wg.Wait()
	close(errs)
	success := 0
	for err := range errs {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successes %d", success)
	}
}
func TestContextCancellation(t *testing.T) {
	db, r, ctx := fixture(t)
	_, g, e := seed(t, db, r, ctx)
	cs := &CareerService{DB: db, Repo: r}
	cctx, cancel := context.WithCancel(ctx)
	cancel()
	err := cs.AddEmployment(cctx, domain.EmploymentRecord{GraduateID: g.ID, EmployerID: e.ID, Title: "cancel", SalaryMin: 1, SalaryMax: 2, StartDate: time.Now()})
	if err == nil {
		t.Fatal("cancelled context accepted")
	}
}
func TestFixedClock(t *testing.T) {
	db, r, ctx := fixture(t)
	_, g, e := seed(t, db, r, ctx)
	fixed := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	ctx = clock.With(ctx, clock.Fixed{T: fixed})
	cs := &CareerService{DB: db, Repo: r}
	emp := domain.EmploymentRecord{ID: "clock", GraduateID: g.ID, EmployerID: e.ID, Title: "clock", SalaryMin: 1, SalaryMax: 2, StartDate: fixed}
	if err := cs.AddEmployment(ctx, emp); err != nil {
		t.Fatal(err)
	}
	got, _ := r.Employment(ctx, emp.ID)
	if !got.CreatedAt.Equal(fixed) {
		t.Fatalf("created %v", got.CreatedAt)
	}
}
func TestTimeWindowErrors(t *testing.T) {
	now := time.Now()
	if ValidateTimeWindow(now, now.Add(-time.Minute)) == nil {
		t.Fatal("reverse")
	}
	if ValidateTimeWindow(now, now.Add(11*365*24*time.Hour)) == nil {
		t.Fatal("long")
	}
	if ValidateTimeWindow(now, now.Add(time.Minute)) != nil {
		t.Fatal("valid")
	}
	_ = errors.New
}
