package skiplist

import (
	"math/rand"
	"sync"
)

type Node struct {
	next []*Element
}

type Element struct {
	Node
	key   uint64
	value interface{}
}

// Key allows retrieval of the key for a given Element
func (e *Element) Key() uint64 {
	return e.key
}

// Value allows retrieval of the value for a given Element
func (e *Element) Value() interface{} {
	return e.value
}

//Next 返回前向迭代器
func (e *Element) Next() *Element {
	return e.next[0]
}

//SkipList 表示一个跳表
type SkipList struct {
	Node                    //前向指针数组，根据element的层数生成大小
	maxLevel    int         //跳表的最大层数
	Length      int         //跳表已经存下的数据量
	randSource  rand.Source //
	probability float64
	probTable   []float64
	mutex       sync.RWMutex
	nodesCache  []*Node
}
