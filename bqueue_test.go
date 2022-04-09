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

	// SUT
	start := time.Now()
	_, err := q.Poll(ctx, 1)
	end := time.Now()

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("deadline")
	}

	if end.Sub(start) < wait {
		t.Fatal("should have waited longer")
	}
}

func TestPollAfterPut(t *testing.T) {
	t.Parallel()
	wait := time.Millisecond * 100
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	q := bqueue.New[testItem]()

	// Put items in queue then unblock chan
	block := make(chan struct{})
	go func() {
		q.Put(testItem{})
		q.Put(testItem{})
		block <- struct{}{}
	}()

	// SUT
	<-block
	start := time.Now()
	items, err := q.Poll(ctx, 2)
	end := time.Now()

	if err != nil {
		t.Fatal("should not err: ", err)
	}

	if end.Sub(start) > time.Millisecond*50 {
		t.Fatal("should have waited longer")
	}

	if len(items) != 2 {
		t.Fatal("expected 2 items")
	}
}

func TestPutAfterPoll(t *testing.T) {
	t.Parallel()
	wait := time.Millisecond * 100
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	q := bqueue.New[testItem]()

	var (
		timeit time.Duration
		items  []testItem
		err    error
	)

	block := make(chan struct{})
	go func() {
		// SUT
		start := time.Now()
		items, err = q.Poll(ctx, 2)
		timeit = time.Now().Sub(start)
		block <- struct{}{}
	}()

	// Put items in queue then unblock chan
	q.Put(testItem{})
	q.Put(testItem{})

	if err != nil {
		t.Fatal("should not err: ", err)
	}

	if timeit > time.Millisecond*50 {
		t.Fatal("should have waited longer")
	}

	// block until Poll returns
	<-block
	if len(items) != 2 {
		t.Fatalf("expected 2 items: got %d", len(items))
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
	time.AfterFunc(time.Millisecond*500, func() {
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
