package service

import (
	"careerprogression/internal/domain"
	"sort"
	"strings"
	"time"
)

type EventFilter struct {
	Kinds []string
	From  *time.Time
	To    *time.Time
	Text  string
}

func FilterEvents(events []domain.CareerEvent, f EventFilter) []domain.CareerEvent {
	allowed := map[string]bool{}
	for _, k := range f.Kinds {
		allowed[k] = true
	}
	out := []domain.CareerEvent{}
	for _, e := range events {
		if len(allowed) > 0 && !allowed[e.Kind] {
			continue
		}
		if f.From != nil && e.OccurredAt.Before(*f.From) {
			continue
		}
		if f.To != nil && !e.OccurredAt.Before(*f.To) {
			continue
		}
		if f.Text != "" && !strings.Contains(strings.ToLower(e.Summary), strings.ToLower(f.Text)) {
			continue
		}
		out = append(out, e)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].OccurredAt.Before(out[j].OccurredAt) })
	return out
}
func FilterEmployment(records []domain.EmploymentRecord, status domain.EmploymentStatus, minSalary int64) []domain.EmploymentRecord {
	out := []domain.EmploymentRecord{}
	for _, r := range records {
		if status != "" && r.Status != status {
			continue
		}
		if r.SalaryMax < minSalary {
			continue
		}
		out = append(out, r)
	}
	return out
}
func RedactEvents(events []domain.CareerEvent, role domain.Role) []domain.CareerEvent {
	out := make([]domain.CareerEvent, len(events))
	copy(out, events)
	if role == domain.RoleEmployer {
		for i := range out {
			out[i].Summary = "feedback restricted"
			out[i].EvidenceURL = ""
		}
	}
	return out
}
