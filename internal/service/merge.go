package service

import (
	"careerprogression/internal/domain"
	"sort"
	"time"
)

type MergeResult struct {
	Records   []domain.EmploymentRecord
	Conflicts []string
}

func MergeEmployment(records []domain.EmploymentRecord) MergeResult {
	sorted := append([]domain.EmploymentRecord(nil), records...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].StartDate.Before(sorted[j].StartDate) })
	out := MergeResult{}
	for _, r := range sorted {
		conflict := false
		for _, x := range out.Records {
			if overlap(x, r) {
				conflict = true
				out.Conflicts = append(out.Conflicts, r.ID)
				break
			}
		}
		if !conflict {
			out.Records = append(out.Records, r)
		}
	}
	return out
}
func overlap(a, b domain.EmploymentRecord) bool {
	aEnd := time.Now().UTC()
	bEnd := time.Now().UTC()
	if a.EndDate != nil {
		aEnd = *a.EndDate
	}
	if b.EndDate != nil {
		bEnd = *b.EndDate
	}
	return a.StartDate.Before(bEnd) && b.StartDate.Before(aEnd)
}
func Timeline(records []domain.EmploymentRecord) []string {
	sort.Slice(records, func(i, j int) bool { return records[i].StartDate.Before(records[j].StartDate) })
	out := make([]string, 0, len(records))
	for _, r := range records {
		out = append(out, r.ID+":"+string(r.Status))
	}
	return out
}
