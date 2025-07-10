package stack

import "github.com/aiden2048/pkg/utils/baselib/container/deque"

// Stack is a LIFO (Last in first out) data structure implementation.
// It is based on a deque container and focuses its API on core
// functionalities: Push, Pop, Head, Size, Empty. Every operations time complexity
// is O(1).
//
// As it is implemented using a Deque container,not safe for concurrent usage.
type Stack struct {
	*deque.Deque
}

func NewStack() *Stack {
	return &Stack{
		Deque: deque.NewDeque(),
	}
}

// Push adds on an item on the top of the Stack
func (s *Stack) Push(item interface{}) {
	s.Prepend(item)
}

// Pop removes and returns the item on the top of the Stack
func (s *Stack) Pop() interface{} {
	return s.Shift()
}

// Head returns the item on the top of the stack
func (s *Stack) Head() interface{} {
	return s.First()
}
