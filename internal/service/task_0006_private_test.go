package service
import "testing"

func TestAppealResolutionRequiresExplanation(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	_, err := db.SQL.Exec("INSERT INTO appeals(id,graduate_id,employment_id,reason,status,resolution,created_by,resolved_by,created_at,updated_at) VALUES('a6','g','e','reason','open','','u','',datetime('now'),datetime('now'))")
	if err != nil { t.Fatal(err) }
	s := &AppealService{DB: db}
	if err := s.Transition(ctx, "a6", "resolved", "", "fixed"); err == nil { t.Fatal("appeal transition accepted without actor") }
}
