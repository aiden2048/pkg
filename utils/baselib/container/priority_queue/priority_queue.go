package priority_queue

import "github.com/aiden2048/pkg/utils/baselib/container/skiplist"

type PriorityQueue struct {
	SL *skiplist.SkipList
}

func NewPriorityQueue(comparePriority func(interface{}, interface{}) int) *PriorityQueue {
	return &PriorityQueue{SL: skiplist.NewSkipList(comparePriority)}
}

func (p *PriorityQueue) Size() int {
	return p.SL.Num
}

func (p *PriorityQueue) Empty() bool {
	return p.SL.Num == 0
}

func (p *PriorityQueue) Top() *skiplist.Node {
	if !p.Empty() {
		return p.SL.Header.Next[0]
	} else {
		return nil
	}
}

func (p *PriorityQueue) Pop() {
	node := p.Top()
	p.SL.RemoveNode(node)
}

func (p *PriorityQueue) Push(Data skiplist.DataType) {
	p.SL.Insert(Data)
}
