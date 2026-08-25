package service

import (
	"careerprogression/internal/domain"
	"context"
	"testing"
	"time"
)

func TestHistoryChronology(t *testing.T) {
	h := HistoryService{}
	now := time.Now()
	events := []domain.CareerEvent{{ID: "2", Kind: "promotion", OccurredAt: now.Add(time.Hour)}, {ID: "1", Kind: "hire", OccurredAt: now}}
	if e := h.ValidateChronology(events); e != nil {
		t.Fatal(e)
	}
	duplicate := append(events, domain.CareerEvent{Kind: "hire", OccurredAt: now})
	if e := h.ValidateChronology(duplicate); e == nil {
		t.Fatal("duplicate accepted")
	}
}
func TestHistoryWindow(t *testing.T) {
	h := HistoryService{}
	now := time.Now()
	events := []domain.CareerEvent{{ID: "a", OccurredAt: now.Add(-time.Hour)}, {ID: "b", OccurredAt: now}, {ID: "c", OccurredAt: now.Add(time.Hour)}}
	got := h.Window(events, now.Add(-time.Minute), now.Add(time.Minute))
	if len(got) != 1 || got[0].ID != "b" {
		t.Fatalf("%v", got)
	}
}
func TestHistoryContext(t *testing.T) {
	h := HistoryService{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := h.WithContext(ctx); e == nil {
		t.Fatal("cancel ignored")
	}
	if e := h.WithContext(context.Background()); e != nil {
		t.Fatal(e)
	}
}
func TestReportVariants(t *testing.T) {
	for _, r := range []CareerReport{{}, {Graduate: domain.Graduate{Name: "A", Major: "M"}}, {Employment: []domain.EmploymentRecord{{Status: domain.EmploymentEnded, SalaryMin: 3, SalaryMax: 4}}}} {
		_ = r.Summary()
		_ = r.HasPromotion()
		_ = r.ActiveEmployment()
		_ = r.SalaryBand()
	}
}
