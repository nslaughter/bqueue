package bqueue_test

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/nslaughter/bqueue"
)

type testItem struct {
	id int
}

func newTestItem(id int) testItem {
	return testItem{id}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Invariants
// When
// 		cumulative # items Put previous
// 		- cumulative # items Take previous
// 		>= current # items Take
// Then
// 		Take returns # items
// Else
// 		Take does not return

func TestSimplePutTake(t *testing.T) {
	t.Parallel()
	in := newTestItem(99)
	q := bqueue.New[testItem]()

	q.Put(in)
	out := q.Take(1)[0]

	if out.id != in.id {
		t.Fatal("expected out to be same as in")
	}
}

// When
// 		cumulative # items Put previous
// 		- cumulative # items Take previous
// 		>= current # items Take
// Then
// 		Poll returns # items
// Else
// 		Poll waits for time.Duration

func TestPollTimeout(t *testing.T) {
	t.Parallel()
	wait := time.Millisecond * 50
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	q := bqueue.New[testItem]()

	start := time.Now()
	_, err := q.Poll(ctx, 1)
	if !errors.Is(err, bqueue.ErrDeadlineExceeded) {
		t.Fatal("currently don't want this test passing")
	}
	end := time.Now()

	if end.Sub(start) < wait {
		t.Fatal("should have waited longer")
	}
}

// When
// 		cumulative # items Put previous
// 		- cumulative # items Take previous
// 		>= current # items Take
// Then
// 		Poll returns # items
// Else
// 		Poll waits for time.Duration

func TestQueue(t *testing.T) {
	t.Parallel()
	t.Log("starting goroutines: ", runtime.NumGoroutine())
	n := 32
	tcs := []int{2, 3, 4, 9, 5, 6, 3}
	b := bqueue.New[testItem]()

	t.Log("adding items")
	for i := 0; i < n; i++ {
		b.Put(testItem{})
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
			res[i] = len(b.Take(tcs[i]))
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
