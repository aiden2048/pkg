package set

import (
	"bytes"
	"fmt"
)

// 以map为底层实现的集合类
// 不支持并发安全
type HashSet struct {
	m map[interface{}]bool
}

// 新建一个集合
func NewHashSet() *HashSet {
	return &HashSet{m: make(map[interface{}]bool)}
}

/*
	@desc: 添加元素至集合
	@params:
		e: 待添加的元素
	@returns:
		bool: 元素已存在，返回flase；元素不存在，返回true；
*/
func (set *HashSet) Add(e interface{}) bool {
	//判断元素是否在set中已存在，若不存在，则添加并将键对应的值置为true，若已存在，则返回false
	if !set.m[e] {
		set.m[e] = true
		return true
	}
	return false
}

/*
	@desc: 从集合中删除元素
	@params:
		e: 待删除的元素
	@returns:
*/
func (set *HashSet) Remove(e interface{}) {
	delete(set.m, e)
}

/*
	@desc: 清除集合中的所有元素
	@params:
	@returns:
*/
func (set *HashSet) Clear() {
	set.m = make(map[interface{}]bool)
}

/*
	@desc: 判断集合中是否存在某个元素
	@params:
		e: 待判断的元素
	@returns:
		bool: 如果集合中存在元素e，则返回true；若不存在，则返回false
*/
func (set *HashSet) IfContains(e interface{}) bool {
	return set.m[e]
}

/*
	@desc: 返回集合的元素个数
	@params:
		e: 待判断的元素
	@returns:
		bool: 如果集合中存在元素e，则返回true；若不存在，则返回false
*/
func (set *HashSet) Len() int {
	return len(set.m)
}

/*
	@desc: 返回集合是否为空
	@params:
	@returns:
		bool: 如果集合为空（含元素个数为0），返回true;不为空，则返回false
*/
func (set *HashSet) Empty() bool {
	for _, val := range set.m {
		if val == true {
			return false
		}
	}
	return true
}

/*
	@desc: 判断集合one与集合other是否完全相同（即包含的元素完全相同）
	@params:
		other： 不支持并发安全的集合接口
	@returns:
		bool: 相同，返回true;不相同，则返回false
*/
func (set *HashSet) Same(other Set) bool {
	if other == nil {
		return false
	}
	if set.Len() != other.Len() {
		return false
	}
	for key := range set.m {
		if !other.IfContains(key) {
			return false
		}
	}
	return true
}

/*
	@desc: 返回集合中的所有元素，并以切片的形式返回
	@params:
	@returns:
		[]interface{}: 集合元素组成的切片
*/
func (set *HashSet) Elements() []interface{} {
	initialLen := len(set.m)
	snapshot := make([]interface{}, initialLen)
	actualLen := 0
	for key := range set.m {
		if actualLen < initialLen {
			snapshot[actualLen] = key
		} else {
			snapshot = append(snapshot, key)
		}
		actualLen++
	}
	if actualLen < initialLen {
		snapshot = snapshot[:actualLen]
	}
	return snapshot
}

/*
	@desc: 将集合转换为字符串，字符串形式为HashSet{A B C}
	@params:
	@returns:
		string: 返回的字符串
*/
//输出字符串形式为“HashSet{A B C}”
func (set *HashSet) String() string {
	var buf bytes.Buffer
	buf.WriteString("HashSet{")
	first := true
	for key := range set.m {
		if first {
			first = false
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString(fmt.Sprintf("%v", key))
	}
	buf.WriteString("}")
	return buf.String()
}
