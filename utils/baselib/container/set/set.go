package set

// 提供了一个集合的接口，支持集合的增删改查，也支持两个集合间的交、并、差等操作
// 该版本不支持并发安全
type Set interface {
	Add(e interface{}) bool
	Remove(e interface{})
	Clear()
	IfContains(e interface{}) bool
	Len() int
	Same(other Set) bool
	Elements() []interface{}
	String() string
}

// 判断集合 one 是否是集合 other 的超集
func IsSuperset(one Set, other Set) bool {
	if one == nil || other == nil {
		return false
	}
	oneLen := one.Len()
	otherLen := other.Len()
	if oneLen == 0 || oneLen == otherLen {
		return false
	}
	if oneLen > 0 && otherLen == 0 {
		return true
	}
	for _, v := range other.Elements() {
		if !one.IfContains(v) {
			return false
		}
	}
	return true
}

// 生成集合 one 和集合 other 的并集
func Union(one Set, other Set) Set {
	if one == nil || other == nil {
		return nil
	}
	unionedSet := NewSimpleSet()
	for _, v := range one.Elements() {
		unionedSet.Add(v)
	}
	if other.Len() == 0 {
		return unionedSet
	}
	for _, v := range other.Elements() {
		unionedSet.Add(v)
	}
	return unionedSet
}

// 生成集合 one 和集合 other 的交集
func Intersect(one Set, other Set) Set {
	if one == nil || other == nil {
		return nil
	}
	intersectedSet := NewSimpleSet()
	if other.Len() == 0 {
		return intersectedSet
	}
	if one.Len() < other.Len() {
		for _, v := range one.Elements() {
			if other.IfContains(v) {
				intersectedSet.Add(v)
			}
		}
	} else {
		for _, v := range other.Elements() {
			if one.IfContains(v) {
				intersectedSet.Add(v)
			}
		}
	}
	return intersectedSet
}

// 生成集合 one 对集合 other 的差集
func Difference(one Set, other Set) Set {
	if one == nil || other == nil {
		return nil
	}
	differencedSet := NewSimpleSet()
	if other.Len() == 0 {
		for _, v := range one.Elements() {
			differencedSet.Add(v)
		}
		return differencedSet
	}
	for _, v := range one.Elements() {
		if !other.IfContains(v) {
			differencedSet.Add(v)
		}
	}
	return differencedSet
}

// 生成集合 one 和集合 other 的对称差集
func SymmetricDifference(one Set, other Set) Set {
	if one == nil || other == nil {
		return nil
	}
	diffA := Difference(one, other)
	if other.Len() == 0 {
		return diffA
	}
	diffB := Difference(other, one)
	return Union(diffA, diffB)
}

// 新建一个set对象，内部调用的是NewHashSet
func NewSimpleSet() Set {
	return NewHashSet()
}

// 判断某个value是否符合set接口形式
func IsSet(value interface{}) bool {
	if _, ok := value.(Set); ok {
		return true
	}
	return false
}
