// Package bqueue provides a blocking queue.
package dq

type Queue[T any] struct {
	s chan *state[T]
}

type waiter[T any] struct {
	n int
	c chan []T
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
	s <- &state[T]{}
	return &Queue[T]{s}
}

// Take takes items from the queue. Blocks until
// items are available.
func (b *Queue[T]) Take(n int) []T {
	s := <-b.s
	if len(s.wait) == 0 && len(s.items) >= n {
		items := s.popN(n)
		b.s <- s
		return items
	}

	c := make(chan []T)
	s.wait = append(s.wait, waiter[T]{n, c})
	b.s <- s

	return <-c
}

// Put enqueues an item.
func (b *Queue[T]) Put(item T) {
	s := <-b.s
	s.items = append(s.items, item)
	for len(s.wait) > 0 {
		w := s.wait[0]
		if len(s.items) < w.n {
			break
		}
		w.c <- s.popN(w.n)
		s.wait = s.wait[1:]
	}
	b.s <- s
}

// Stop flushes what's buffered and closes the queue.
func (b *Queue[T]) Stop() {
	s := <-b.s
	b.flush(s)
	for _, w := range s.wait {
		close(w.c)
	}
}

// flush sends what's enqueued to the current waiter, regardless
// of the demand size
func (b *Queue[T]) flush(s *state[T]) {
	w := s.wait[0]
	if len(s.items) == 0 {
		return
	}
	w.c <- s.items
	s.wait = s.wait[1:]
}

// Flush sends enqueued items to the current waiter, regardless
// of not satisfying the full demand.
func (b *Queue[T]) Flush() {
	s := <-b.s
	b.flush(s)
	b.s <- s
}
