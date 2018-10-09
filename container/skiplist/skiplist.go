package skiplist

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultMaxLevel    int     = 20         //18 //defaultMaxLevel level最深层次
	defaultProbability float64 = 1 / math.E //defaultProbability 即是p: 1.0/3.0 ; 1.0/4.0 ; 1/math.E
)

// Head 返回第一个Element
func (list *SkipList) Head() *Element {
	return list.next[0]
}

//RenewKey 主要是刷新key值，所以先写入新的节点，再删除旧的节点，返回旧的节点
func (list *SkipList) RenewKey(newkey, oldkey uint64, value interface{}) *Element {
	// list.mutex.Lock()
	// defer list.mutex.Unlock()
	list.Set(newkey, value)
	list.RemoveValue(oldkey, value)
	return nil
}

//Set 注意可以存重复值
func (list *SkipList) Set(key uint64, value interface{}) *Element {
	// list.mutex.Lock()
	// defer list.mutex.Unlock()

	var element *Element
	forwards := list.nextNodes(key)
	// // 可以存相同的key
	// if element = forwards[0].next[0]; element != nil && element.key <= key {
	// 	element.value = value
	// 	return element
	// }

	element = &Element{
		Node: Node{
			next: make([]*Element, list.randLevel()),
		},
		key:   key,
		value: value,
	}

	for i := range element.next {
		element.next[i] = forwards[i].next[i]
		forwards[i].next[i] = element
	}

	list.Length++
	return element
}
func (list *SkipList) RangeFilter(start, end uint64, filter func(I interface{}) bool) (err error) {
	var forward = &list.Node
	var next *Element

	for i := list.maxLevel - 1; i >= 0; i-- {
		next = forward.next[i]
		for next != nil && start > next.key {
			forward = &next.Node
			next = forward.next[i]
		}
	}

	//for ; next != nil && next.key <= end; next = next.next[0] {
	for ; next != nil && next.key < end; next = next.next[0] {
		if !filter(next.value) {
			return //filter 返回false时退出循环
		}
	}
	return
}
func (list *SkipList) RangeGet(start, end uint64) (es []interface{}) {
	var forward = &list.Node
	var next *Element

	for i := list.maxLevel - 1; i >= 0; i-- {
		next = forward.next[i]
		for next != nil && start > next.key {
			forward = &next.Node
			next = forward.next[i]
		}
	}

	//for ; next != nil && next.key <= end; next = next.next[0] {
	for ; next != nil && next.key < end; next = next.next[0] {
		es = append(es, next.value)
	}
	return
}
func (list *SkipList) Get(key uint64) *Element {
	// list.mutex.RLock()
	// defer list.mutex.RUnlock()

	//return list.rangeGet(key, key)

	var forward = &list.Node
	var next *Element

	for i := list.maxLevel - 1; i >= 0; i-- {
		next = forward.next[i]
		for next != nil && key > next.key {
			forward = &next.Node
			next = forward.next[i]
		}
	}

	if next != nil && next.key <= key {
		return next
	}
	return nil
}

func (list *SkipList) Get2(key, len uint64) (es []interface{}) {
	mask := uint64(1<<(64-len)) - 1
	start, end := key&(^mask), (key+mask)&(^mask)
	//fmt.Printf("mask:%x, start:%x, end:%x, key:%x, len:%d\n", mask, start, end, key, len)

	return list.RangeGet(start, end)
}

func (list *SkipList) Remove(key uint64) *Element {
	// list.mutex.Lock()
	// defer list.mutex.Unlock()
	forwards := list.nextNodes(key)

	// found the element, remove it
	for e := forwards[0].next[0]; e != nil && e.key <= key; e = e.next[0] {
		for k, v := range e.next {
			forwards[k].next[k] = v
		}

		list.Length--
		return e
	}

	return nil
}
func (list *SkipList) RemoveValue(key uint64, value interface{}) *Element {
	// list.mutex.Lock()
	// defer list.mutex.Unlock()
	forwards := list.nextNodes(key)

	// found the element, remove it
	for e := forwards[0].next[0]; e != nil && e.key <= key; e = e.next[0] {
		if e.value == value { //指针的interface{}是可以比较的
			for k, v := range e.next {
				forwards[k].next[k] = v
			}

			list.Length--
			return e
		}
		//forwards 也要修改
		for i := range e.next {
			//forwards[i] = &forwards[0].next[0].Node
			forwards[i] = &e.Node
		}
	}

	return nil
}

func (list *SkipList) nextNodes(key uint64) []*Node {
	forwards := list.nodesCache

	forward := &list.Node //类型为*Node
	for level := list.maxLevel - 1; level >= 0; level-- {
		for forward.next[level] != nil && key > forward.next[level].key {
			forward = &forward.next[level].Node
		}
		forwards[level] = forward
	}

	return forwards
}

// SetProbability changes the current P value of the list.
// It doesn't alter any existing data, only changes how future insert heights are calculated.
func (list *SkipList) SetProbability(newProbability float64) {
	list.probability = newProbability
	list.probTable = probabilityTable(list.probability, list.maxLevel)
}

func (list *SkipList) randLevel() (level int) {
	r := float64(list.randSource.Int63()) / (1 << 63) //获取[0,1)区间内的平均分布

	level = 1
	for level < list.maxLevel && r < list.probTable[level] {
		level++
	}
	return
	//return algorithms.BinSearchFloat64(list.probTable, r)//二分查找不见得比直接遍历快，因为高level是小概率事件,期望是1/p=e
}

//probabilityTable 为了减少计算随机数次数，用表存下层数对应的随机值
func probabilityTable(probability float64, MaxLevel int) (table []float64) {
	for i := 1; i <= MaxLevel; i++ {
		prob := math.Pow(probability, float64(i))
		table = append(table, prob)
	}
	return table
}

// NewWithMaxLevel creates a new skip list with MaxLevel set to the provided number.
// Returns a pointer to the new list.
func NewWithMaxLevel(maxLevel int) *SkipList {
	if maxLevel < 1 || maxLevel > 64 {
		panic("maxLevel for a SkipList must be a positive integer <= 64")
	}

	return &SkipList{
		Node:        Node{next: make([]*Element, maxLevel)},
		nodesCache:  make([]*Node, maxLevel),
		maxLevel:    maxLevel,
		randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
		probability: defaultProbability,
		probTable:   probabilityTable(defaultProbability, maxLevel),
	}
}

// New creates a new skip list with default parameters. Returns a pointer to the new list.
func New() *SkipList {
	return NewWithMaxLevel(defaultMaxLevel)
}
