package service
import "testing"

func TestAppealResolutionRequiresExplanation(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	_, err := db.SQL.Exec("INSERT INTO employers(id,name,registration_no,contact,created_at,updated_at) VALUES('e6','E6','r6','c',datetime('now'),datetime('now'))")
	if err != nil { t.Fatal(err) }
	_, err = db.SQL.Exec("INSERT INTO employment_records(id,graduate_id,employer_id,title,salary_min,salary_max,start_date,status,version,source,created_at,updated_at) VALUES('emp6','g','e6','dev',1,2,datetime('now'),'active',1,'s',datetime('now'),datetime('now'))")
	if err != nil { t.Fatal(err) }
	_, err = db.SQL.Exec("INSERT INTO appeals(id,graduate_id,employment_id,reason,status,resolution,created_by,resolved_by,created_at,updated_at) VALUES('a6','g','emp6','reason','open','','u','',datetime('now'),datetime('now'))")
	if err != nil { t.Fatal(err) }
	s := &AppealService{DB: db}
	if err := s.Transition(ctx, "a6", "in_review", "", "fixed"); err == nil { t.Fatal("appeal transition accepted without actor") }
}
