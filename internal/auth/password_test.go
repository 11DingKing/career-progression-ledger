package auth

import "testing"

func TestPasswordHash(t *testing.T) {
	h := HashPassword("secret")
	if h == "" || len(h) != 64 {
		t.Fatal("bad hash")
	}
	if !CheckPassword(h, "secret") {
		t.Fatal("match failed")
	}
	if CheckPassword(h, "other") {
		t.Fatal("wrong password accepted")
	}
}
func TestPasswordDeterministic(t *testing.T) {
	if HashPassword("x") != HashPassword("x") {
		t.Fatal("not deterministic")
	}
}
