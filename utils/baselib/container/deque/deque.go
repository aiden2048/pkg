package deque

import (
	"container/list"
)

// Deque is a head-tail linked list data structure implementation.
// It is based on a doubly linked list container, so that every
// operations time complexity is O(1).
//
// not safe for concurrent usage.
type Deque struct {
	container *list.List
	capacity  int
}

// NewDeque creates a Deque.
/*
	@desc: 默认的是生成一个不限容量的deque
	@params:
	@returns:
		string: 返回的deque指针
*/
func NewDeque() *Deque {
	return NewCappedDeque(-1)
}

// NewCappedDeque creates a Deque with the specified capacity limit.
/*
	@desc: 生成一个指定容量的deque
	@params:
		capacity: 指定deque的容量限制
	@returns:
		string: 返回的deque指针
*/
func NewCappedDeque(capacity int) *Deque {
	return &Deque{
		container: list.New(),
		capacity:  capacity,
	}
}

// Append inserts element at the back of the Deque in a O(1) time complexity,
// returning true if successful or false if the deque is at capacity.
/*
	@desc: 在deque尾端插入一个元素
	@params:
		item: 待插入的元素
	@returns:
		bool: 插入失败（超过deque指定容量），返回false;插入成功，返回true
*/
func (s *Deque) Append(item interface{}) bool {
	if s.capacity < 0 || s.container.Len() < s.capacity {
		s.container.PushBack(item)
		return true
	}

	return false
}

// Prepend inserts element at the Deques front in a O(1) time complexity,
// returning true if successful or false if the deque is at capacity.
/*
	@desc: 在deque头部插入一个元素
	@params:
		item: 待插入的元素
	@returns:
		bool: 插入失败（超过deque指定容量），返回false;插入成功，返回true
*/
func (s *Deque) Prepend(item interface{}) bool {
	if s.capacity < 0 || s.container.Len() < s.capacity {
		s.container.PushFront(item)
		return true
	}

	return false
}

// Pop removes the last element of the deque in a O(1) time complexity
/*
	@desc: 在deque尾部移除一个元素
	@params:
		item: 待移除的元素
	@returns:
		interface{}: 当尾部元素不为空的时候，返回尾部元素；尾部元素为空时，返回空
*/
func (s *Deque) Pop() interface{} {
	var item interface{} = nil
	var lastContainerItem *list.Element = nil

	lastContainerItem = s.container.Back()
	if lastContainerItem != nil {
		item = s.container.Remove(lastContainerItem)
	}

	return item
}

// Shift removes the first element of the deque in a O(1) time complexity
/*
	@desc: 在deque头部移除一个元素
	@params:
		item: 待移除的元素
	@returns:
		interface{}: 当头部元素不为空的时候，返回头部元素；头部元素为空时，返回空
*/
func (s *Deque) Shift() interface{} {
	var item interface{} = nil
	var firstContainerItem *list.Element = nil

	firstContainerItem = s.container.Front()
	if firstContainerItem != nil {
		item = s.container.Remove(firstContainerItem)
	}

	return item
}

// First returns the first value stored in the deque in a O(1) time complexity
/*
	@desc: 返回deque的首元素
	@params:
	@returns:
		interface{}: deque的首元素不为空时，返回首元素；否则返回空
*/
func (s *Deque) First() interface{} {
	item := s.container.Front()
	if item != nil {
		return item.Value
	} else {
		return nil
	}
}

// Last returns the last value stored in the deque in a O(1) time complexity
/*
	@desc: 返回deque的尾元素
	@params:
	@returns:
		interface{}: deque的尾元素不为空时，返回尾元素；否则返回空
*/
func (s *Deque) Last() interface{} {
	item := s.container.Back()
	if item != nil {
		return item.Value
	} else {
		return nil
	}
}

// Size returns the actual deque size
/*
	@desc: 返回deque的元素个数
	@params:
	@returns:
		int: deque的元素个数
*/
func (s *Deque) Size() int {
	return s.container.Len()
}

// Capacity returns the capacity of the deque, or -1 if unlimited
/*
	@desc: 返回deque的容量大小
	@params:
	@returns:
		int: deque的容量大小
*/
func (s *Deque) Capacity() int {
	return s.capacity
}

// Empty checks if the deque is empty
/*
	@desc: 判断deque是否为空
	@params:
	@returns:
		bool: 空则返回true，否则返回false
*/
func (s *Deque) Empty() bool {
	return s.container.Len() == 0
}

// Full checks if the deque is full
/*
	@desc: 判断deque容量是否用完
	@params:
	@returns:
		bool: full则返回true，否则返回false
*/
func (s *Deque) Full() bool {
	return s.capacity >= 0 && s.container.Len() >= s.capacity
}
