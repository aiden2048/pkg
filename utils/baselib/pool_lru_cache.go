package baselib

import (
	"container/list"
	"errors"
	"sync"
)

// LRU Cache
/*
   @原理：
       map 存储数据 , list 存储访问顺序，经常访问的放在 list 的头部，如果超过容量了就从 list 的尾部淘汰
*/

// LRUCache 的内部节点结构
type CacheNode struct {
	Key, Value interface{}
}

func NewCacheNode(k, v interface{}) *CacheNode {
	return &CacheNode{k, v}
}

// LRUCache 结构
type LRUCache struct {
	mu       *sync.Mutex //实现并发安全
	Capacity int
	dlist    *list.List
	cacheMap map[interface{}]*list.Element
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		mu:       &sync.Mutex{},
		Capacity: cap,
		dlist:    list.New(),
		cacheMap: make(map[interface{}]*list.Element)}
}

func (lru *LRUCache) Size() int {
	//add>>>>>>>>>>>>>
	lru.mu.Lock()
	defer func() {
		lru.mu.Unlock()
	}()
	//<<<<<<<<<<<<<add
	return lru.dlist.Len()
}

func (lru *LRUCache) Store(k, v interface{}) error {

	//add>>>>>>>>>>>>>
	lru.mu.Lock()
	defer func() {
		lru.mu.Unlock()
	}()
	//<<<<<<<<<<<<<add
	if lru.dlist == nil {
		return errors.New("LRUCache don't initialize.")
	}

	if pElement, ok := lru.cacheMap[k]; ok {
		lru.dlist.MoveToFront(pElement)
		pElement.Value.(*CacheNode).Value = v
		return nil
	}

	newElement := lru.dlist.PushFront(&CacheNode{k, v})
	lru.cacheMap[k] = newElement

	if lru.dlist.Len() > lru.Capacity {
		// 移掉最后一个
		lastElement := lru.dlist.Back()
		if lastElement == nil {
			return nil
		}
		cacheNode := lastElement.Value.(*CacheNode)
		delete(lru.cacheMap, cacheNode.Key)
		lru.dlist.Remove(lastElement)
	}
	return nil
}

func (lru *LRUCache) Load(k interface{}) (v interface{}, ret bool, err error) {

	//add>>>>>>>>>>>>>
	lru.mu.Lock()
	defer func() {
		lru.mu.Unlock()
	}()
	//<<<<<<<<<<<<<add
	if lru.cacheMap == nil {
		return v, false, errors.New("LRUCache don't initialize.")
	}

	if pElement, ok := lru.cacheMap[k]; ok {
		lru.dlist.MoveToFront(pElement)
		return pElement.Value.(*CacheNode).Value, true, nil
	}
	return v, false, nil
}

func (lru *LRUCache) Delete(k interface{}) bool {

	//add>>>>>>>>>>>>>
	lru.mu.Lock()
	defer func() {
		lru.mu.Unlock()
	}()
	//<<<<<<<<<<<<<add
	if lru.cacheMap == nil {
		return false
	}

	if pElement, ok := lru.cacheMap[k]; ok {
		cacheNode := pElement.Value.(*CacheNode)
		delete(lru.cacheMap, cacheNode.Key)
		lru.dlist.Remove(pElement)
		return true
	}
	return false
}
