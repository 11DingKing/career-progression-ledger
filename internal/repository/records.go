package repository

import (
	"careerprogression/internal/domain"
	"careerprogression/internal/id"
	"context"
	"database/sql"
	"time"
)

func (r *SQLite) AddSkill(ctx context.Context, tx *sql.Tx, s domain.Skill) error {
	if s.ID == "" {
		s.ID = id.New("skill")
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	_, e := tx.ExecContext(ctx, "INSERT INTO skills(id,graduate_id,name,level,verified_at,created_at) VALUES(?,?,?,?,?,?)", s.ID, s.GraduateID, s.Name, s.Level, nullableTime(s.VerifiedAt), s.CreatedAt)
	return e
}
func (r *SQLite) Skills(ctx context.Context, gid string) ([]domain.Skill, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,graduate_id,name,level,verified_at,created_at FROM skills WHERE graduate_id=? ORDER BY name", gid)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Skill{}
	for rows.Next() {
		var s domain.Skill
		var verified, created sql.NullString
		if e := rows.Scan(&s.ID, &s.GraduateID, &s.Name, &s.Level, &verified, &created); e != nil {
			return nil, e
		}
		if verified.Valid {
			t := parse(verified.String)
			s.VerifiedAt = &t
		}
		s.CreatedAt = parse(created.String)
		out = append(out, s)
	}
	return out, rows.Err()
}
func (r *SQLite) AddTraining(ctx context.Context, tx *sql.Tx, t domain.TrainingRecord) error {
	if t.ID == "" {
		t.ID = id.New("training")
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	_, e := tx.ExecContext(ctx, "INSERT INTO training_records(id,graduate_id,provider,name,completed_at,certificate_no,created_at) VALUES(?,?,?,?,?,?,?)", t.ID, t.GraduateID, t.Provider, t.Name, t.CompletedAt, t.CertificateNo, t.CreatedAt)
	return e
}
func (r *SQLite) Training(ctx context.Context, gid string) ([]domain.TrainingRecord, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,graduate_id,provider,name,completed_at,certificate_no,created_at FROM training_records WHERE graduate_id=? ORDER BY completed_at DESC", gid)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.TrainingRecord{}
	for rows.Next() {
		var t domain.TrainingRecord
		var completed, created string
		if e := rows.Scan(&t.ID, &t.GraduateID, &t.Provider, &t.Name, &completed, &t.CertificateNo, &created); e != nil {
			return nil, e
		}
		t.CompletedAt = parse(completed)
		t.CreatedAt = parse(created)
		out = append(out, t)
	}
	return out, rows.Err()
}
func (r *SQLite) AddAudit(ctx context.Context, tx *sql.Tx, a domain.AuditEvent) error {
	_, e := tx.ExecContext(ctx, "INSERT INTO audit_events(id,actor_id,object_type,object_id,action,result,request_id,details,created_at) VALUES(?,?,?,?,?,?,?,?,?)", a.ID, a.ActorID, a.ObjectType, a.ObjectID, a.Action, a.Result, a.RequestID, a.Details, a.CreatedAt)
	return e
}
func (r *SQLite) AuditFor(ctx context.Context, obj string) ([]domain.AuditEvent, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,actor_id,object_type,object_id,action,result,request_id,details,created_at FROM audit_events WHERE object_id=? ORDER BY created_at", obj)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.AuditEvent{}
	for rows.Next() {
		var a domain.AuditEvent
		if e := rows.Scan(&a.ID, &a.ActorID, &a.ObjectType, &a.ObjectID, &a.Action, &a.Result, &a.RequestID, &a.Details, &a.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
