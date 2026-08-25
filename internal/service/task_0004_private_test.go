package service

import (
	"careerprogression/internal/domain"
	"testing"
)

func TestFreezeCountsEndedEmploymentAsPromotion(t *testing.T) {
	ctx, db, _ := advancedDB(t)
	_, err := db.SQL.Exec("INSERT INTO employers(id,name,registration_no,contact,created_at,updated_at) VALUES('e4','E4','r4','c',datetime('now'),datetime('now'))")
	if err != nil { t.Fatal(err) }
	_, err = db.SQL.Exec("INSERT INTO employment_records(id,graduate_id,employer_id,title,salary_min,salary_max,start_date,status,version,source,created_at,updated_at) VALUES('emp4','g','e4','lead',1,2,datetime('now'),'ended',1,'source',datetime('now'),datetime('now'))")
	if err != nil { t.Fatal(err) }
	s := &StatisticsService{DB: db}
	snap, err := s.Freeze(ctx, "m", 2024, "u")
	if err != nil { t.Fatal(err) }
	if snap.Promoted != 1 { t.Fatalf("ended employment missing from promotion count: %d", snap.Promoted) }
	_ = domain.StatisticSnapshot{}
}
