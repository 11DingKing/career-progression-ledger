package domain

import (
	"errors"
	"strings"
	"time"
)

func ValidateEmployment(e EmploymentRecord) error {
	if e.GraduateID == "" || e.EmployerID == "" || strings.TrimSpace(e.Title) == "" {
		return errors.New("graduate, employer and title are required")
	}
	if e.SalaryMin < 0 || e.SalaryMax < e.SalaryMin {
		return errors.New("salary range is invalid")
	}
	if e.StartDate.IsZero() {
		return errors.New("start date is required")
	}
	if e.EndDate != nil && e.EndDate.Before(e.StartDate) {
		return errors.New("end before start")
	}
	return nil
}
func ValidateEvent(e CareerEvent) error {
	if e.EmploymentID == "" || strings.TrimSpace(e.Kind) == "" || e.OccurredAt.IsZero() {
		return errors.New("event fields are required")
	}
	if e.OccurredAt.After(time.Now().UTC().Add(24 * time.Hour)) {
		return errors.New("event cannot be far future")
	}
	return nil
}
func NormalizeEmail(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
func ValidateTimeWindow(start, end time.Time) error {
	if end.Before(start) {
		return errors.New("invalid time window")
	}
	if end.Sub(start) > 10*365*24*time.Hour {
		return errors.New("window too long")
	}
	return nil
}
