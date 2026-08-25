package repository

import (
	"careerprogression/internal/domain"
	"context"
	"database/sql"
)

type UserRepo interface {
	Create(context.Context, *sql.Tx, domain.User) error
	ByEmail(context.Context, string) (domain.User, error)
	ByID(context.Context, string) (domain.User, error)
}
type SessionRepo interface {
	Create(context.Context, *sql.Tx, domain.Session) error
	Get(context.Context, string) (domain.Session, error)
	Revoke(context.Context, *sql.Tx, string) error
}
type GraduateRepo interface {
	Create(context.Context, *sql.Tx, domain.Graduate) error
	ByID(context.Context, string) (domain.Graduate, error)
	List(context.Context, string, int, int) ([]domain.Graduate, error)
}
type EmploymentRepo interface {
	Create(context.Context, *sql.Tx, domain.EmploymentRecord) error
	ByID(context.Context, string) (domain.EmploymentRecord, error)
	UpdateStatus(context.Context, *sql.Tx, string, domain.EmploymentStatus, int) error
	AddEvent(context.Context, *sql.Tx, domain.CareerEvent) error
	Events(context.Context, string) ([]domain.CareerEvent, error)
}
