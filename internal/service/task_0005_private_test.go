package service

import (
	"testing"
)

func TestIdempotencyScopesReplayByOperation(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	s := &IdempotencyService{DB: db}
	if _, err := s.Execute(ctx, "same-key", "u", "grant", func() any { return map[string]any{"value": "grant"} }); err != nil { t.Fatal(err) }
	v, err := s.Execute(ctx, "same-key", "u", "revoke", func() any { return map[string]any{"value": "revoke"} })
	if err != nil { t.Fatal(err) }
	got := v.(map[string]any)["value"]
	if got != "revoke" { t.Fatalf("operation replay crossed boundary: %v", got) }
}
