package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Counters struct {
	Requests atomic.Uint64
	Failures atomic.Uint64
	Jobs     atomic.Uint64
	Mu       sync.RWMutex
	Last     map[string]time.Time
}

func New() *Counters { return &Counters{Last: map[string]time.Time{}} }
func (c *Counters) ObserveRequest(path string, failed bool) {
	c.Requests.Add(1)
	if failed {
		c.Failures.Add(1)
	}
	c.Mu.Lock()
	c.Last[path] = time.Now().UTC()
	c.Mu.Unlock()
}
func (c *Counters) ObserveJob() { c.Jobs.Add(1) }
func (c *Counters) Snapshot() (uint64, uint64, uint64, map[string]time.Time) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	m := map[string]time.Time{}
	for k, v := range c.Last {
		m[k] = v
	}
	return c.Requests.Load(), c.Failures.Load(), c.Jobs.Load(), m
}
