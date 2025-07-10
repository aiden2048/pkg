package baselib

import (
	"container/list"
	"sync"
)

type Pool struct {
	list     *list.List
	dataLock *sync.Mutex
	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() interface{}
}

func NewPool(New func() interface{}) *Pool {
	return &Pool{New: New, list: list.New(), dataLock: &sync.Mutex{}}
}

func (p *Pool) Put(x interface{}) {
	p.dataLock.Lock()
	p.list.PushBack(x)
	p.dataLock.Unlock()
}

func (p *Pool) Get() (x interface{}) {
	p.dataLock.Lock()
	if p.list.Len() > 0 {
		x = p.list.Remove(p.list.Front())
	} else {
		x = p.New()
	}
	p.dataLock.Unlock()
	return x
}
