// Package bqueue provides a blocking queue.
package bqueue

import (
	"context"
)

type Queue[T any] struct {
	s chan *state[T]
}

type waiter[T any] struct {
	ctx context.Context
	n   int
	c   chan []T
}

type state[T any] struct {
	items []T
	wait  []waiter[T]
}

func (s *state[T]) popN(n int) []T {
	items := s.items[:n:n]
	s.items = s.items[n:]
	return items
}

func New[T any]() *Queue[T] {
	s := make(chan *state[T], 1)
	s <- &state[T]{
		items: make([]T, 0),
		wait:  make([]waiter[T], 0),
	}
	return &Queue[T]{s}
}

// Take takes items from the queue when it can satisfy demand.
func (b *Queue[T]) Take(n int) []T {
	ctx := context.Background()

	s := <-b.s
	if len(s.wait) == 0 && len(s.items) >= n {
		items := s.popN(n)
		b.s <- s
		return items
	}

	c := make(chan []T)
	s.wait = append(s.wait, waiter[T]{ctx, n, c})
	b.s <- s

	return <-c
}

func (b *Queue[T]) Poll(ctx context.Context, n int) ([]T, error) {
	// check for cancellation at start
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// just satisfy demand if we have the items
	s := <-b.s
	if len(s.wait) == 0 && len(s.items) >= n {
		items := s.popN(n)
		b.s <- s
		return items, nil
	}

	// add waiter to queue
	c := make(chan []T)
	s.wait = append(s.wait, waiter[T]{ctx, n, c})
	b.s <- s

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case items := <-c:
		return items, nil
	}
}

// Put enqueues an item.
func (b *Queue[T]) Put(item T) {
	s := <-b.s
	s.items = append(s.items, item)
	for len(s.wait) > 0 {
		w := s.wait[0]
		if w.ctx.Err() != nil {
			s.wait = s.wait[1:]
			continue
		}
		if len(s.items) < w.n {
			break
		}
		w.c <- s.popN(w.n)
		s.wait = s.wait[1:]
	}
	b.s <- s
}
