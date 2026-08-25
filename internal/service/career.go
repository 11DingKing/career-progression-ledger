package service

import (
	"careerprogression/internal/apperr"
	"careerprogression/internal/clock"
	"careerprogression/internal/domain"
	"careerprogression/internal/id"
	"careerprogression/internal/repository"
	"careerprogression/internal/store"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type CareerService struct {
	DB   *store.DB
	Repo *repository.SQLite
}

func (s *CareerService) CreateGraduate(ctx context.Context, g domain.Graduate) error {
	if g.ID == "" {
		g.ID = id.New("grad")
	}
	now := clock.From(ctx).Now()
	g.CreatedAt = now
	g.UpdatedAt = now
	if g.StudentNo == "" || g.Name == "" || g.Major == "" {
		return apperr.New(apperr.Invalid, "graduate fields required")
	}
	return s.DB.Tx(ctx, func(tx *sql.Tx) error { return s.Repo.CreateGraduate(ctx, tx, g) })
}
func (s *CareerService) ListGraduates(ctx context.Context, major string, limit, offset int) ([]domain.Graduate, error) {
	return s.Repo.ListGraduates(ctx, major, limit, offset)
}
func (s *CareerService) AddEmployment(ctx context.Context, e domain.EmploymentRecord) error {
	if e.ID == "" {
		e.ID = id.New("emp")
	}
	if err := domain.ValidateEmployment(e); err != nil {
		return apperr.Wrap(apperr.Invalid, "employment invalid", err)
	}
	if e.Status == "" {
		e.Status = domain.EmploymentDraft
	}
	if e.Version == 0 {
		e.Version = 1
	}
	now := clock.From(ctx).Now()
	e.CreatedAt = now
	e.UpdatedAt = now
	return s.DB.Tx(ctx, func(tx *sql.Tx) error {
		if e.Status == domain.EmploymentActive {
			var n int
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM employment_records WHERE graduate_id=? AND status='active' AND start_date<=? AND (end_date IS NULL OR end_date>?)", e.GraduateID, tmService(e.StartDate), tmService(e.StartDate)).Scan(&n); err != nil {
				return err
			}
			if n > 0 {
				return apperr.New(apperr.Conflict, "graduate already has active employment")
			}
		}
		return s.Repo.CreateEmployment(ctx, tx, e)
	})
}
func tmService(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }
func (s *CareerService) GetEmployment(ctx context.Context, id string) (domain.EmploymentRecord, error) {
	e, err := s.Repo.Employment(ctx, id)
	if err != nil {
		return e, apperr.Wrap(apperr.NotFound, "employment not found", err)
	}
	return e, nil
}
func (s *CareerService) TransitionEmployment(ctx context.Context, id string, to domain.EmploymentStatus, version int) error {
	return s.DB.Tx(ctx, func(tx *sql.Tx) error {
		e, err := s.Repo.Employment(ctx, id)
		if err != nil {
			return apperr.Wrap(apperr.NotFound, "employment not found", err)
		}
		if !e.Status.CanTransition(to) {
			return apperr.New(apperr.Conflict, "invalid employment transition")
		}
		if err := s.Repo.UpdateEmploymentStatus(ctx, tx, id, to, version); err != nil {
			return apperr.Wrap(apperr.Conflict, "employment changed", err)
		}
		return nil
	})
}
func (s *CareerService) AddEvent(ctx context.Context, e domain.CareerEvent) error {
	if e.ID == "" {
		e.ID = id.New("event")
	}
	if err := domain.ValidateEvent(e); err != nil {
		return apperr.Wrap(apperr.Invalid, "event invalid", err)
	}
	now := clock.From(ctx).Now()
	e.CreatedAt = now
	return s.DB.Tx(ctx, func(tx *sql.Tx) error {
		emp, err := s.Repo.Employment(ctx, e.EmploymentID)
		if err != nil {
			return err
		}
		if emp.Status == domain.EmploymentEnded {
			return apperr.New(apperr.Conflict, "ended employment is immutable")
		}
		return s.Repo.AddEvent(ctx, tx, e)
	})
}
func (s *CareerService) Events(ctx context.Context, id string) ([]domain.CareerEvent, error) {
	return s.Repo.Events(ctx, id)
}
func ValidateTimeWindow(start, end time.Time) error {
	if end.Before(start) {
		return errors.New("invalid time window")
	}
	if end.Sub(start) > 10*365*24*time.Hour {
		return fmt.Errorf("window too long")
	}
	return nil
}
