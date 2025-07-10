package skiplist

import (
	"fmt"
	"math/rand"
)

const SKIPLIST_MAXLEVEL = 32
const SKIPLIST_P = 4

type Node struct {
	Prev    []*Node
	Next    []*Node
	Data  * DataType
}

func NewNode(d * DataType, level int) * Node {
	return &Node{Data:d, Prev: make([]*Node, level), Next: make([]*Node, level)}
}

type DataType struct {
	Key   		interface{}  // 键值，一定是唯一的
	Value 		interface{}  // 其它一些数据 
}

type SkipList struct {
	Compare func (interface{}, interface{}) int 	// 优先级的比较函数
	Header * Node 									// 跳表的头结点 
    Tail   * Node 									// 跳表的尾结点	
	Level int     									// 跳表的层数
	Num   int                                       // 当前跳表已经有的节点个数
}

func NewSkipList(comparePriority func (interface{}, interface{}) int) *SkipList {
	sl := &SkipList{Compare:comparePriority,
					 Level:1,
					 Header:NewNode(nil, SKIPLIST_MAXLEVEL), 
					 Tail:NewNode(nil, SKIPLIST_MAXLEVEL)}
	for i:= 0; i < SKIPLIST_MAXLEVEL; i++ {
		sl.Header.Next[i] = sl.Tail
		sl.Tail.Prev[i] = sl.Header
	}
	return sl
 }

func (s * SkipList)RandomLevel() int {
	level := (int)(rand.Int31()) % SKIPLIST_P

	for level == 0 {
		level += 1
	}

	if level < SKIPLIST_MAXLEVEL {
		return level
	} else {
		return SKIPLIST_MAXLEVEL
	}
}

// 找到大于等于key的第一个元素，例如
// 数组元素为： 1 3 5 7 7 9
// key为：7
// 那么第一个大于等于7的第一个元素就是第一个7
func (s * SkipList)FindFirstGreaterOrEqual(Key interface{}) map[int]*Node {
	target := make(map[int]*Node)
	node := s.Header

	for i:=s.Level-1; i >= 0; i-- {
		for {
			if node.Next[i].Data == nil {
				break
			}
			if s.Compare(node.Next[i].Data.Key, Key) < 0 {
				node = node.Next[i]
			} else {
				break
			}
		}
		target[i] = node.Next[i]
	}

	return target 
}

// 找到大于等于key的最后一个元素，例如
// 数组元素为： 1 3 5 7 7 9
// key为：7
// 那么第一个大于等于7的第一个元素就是第二个7
func (s * SkipList)FindLastGreaterOrEqual(Key interface{}) map[int]*Node {
	target := make(map[int]*Node)
	node := s.Header

	for i:=s.Level-1; i >= 0; i-- {
		for {
			if node.Next[i].Data == nil {
				break
			}
			if s.Compare(node.Next[i].Data.Key, Key) <= 0 {
				node = node.Next[i]
			} else {
				break
			}
		}
		target[i] = node
	}

	return target
}

// 找到小于key的第一个元素，例如
// 数组元素为： 1 3 5 7 9
// key为：7
// 那么第一个小于7的第一个元素就是5
func (s * SkipList)FindLessThan(Key interface{}) map[int]*Node {
	target := make(map[int]*Node)
	node := s.Header

	for i:=s.Level-1; i >= 0; i-- {
		for {
			if node.Next[i].Data == nil {
				break	
			}
			if s.Compare(node.Next[i].Data.Key, Key) < 0 {
				node = node.Next[i]
			} else {
				break
			}
		}
		target[i] = node
	}

	return target 
}

func (s * SkipList)Insert(data DataType) {
	update := s.FindLastGreaterOrEqual(data.Key)
	level := s.RandomLevel()
	fmt.Println("get random level = ", level)
	if level > s.Level {
		for j:= s.Level; j < level; j++ {
			update[j] = s.Header
		}
		s.Level = level
	}
	
	newNode := NewNode(&data, level)
	for i := 0; i < level; i++ {
		newNode.Next[i] = update[i].Next[i]
		newNode.Prev[i] = update[i]
		update[i].Next[i].Prev[i] = newNode
		update[i].Next[i] = newNode
	}
	s.Num++
}

func (s * SkipList)SearchKey(Key interface{}) []*Node {
	var target []*Node
	sl := s.SearchKeyLess(Key)
	node := sl[0]
	for {
		if s.Compare(node.Next[0].Data.Key, Key) == 0 {
			target = append(target, node.Next[0])
			node = node.Next[0]	
		} else {
			break
		}
	}
	
	return target
}

func (s * SkipList)SearchKeyLess(Key interface{}) map[int]*Node {
	target := make(map[int]*Node)
	allLess := s.FindLessThan(Key)
	for i, v := range allLess {
		if v.Next[i].Data == nil {
			continue	
		}
		if s.Compare(v.Next[i].Data.Key, Key) == 0 {
			target[i] = v
		}
	}

	return target
}
	
func (s * SkipList)RemoveNode(node * Node) {
	if node == nil {
		return 
	}
	
	level := len(node.Next)
	
	for i:= 0; i < level; i++ {
		if node.Next[i] == s.Tail &&
		   node.Prev[i] == s.Header {
				s.Level--
		   }
		node.Next[i].Prev[i] = node.Prev[i]
		node.Prev[i].Next[i] = node.Next[i]
	}
	s.Num--
}

func (s * SkipList)RemoveKey(Key interface{}) {
	target := s.SearchKeyLess(Key)
	
	for i, v := range target {
		next := v.Next[i]
		for {
			if next.Data == nil {
				break
			}
			if s.Compare(next.Data.Key, Key) == 0 {
				s.Num--
				next = next.Next[i]
			} else {
				break
			}
		}

		if v == s.Header && next == s.Tail {
				s.Level--
		}

		next.Prev[i] = v
		v.Next[i] = next
	}
}

func (s * SkipList)PrintSkipList() {
	fmt.Println("-----------SkipList start--------------")
	for i:= SKIPLIST_MAXLEVEL - 1; i >= 0; i-- {
		fmt.Println("level:", i)
		node := s.Header.Next[i]
		for {
			if node != nil && node.Data != nil && node.Data.Value != nil {
				fmt.Println(node.Data.Key.(string), " ", node.Data.Value.(int))
				node = node.Next[i]
			} else {
				break
			}
		}
		fmt.Println(" ")
	}
	fmt.Println("-----------SkipList end--------------")
}
