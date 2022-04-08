// Package dq provides a demand queue which has useful properties for building
// pull based systems. 
package dq

type Item struct{}

type Queue struct {
	s chan *state
}

type waiter struct {
	n int
	c chan []Item
}

type state struct {
	items []Item
	wait  []waiter
}

func (s *state) popN(n int) []Item {
	items := s.items[:n:n]
	s.items = s.items[n:]
	return items
}

func New() *Queue {
	s := make(chan *state, 1)
	s <- &state{}
	return &Queue{s}
}

// GetMany gets
func (b *Queue) GetMany(n int) []Item {
	s := <-b.s
	if len(s.wait) == 0 && len(s.items) >= n {
		items := s.popN(n)
		b.s <- s
		return items
	}

	c := make(chan []Item)
	s.wait = append(s.wait, waiter{n, c})
	b.s <- s

	return <-c
}

// Put enqueues an item.
func (b *Queue) Put(item Item) {
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
func (b *Queue) Stop() {
	s := <-b.s
	b.flush(s)
	for _, w := range s.wait {
		close(w.c)
	}
}

// flush sends what's enqueued to the current waiter, regardless
// of the demand size
func (b *Queue) flush(s *state) {
	w := s.wait[0]
	if len(s.items) == 0 {
		return
	}
	w.c <- s.items[:]
	s.wait = s.wait[1:]
}

// Flush sends enqueued items to the current waiter, regardless
// of not satisfying the full demand.
func (b *Queue) Flush() {
	s := <-b.s
	b.flush(s)
	b.s <- s
}
