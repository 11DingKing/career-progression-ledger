package repository

import (
	"careerprogression/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SQLite struct{ DB *sql.DB }

func New(db *sql.DB) *SQLite   { return &SQLite{db} }
func tm(t time.Time) string    { return t.UTC().Format(time.RFC3339Nano) }
func parse(s string) time.Time { t, _ := time.Parse(time.RFC3339Nano, s); return t }
func (r *SQLite) CreateUser(ctx context.Context, tx *sql.Tx, u domain.User) error {
	_, e := tx.ExecContext(ctx, "INSERT INTO users(id,email,name,role,password_hash,active,created_at) VALUES(?,?,?,?,?,?,?)", u.ID, u.Email, u.Name, u.Role, u.PasswordHash, boolInt(u.Active), tm(u.CreatedAt))
	return e
}
func (r *SQLite) ByEmail(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	var active int
	var created string
	e := r.DB.QueryRowContext(ctx, "SELECT id,email,name,role,password_hash,active,created_at FROM users WHERE email=?", email).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.PasswordHash, &active, &created)
	u.Active = active == 1
	u.CreatedAt = parse(created)
	return u, e
}
func (r *SQLite) ByUserID(ctx context.Context, id string) (domain.User, error) {
	var u domain.User
	var active int
	var created string
	e := r.DB.QueryRowContext(ctx, "SELECT id,email,name,role,password_hash,active,created_at FROM users WHERE id=?", id).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.PasswordHash, &active, &created)
	u.Active = active == 1
	u.CreatedAt = parse(created)
	return u, e
}
func (r *SQLite) CreateSession(ctx context.Context, tx *sql.Tx, s domain.Session) error {
	_, e := tx.ExecContext(ctx, "INSERT INTO sessions(id,user_id,expires_at,revoked_at,created_at) VALUES(?,?,?,?,?)", s.ID, s.UserID, tm(s.ExpiresAt), nil, tm(s.CreatedAt))
	return e
}
func (r *SQLite) GetSession(ctx context.Context, id string) (domain.Session, error) {
	var s domain.Session
	var ex, cr string
	var rv sql.NullString
	e := r.DB.QueryRowContext(ctx, "SELECT id,user_id,expires_at,revoked_at,created_at FROM sessions WHERE id=?", id).Scan(&s.ID, &s.UserID, &ex, &rv, &cr)
	s.ExpiresAt = parse(ex)
	s.CreatedAt = parse(cr)
	if rv.Valid {
		t := parse(rv.String)
		s.RevokedAt = &t
	}
	return s, e
}
func (r *SQLite) RevokeSession(ctx context.Context, tx *sql.Tx, id string) error {
	res, e := tx.ExecContext(ctx, "UPDATE sessions SET revoked_at=? WHERE id=? AND revoked_at IS NULL", tm(time.Now()), id)
	if e == nil {
		n, _ := res.RowsAffected()
		if n == 0 {
			return fmt.Errorf("session already revoked")
		}
	}
	return e
}
func (r *SQLite) CreateGraduate(ctx context.Context, tx *sql.Tx, g domain.Graduate) error {
	_, e := tx.ExecContext(ctx, "INSERT INTO graduates(id,user_id,student_no,name,major,graduation_year,contact,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)", g.ID, g.UserID, g.StudentNo, g.Name, g.Major, g.GraduationYear, g.Contact, tm(g.CreatedAt), tm(g.UpdatedAt))
	return e
}
func (r *SQLite) Graduate(ctx context.Context, id string) (domain.Graduate, error) {
	var g domain.Graduate
	var cr, up string
	e := r.DB.QueryRowContext(ctx, "SELECT id,user_id,student_no,name,major,graduation_year,contact,created_at,updated_at FROM graduates WHERE id=?", id).Scan(&g.ID, &g.UserID, &g.StudentNo, &g.Name, &g.Major, &g.GraduationYear, &g.Contact, &cr, &up)
	g.CreatedAt = parse(cr)
	g.UpdatedAt = parse(up)
	return g, e
}
func (r *SQLite) ListGraduates(ctx context.Context, major string, limit, offset int) ([]domain.Graduate, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,user_id,student_no,name,major,graduation_year,contact,created_at,updated_at FROM graduates WHERE (?='' OR major=?) ORDER BY student_no LIMIT ? OFFSET ?", major, major, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Graduate{}
	for rows.Next() {
		var g domain.Graduate
		var cr, up string
		if e := rows.Scan(&g.ID, &g.UserID, &g.StudentNo, &g.Name, &g.Major, &g.GraduationYear, &g.Contact, &cr, &up); e != nil {
			return nil, e
		}
		g.CreatedAt = parse(cr)
		g.UpdatedAt = parse(up)
		out = append(out, g)
	}
	return out, rows.Err()
}
func (r *SQLite) CreateEmployment(ctx context.Context, tx *sql.Tx, e domain.EmploymentRecord) error {
	_, x := tx.ExecContext(ctx, "INSERT INTO employment_records(id,graduate_id,employer_id,title,salary_min,salary_max,start_date,end_date,status,version,source,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)", e.ID, e.GraduateID, e.EmployerID, e.Title, e.SalaryMin, e.SalaryMax, tm(e.StartDate), nullableTime(e.EndDate), e.Status, e.Version, e.Source, tm(e.CreatedAt), tm(e.UpdatedAt))
	return x
}
func (r *SQLite) CountActiveEmployment(ctx context.Context, graduateID string, at time.Time) (int, error) {
	var count int
	err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM employment_records WHERE graduate_id=? AND status='active' AND start_date<=? AND (end_date IS NULL OR end_date>?)", graduateID, tm(at), tm(at)).Scan(&count)
	return count, err
}
func (r *SQLite) Employment(ctx context.Context, id string) (domain.EmploymentRecord, error) {
	var e domain.EmploymentRecord
	var st, cr, up string
	var en sql.NullString
	err := r.DB.QueryRowContext(ctx, "SELECT id,graduate_id,employer_id,title,salary_min,salary_max,start_date,end_date,status,version,source,created_at,updated_at FROM employment_records WHERE id=?", id).Scan(&e.ID, &e.GraduateID, &e.EmployerID, &e.Title, &e.SalaryMin, &e.SalaryMax, &st, &en, &e.Status, &e.Version, &e.Source, &cr, &up)
	e.StartDate = parse(st)
	if en.Valid {
		t := parse(en.String)
		e.EndDate = &t
	}
	e.CreatedAt = parse(cr)
	e.UpdatedAt = parse(up)
	return e, err
}
func (r *SQLite) UpdateEmploymentStatus(ctx context.Context, tx *sql.Tx, id string, status domain.EmploymentStatus, version int) error {
	res, e := tx.ExecContext(ctx, "UPDATE employment_records SET status=?,version=version+1,updated_at=? WHERE id=? AND version=?", status, tm(time.Now()), id, version)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("version conflict")
	}
	return nil
}
func (r *SQLite) AddEvent(ctx context.Context, tx *sql.Tx, e domain.CareerEvent) error {
	_, x := tx.ExecContext(ctx, "INSERT INTO career_events(id,employment_id,kind,occurred_at,summary,evidence_url,created_by,created_at) VALUES(?,?,?,?,?,?,?,?)", e.ID, e.EmploymentID, e.Kind, tm(e.OccurredAt), e.Summary, e.EvidenceURL, e.CreatedBy, tm(e.CreatedAt))
	return x
}
func (r *SQLite) Events(ctx context.Context, id string) ([]domain.CareerEvent, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,employment_id,kind,occurred_at,summary,evidence_url,created_by,created_at FROM career_events WHERE employment_id=? ORDER BY occurred_at", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.CareerEvent{}
	for rows.Next() {
		var c domain.CareerEvent
		var oc, cr string
		if e := rows.Scan(&c.ID, &c.EmploymentID, &c.Kind, &oc, &c.Summary, &c.EvidenceURL, &c.CreatedBy, &cr); e != nil {
			return nil, e
		}
		c.OccurredAt = parse(oc)
		c.CreatedAt = parse(cr)
		out = append(out, c)
	}
	return out, rows.Err()
}
func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return tm(*t)
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
