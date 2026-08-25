package domain

import "time"

type Role string

const (
	RoleGraduate   Role = "graduate"
	RoleEmployer   Role = "employer"
	RoleCounselor  Role = "counselor"
	RoleMajorAdmin Role = "major_admin"
)

type EmploymentStatus string

const (
	EmploymentDraft    EmploymentStatus = "draft"
	EmploymentActive   EmploymentStatus = "active"
	EmploymentEnded    EmploymentStatus = "ended"
	EmploymentDisputed EmploymentStatus = "disputed"
)

type AppealStatus string

const (
	AppealOpen     AppealStatus = "open"
	AppealReview   AppealStatus = "in_review"
	AppealResolved AppealStatus = "resolved"
	AppealRejected AppealStatus = "rejected"
)

type User struct {
	ID           string
	Email        string
	Name         string
	Role         Role
	PasswordHash string
	Active       bool
	CreatedAt    time.Time
}
type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
type Graduate struct {
	ID             string
	UserID         string
	StudentNo      string
	Name           string
	Major          string
	GraduationYear int
	Contact        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
type Employer struct {
	ID             string
	Name           string
	RegistrationNo string
	Contact        string
	Verified       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
type EmploymentRecord struct {
	ID         string
	GraduateID string
	EmployerID string
	Title      string
	SalaryMin  int64
	SalaryMax  int64
	StartDate  time.Time
	EndDate    *time.Time
	Status     EmploymentStatus
	Version    int
	Source     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
type CareerEvent struct {
	ID           string
	EmploymentID string
	Kind         string
	OccurredAt   time.Time
	Summary      string
	EvidenceURL  string
	CreatedBy    string
	CreatedAt    time.Time
}
type Skill struct {
	ID         string
	GraduateID string
	Name       string
	Level      int
	VerifiedAt *time.Time
	CreatedAt  time.Time
}
type TrainingRecord struct {
	ID            string
	GraduateID    string
	Provider      string
	Name          string
	CompletedAt   time.Time
	CertificateNo string
	CreatedAt     time.Time
}
type Consent struct {
	ID         string
	GraduateID string
	Scope      string
	Granted    bool
	GrantedAt  time.Time
	RevokedAt  *time.Time
	Version    int
}
type Appeal struct {
	ID           string
	GraduateID   string
	EmploymentID string
	Reason       string
	Status       AppealStatus
	Resolution   string
	CreatedBy    string
	ResolvedBy   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
type AuditEvent struct {
	ID         string
	ActorID    string
	ObjectType string
	ObjectID   string
	Action     string
	Result     string
	RequestID  string
	Details    string
	CreatedAt  time.Time
}
type IdempotencyKey struct {
	Key       string
	ActorID   string
	Operation string
	Response  string
	CreatedAt time.Time
}
type FollowupJob struct {
	ID         string
	GraduateID string
	Kind       string
	DueAt      time.Time
	Attempts   int
	Status     string
	LastError  string
	LockedAt   *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
type StatisticSnapshot struct {
	ID             string
	Major          string
	GraduationYear int
	Total          int
	Promoted       int
	FrozenAt       time.Time
	FrozenBy       string
}

func (e EmploymentStatus) CanTransition(to EmploymentStatus) bool {
	switch e {
	case EmploymentDraft:
		return to == EmploymentActive || to == EmploymentDisputed
	case EmploymentActive:
		return to == EmploymentEnded || to == EmploymentDisputed
	case EmploymentDisputed:
		return to == EmploymentActive || to == EmploymentEnded
	case EmploymentEnded:
		return false
	}
	return false
}
func (a AppealStatus) CanTransition(to AppealStatus) bool {
	switch a {
	case AppealOpen:
		return to == AppealReview || to == AppealRejected
	case AppealReview:
		return to == AppealResolved || to == AppealRejected
	default:
		return false
	}
}
