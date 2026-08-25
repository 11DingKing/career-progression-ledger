package metrics

import (
	"sync"
	"testing"
)

func TestCountersSnapshotCopiesMap(t *testing.T) {
	c := New()
	c.ObserveRequest("/x", false)
	c.ObserveRequest("/x", true)
	c.ObserveJob()
	r, f, j, m := c.Snapshot()
	if r != 2 || f != 1 || j != 1 || len(m) != 1 {
		t.Fatalf("%d %d %d %v", r, f, j, m)
	}
	delete(m, "/x")
	_, _, _, again := c.Snapshot()
	if len(again) != 1 {
		t.Fatal("snapshot shares map")
	}
}
func TestCountersConcurrent(t *testing.T) {
	c := New()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 100; n++ {
				c.ObserveRequest("/parallel", n%3 == 0)
				c.ObserveJob()
			}
		}()
	}
	wg.Wait()
	r, f, j, _ := c.Snapshot()
	if r != 2000 || f != 680 || j != 2000 {
		t.Fatalf("%d %d %d", r, f, j)
	}
}
