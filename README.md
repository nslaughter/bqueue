# bqueue

blocking queue provides ordering for requests and can provide buffering to synchronize consumption and supply levels for a stage.

## Usage

Create pull based flow with dynamic batch size by using blocking methods `Take(int)` and `Poll(context.Context, int)`. 

## Roadmap

Configuration and bounding for queues. Right now queues are unbouded.

## API

### Method set

```
type BQueue[T any] interface
  Put(T) 
  Offer(T, time.Duration) error
  Take(int) []T
  Poll(int, time.Duration)
```

## Planned Variations

### Priority Blocking Queue

Queues are usually FIFO, but we can add additional criteria and keep them sorted. This is good for work scheduling that needs to depend on more than arrival order.

### Blocking Deque

This supports work stealing algorithms that need to take items from the back. 

### Synchronous Queue

This is important as a concept for the related work of supporting producer/consumer designs, but it doesn't have storage capacity as a work queue. So put/take operations simply block if queue is full/empty. This will likely not be implemented in this package.

