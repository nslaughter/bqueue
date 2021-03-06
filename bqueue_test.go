package bqueue_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nslaughter/bqueue"
)

const (
	msgUnexpectedWait = "unexpected wait: "
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
		t.Fatal(msgUnexpectedWait, end.Sub(start))
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
		q.Put(newTestItem(1))
		q.Put(newTestItem(2))
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

	if len(items) != 2 {
		t.Fatal("expected 2 items")
	}
	if end.Sub(start) > wait {
		t.Fatal(msgUnexpectedWait, end.Sub(start))
	}
}

func TestPutAfterPoll(t *testing.T) {
	t.Parallel()
	wait := time.Millisecond * 100
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	q := bqueue.New[testItem]()

	results := make(chan struct {
		is  []testItem
		d   time.Duration
		err error
	})
	go func() {
		// SUT
		start := time.Now()
		items, err := q.Poll(ctx, 2)
		d := time.Now().Sub(start)
		results <- struct {
			is  []testItem
			d   time.Duration
			err error
		}{items, d, err}
	}()

	// Put items in queue then unblock chan
	q.Put(newTestItem(1))
	q.Put(newTestItem(2))

	res := <-results
	if res.err != nil {
		t.Fatal("should not err: ", res.err)
	}

	if res.d > wait {
		t.Fatal(msgUnexpectedWait, wait)
	}

	// block until Poll returns
	if len(res.is) != 2 {
		t.Fatalf("expected 2 items: got %d", len(res.is))
	}
}

func TestTakeFirst(t *testing.T) {
	t.Parallel()
	b := bqueue.New[testItem]()
	resC := make(chan []testItem)
	wait := make(chan struct{})

	// Take
	go func() {
		wait <- struct{}{}
		resC <- b.Take(1)
	}()

	// Put
	go func() {
		<-wait
		b.Put(newTestItem(1))
	}()

	res := <-resC
	if len(res) != 1 {
		t.Fatal("expected 1 testItem")
	}
}
