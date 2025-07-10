package queue

import "github.com/aiden2048/pkg/utils/baselib/container/deque"

// Queue is a FIFO (First in first out) data structure implementation.
// It is based on a deque container and focuses its API on core
// functionalities: Enqueue, Dequeue, Head, Size, Empty. Every operations time complexity
// is O(1).
//
// As it is implemented using a Deque container, not safe for concurrent usage.
type Queue struct {
	*deque.Deque
}

func NewQueue() *Queue {
	return &Queue{
		Deque: deque.NewDeque(),
	}
}

// Enqueue adds an item at the back of the queue
func (q *Queue) Enqueue(item interface{}) {
	q.Prepend(item)
}

// Dequeue removes and returns the front queue item
func (q *Queue) Dequeue() interface{} {
	return q.Pop()
}

// Head returns the front queue item
func (q *Queue) Head() interface{} {
	return q.Last()
}
