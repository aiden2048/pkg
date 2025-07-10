package sync_stack

import "github.com/aiden2048/pkg/utils/baselib/container/sync_deque"

// Stack is a LIFO (Last in first out) data structure implementation.
// It is based on a deque container and focuses its API on core
// functionalities: Push, Pop, Head, Size, Empty. Every operations time complexity
// is O(1).
//
// As it is implemented using a Deque container, every operations
// over an instiated Stack are synchronized and safe for concurrent
// usage.
type SyncStack struct {
	*sync_deque.SyncDeque
}

func NewSyncStack() *SyncStack {
	return &SyncStack{
		SyncDeque: sync_deque.NewSyncDeque(),
	}
}

// Push adds on an item on the top of the Stack
func (s *SyncStack) Push(item interface{}) {
	s.Prepend(item)
}

// Pop removes and returns the item on the top of the Stack
func (s *SyncStack) Pop() interface{} {
	return s.Shift()
}

// Head returns the item on the top of the stack
func (s *SyncStack) Head() interface{} {
	return s.First()
}
