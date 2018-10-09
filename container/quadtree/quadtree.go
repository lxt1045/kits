package quadtree

import (
	"math"
	"unsafe"

	"github.com/lxt1045/kits/log"
)

/*
 *	需要注意：
 *		1、可以存放相同key的数据
 *		2、Renew()函数是必须的
 */
const (
	defaultMaxLevel    int     = 20         //18 //defaultMaxLevel level最深层次
	defaultProbability float64 = 1 / math.E //defaultProbability 即是p: 1.0/3.0 ; 1.0/4.0 ; 1/math.E
)

//Node 四叉树的节点
//叶子节点 将son[0]指针 强制转换成 *LeafNode
type Node struct {
	son [4]*Node
}

//allSons 返回所有旗下的叶子节点depth表示当前节点到叶子节点的深度
func (p *Node) allSons2(qs *[]*Node, depth uint64) {
	if p == nil {
		return
	}
	depth--
	if depth == 0 {
		if p.son[0] != nil {
			*qs = append(*qs, p.son[0])
		}
		if p.son[1] != nil {
			*qs = append(*qs, p.son[1])
		}
		if p.son[2] != nil {
			*qs = append(*qs, p.son[2])
		}
		if p.son[3] != nil {
			*qs = append(*qs, p.son[3])
		}
		return
	}
	if p.son[0] != nil {
		p.son[0].allSons2(qs, depth)
	}
	if p.son[1] != nil {
		p.son[1].allSons2(qs, depth)
	}
	if p.son[2] != nil {
		p.son[2].allSons2(qs, depth)
	}
	if p.son[3] != nil {
		p.son[3].allSons2(qs, depth)
	}

	return
}

//allSons 返回所有旗下的叶子节点depth表示当前节点到叶子节点的深度
func (p *Node) allSons(qs []*Node, depth uint64) []*Node {
	if p == nil {
		return qs
	}
	depth--
	if depth == 0 {
		if p.son[0] != nil {
			qs = append(qs, p.son[0])
		}
		if p.son[1] != nil {
			qs = append(qs, p.son[1])
		}
		if p.son[2] != nil {
			qs = append(qs, p.son[2])
		}
		if p.son[3] != nil {
			qs = append(qs, p.son[3])
		}
		return qs
	}
	if p.son[0] != nil {
		qs = p.son[0].allSons(qs, depth)
	}
	if p.son[1] != nil {
		qs = p.son[1].allSons(qs, depth)
	}
	if p.son[2] != nil {
		qs = p.son[2].allSons(qs, depth)
	}
	if p.son[3] != nil {
		qs = p.son[3].allSons(qs, depth)
	}

	return qs
}

func (p *Node) find(key uint64, start uint64) (q **Node, end uint64) {
	q = &p
	for x := uint64(0); *q != nil && start > 0; q = &(*q).son[x] {
		start -= 2
		x = (key >> start) & 0x3
		//log.Infof("p:%v, start:%v, x:%v", q, start, x)
	}
	return q, start
}

//返回q **Node 比q *Node慢？！！！！
func (p *Node) readFind(key uint64, start uint64) (q *Node) {
	q = p
	for x := uint64(0); q != nil && start > 0; q = q.son[x] {
		start -= 2
		x = (key >> start) & 0x3
		//log.Infof("p:%v, start:%v, x:%v", q, start, x)
	}
	if start > 0 {
		q = nil
	}
	return
}

//LeafNode 叶子节点
type LeafNode struct {
	key   uint64
	value interface{}
}

//Quadtree 四叉树的节点
type Quadtree struct {
	Node

	n uint32 //当前存储元素的数量
}

func New() *Quadtree {
	return &Quadtree{}
}

func (p *Quadtree) Set(key uint64, val interface{}) (pLeaf *LeafNode) {
	// if p == nil {
	// 	log.Error("get error p == nil")
	// 	return nil
	// }
	//pNode, start := p.writeFind(key, 64) //这个64换成32的话，差不多只要一般的时间
	pNode, start := p.find(key, 64) //这个64换成32的话，差不多只要一般的时间
	if pNode == nil {
		log.Error("get error pNode == nil")
		return
	}
	for start > 0 {
		*pNode = &Node{} //一次性申请内存并没提升: nodes:=make([]Node,start/2)
		start = start - 2
		x := (key >> start) & 0x3
		pNode = &(*pNode).son[x]
	}

	if *pNode == nil {
		pLeaf = &LeafNode{}
		*pNode = (*Node)(unsafe.Pointer(pLeaf))
	} else {
		pLeaf = (*LeafNode)(unsafe.Pointer(*pNode))
	}
	pLeaf.key = key
	pLeaf.value = val
	//log.Infof("pLeaf:%v,p:%x", pLeaf, unsafe.Pointer(pLeaf))
	return
}
func (p *Quadtree) Get(key uint64) *LeafNode {
	pNode := p.readFind(key, 64)
	//if pNode == nil || left != 0 {
	if pNode == nil {
		//log.Error("not find")
		return nil
	}
	return (*LeafNode)(unsafe.Pointer(pNode))
}

//Get len是指查询key前面的多少个字节
func (p *Quadtree) Get2(key, len uint64) (leafs []*LeafNode) {
	key = key >> (64 - len)
	pNode := p.readFind(key, len)
	//if pNode == nil || left != 0 {
	if pNode == nil {
		//log.Error("not find")
		return nil
	}
	if len < 64 {
		var qs []*Node
		pNode.allSons2(&qs, (64-len)/2)
		//qs = pNode.allSons(qs, (64-len)/2)

		for i := range qs {
			leafs = append(leafs, (*LeafNode)(unsafe.Pointer(qs[i])))
		}
	} else {
		leafs = append(leafs, (*LeafNode)(unsafe.Pointer(pNode)))
	}
	return
}

//Remove 删除数据，，，这个比较讲究，如果只删除叶子节点，会造成Q-tree只增不减，最终导致内存膨胀;
//	如果每次删除都删除空的中间节点，则中间的竞态条件处理起来就比较麻烦了，除非使用全树大锁！！！
//	不过也可以采用单线程写，多线程读的方式，则50ms可以写入10ws数据，可以接受，
//	延时可能比较长：200ns channel写 + 500ns Quad-tree写
func (p *Quadtree) Remove(key uint64) (node *LeafNode) {
	if p == nil {
		log.Error("get error p == nil")
		return nil
	}
	pNode, left := p.find(key, 64)
	if pNode == nil {
		log.Error("get error pNode == nil")
	}
	if *pNode == nil || left != 0 {
		return nil
	}
	node = (*LeafNode)(unsafe.Pointer(*pNode))
	*pNode = nil
	return nil
}

//
func (p *Quadtree) Range(keyMin, keyMax uint64, f func(key uint64, value interface{})) {
	// if p == nil {
	// 	log.Error("get error p == nil")
	// 	return nil
	// }
	// pNode, left := p.writeFind(key, 64, false)
	// if pNode == nil {
	// 	log.Error("get error pNode == nil")
	// }
	// if *pNode == nil || left != 0 {
	// 	return nil
	// }
	// node = (*LeafNode)(unsafe.Pointer(*pNode))
	// *pNode = nil
	return
}
