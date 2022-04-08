package dq_test 

import (
    "runtime"
	"sync"
	"testing"
	"time"

	"github.com/nslaughter/dq"
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func TestQueue(t *testing.T) {
    t.Log("starting goroutines: ", runtime.NumGoroutine())
	n := 32
	tcs := []int{2, 3, 4, 9, 5, 6, 3}

	b := dq.New()
	t.Log("adding items")
	for i := 0; i < n; i++ {
		b.Put(dq.Item{})
		t.Log("put an item")
	}

	t.Log("making results")
	res := make([]int, len(tcs))

	var wg sync.WaitGroup

	t.Log("starting queue consumers")
	for i := 0; i < len(tcs); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res[i] = len(b.GetMany(tcs[i]))
		}(i)
		t.Log("did sched")
	}

	t.Log("running timer")
	time.AfterFunc(time.Millisecond*100, func() {
		b.Stop()
	})

	wg.Wait()

	bal := n
	for i, n := range tcs {
		if res[i] != min(tcs[i], bal) {
			t.Fatalf("index %d missed: bal %d, input %d, got %d", i, bal, tcs[i], res[i])
		}
		bal -= n
	}
    t.Log("ending goroutines: ", runtime.NumGoroutine())
}
