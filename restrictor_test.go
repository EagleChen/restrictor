package restrictor

import (
	"testing"
	"time"
)

const d = 100 * time.Millisecond

var (
	t0  = time.Unix(1600530888, 0)
	t1  = t0.Add(time.Duration(1) * d)
	t2  = t0.Add(time.Duration(2) * d)
	t3  = t0.Add(time.Duration(3) * d)
	t4  = t0.Add(time.Duration(4) * d)
	t5  = t0.Add(time.Duration(5) * d)
	t10 = t0.Add(time.Duration(10) * d) // add 1 seconds
	t12 = t0.Add(time.Duration(12) * d) // add 1.2 seconds
	t20 = t0.Add(time.Duration(20) * d) // add 2 seconds
)

func TestRestrictorLimit(t *testing.T) {
	store := createMemoryStore(t)
	window := 2 * time.Second
	limit := 5
	key := "123"
	r := NewRestrictor(window, uint32(limit), 2, store)

	for i := 0; i < limit; i++ {
		if reached, _ := r.LimitReached(key); reached {
			t.Errorf("limit should not have been reached at step %d", i)
		}
	}

	if reached, _ := r.LimitReached(key); !reached {
		t.Error("limit has been reached but not reported so")
	}

}

func TestRestrictorResetWindow(t *testing.T) {
	store := createMemoryStore(t)
	key := "123"
	r := NewRestrictor(2*time.Second, 5, 2, store)

	for i, ts := range []time.Time{t0, t1, t2, t3, t4} {
		if reached, _ := r.LimitReachedAtTime(ts, key); reached {
			t.Errorf("limit should not have been reached at step %d", i)
		}
	}

	for _, ts := range []time.Time{t5, t10, t12} {
		if reached, _ := r.LimitReachedAtTime(ts, key); !reached {
			t.Error("limit has been reached but not reported so")
		}
	}

	// reset
	if reached, _ := r.LimitReachedAtTime(t20, key); reached {
		t.Error("limitor reset, but reported 'reached'")
	}
}

func TestRestrictorWithVariousKeys(t *testing.T) {
	store := createMemoryStore(t)
	limit := 5
	r := NewRestrictor(2*time.Second, uint32(limit), 2, store)

	key1 := "123"
	for i := 0; i < limit; i++ {
		r.LimitReached(key1)
	}
	if reached, _ := r.LimitReached(key1); !reached {
		t.Error("limit has been reached but not reported so")
	}

	key2 := "456"
	for i := 0; i < limit; i++ {
		if reached, _ := r.LimitReached(key2); reached {
			t.Errorf("limit should not have been reached at step %d", i)
		}
	}
}

func TestRestrictorThreadSafe(t *testing.T) {
	store := createMemoryStore(t)
	limit := 50
	r := NewRestrictor(2*time.Second, uint32(limit), 2, store)
	reqs := []int{8, 10, 15, 6, 11}

	key := "123"
	done := make(chan bool)
	// use several goroutines to call 'LimitReached'
	for _, reqCount := range reqs {
		go func(reqCount int) {
			for i := 0; i < reqCount; i++ {
				if reached, _ := r.LimitReached(key); reached {
					t.Error("limit should not have been reached")
				}
			}
			done <- true
		}(reqCount)
	}

	key2 := "234"
	if reached, _ := r.LimitReached(key2); reached {
		t.Error("limit should not have been reached")
	}

	for i := 0; i < len(reqs); i++ {
		<-done
	}

	if reached, _ := r.LimitReached(key); !reached {
		t.Error("limit has been reached but not reported so")
	}
}

func createMemoryStore(t *testing.T) Store {
	t.Helper()
	store, err := NewMemoryStore()
	if err != nil {
		t.Error("fail to create memory store")
	}

	return store
}
