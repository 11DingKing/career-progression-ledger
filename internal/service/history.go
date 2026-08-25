package service

import (
	"careerprogression/internal/domain"
	"context"
	"fmt"
	"sort"
	"time"
)

type HistoryService struct{}

func (h HistoryService) ValidateChronology(events []domain.CareerEvent) error {
	if len(events) == 0 {
		return nil
	}
	copyEvents := append([]domain.CareerEvent(nil), events...)
	sort.Slice(copyEvents, func(i, j int) bool { return copyEvents[i].OccurredAt.Before(copyEvents[j].OccurredAt) })
	for i := 1; i < len(copyEvents); i++ {
		if copyEvents[i].OccurredAt.Equal(copyEvents[i-1].OccurredAt) && copyEvents[i].Kind == copyEvents[i-1].Kind {
			return fmt.Errorf("duplicate event at %s", copyEvents[i].OccurredAt)
		}
	}
	return nil
}
func (h HistoryService) Window(events []domain.CareerEvent, start, end time.Time) []domain.CareerEvent {
	out := []domain.CareerEvent{}
	for _, e := range events {
		if !e.OccurredAt.Before(start) && e.OccurredAt.Before(end) {
			out = append(out, e)
		}
	}
	return out
}
func (h HistoryService) WithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
