package pagination

import "testing"

func TestParseDefaults(t *testing.T) {
	if p := Parse(0, -1); p.Limit != 25 || p.Offset != 0 {
		t.Fatalf("%+v", p)
	}
}
func TestParseCaps(t *testing.T) {
	if p := Parse(101, 4); p.Limit != 25 || p.Offset != 4 {
		t.Fatalf("%+v", p)
	}
	if p := Parse(50, 2); p.Limit != 50 {
		t.Fatalf("%+v", p)
	}
}
