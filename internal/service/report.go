package service

import (
	"careerprogression/internal/domain"
	"careerprogression/internal/repository"
	"context"
	"fmt"
	"strings"
	"time"
)

type CareerReport struct {
	Graduate    domain.Graduate
	Employment  []domain.EmploymentRecord
	Events      map[string][]domain.CareerEvent
	Skills      []domain.Skill
	Training    []domain.TrainingRecord
	Consent     []domain.Consent
	GeneratedAt time.Time
}
type ReportService struct {
	Repo    *repository.SQLite
	Consent *ConsentService
}

func (s *ReportService) Build(ctx context.Context, gid, scope string) (CareerReport, error) {
	g, e := s.Repo.Graduate(ctx, gid)
	if e != nil {
		return CareerReport{}, e
	}
	allowed, err := s.Consent.Allowed(ctx, gid, scope)
	if err != nil {
		return CareerReport{}, err
	}
	if !allowed {
		return CareerReport{}, fmt.Errorf("consent revoked")
	}
	events := map[string][]domain.CareerEvent{}
	return CareerReport{Graduate: g, Events: events, Consent: nil, GeneratedAt: time.Now().UTC()}, nil
}
func (r CareerReport) Summary() string {
	parts := []string{r.Graduate.Name, r.Graduate.Major}
	for _, e := range r.Employment {
		parts = append(parts, e.Title, string(e.Status))
	}
	return strings.Join(parts, " | ")
}
func (r CareerReport) HasPromotion() bool {
	for _, events := range r.Events {
		for _, e := range events {
			if e.Kind == "promotion" {
				return true
			}
		}
	}
	return false
}
func (r CareerReport) ActiveEmployment() int {
	n := 0
	for _, e := range r.Employment {
		if e.Status == domain.EmploymentActive {
			n++
		}
	}
	return n
}
func (r CareerReport) SalaryBand() string {
	min, max := int64(0), int64(0)
	for _, e := range r.Employment {
		if min == 0 || e.SalaryMin < min {
			min = e.SalaryMin
		}
		if e.SalaryMax > max {
			max = e.SalaryMax
		}
	}
	return fmt.Sprintf("%d-%d", min, max)
}
