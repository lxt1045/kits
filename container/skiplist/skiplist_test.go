package skiplist

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	fmt.Printf("Structure sizes: SkipList is %v, Element is %v bytes\n", unsafe.Sizeof(SkipList{}), unsafe.Sizeof(Element{}))
}

func checkSanity(list *SkipList) error {
	// 分层检查：
	// 1、每层的level必须满足最低要求
	// 2、每层的key必须是排好序的
	for level, v := range list.next {
		if v == nil {
			continue
		}

		currentNode := v
		cnt := 1
		for currentNode.next[level] != nil {
			if !(currentNode.next[level].key >= currentNode.key) {
				return fmt.Errorf("后续的节点的值必须不小于当前节点. [next:%v] [prev:%v]",
					currentNode.next[level].key, currentNode.key)
			}
			if level > len(currentNode.next) {
				return fmt.Errorf("后续的节点的层数必须不小于当前节点. [cur:%v] [node:%v]", level, currentNode.next)
			}
			currentNode = currentNode.next[level] //从当前层向前跳跃一步
			cnt++
		}
		if level == 0 {
			if cnt != list.Length {
				return fmt.Errorf("第一层的元素数量和跳表元素数量不一致. [cur:%v] [level0:%v]", cnt, list.Length)
			}
		}
	}
	return nil
}
func checkValue(list *SkipList, _xs _nodeSlice) error {
	//和skiplist保持一致可存重复值，但是新的值存前面，所以要先反序，在用稳定排序
	l := len(_xs)
	xs := make([]_node, l)
	for i := range _xs {
		xs[l-i-1] = _xs[i]
	}

	sort.Stable(_nodeSlice(xs))
	i, currentNode := 0, list.next[0]
	for ; currentNode != nil && i < len(xs); i, currentNode = i+1, currentNode.next[0] {
		if currentNode.key != xs[i].k {
			return fmt.Errorf("error, currentNode.key:%d != xs[i]:%d", currentNode.key, xs[i].k)
		}
		if v, ok := currentNode.value.(uint64); !ok || v != xs[i].v {
			return fmt.Errorf("error, currentNode.v:%d != vs[i]:%d", v, xs[i].v)
		}
		//从0层向前跳跃一步
	}
	if currentNode != nil || i < len(xs) {
		return fmt.Errorf("error")
	}
	return nil
}

type _node struct {
	k uint64
	v uint64
}
type _nodeSlice []_node

func (s _nodeSlice) Len() int           { return len(s) }
func (s _nodeSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s _nodeSlice) Less(i, j int) bool { return s[i].k < s[j].k }
func (s *_nodeSlice) Set(k, v uint64) {
	// if i := BinSearch((*s), k); i < len((*s)) && (*s)[i].k == k {
	// 	//(*s)[i].v = v
	// 	(*s) = append([]_node{_node{k, v}}, (*s)...) //和skiplist保持一致可存重复值，但是新的值存前面
	// } else {
	// 	(*s) = append((*s), _node{k, v})
	// 	//sort.Sort((*s))
	// }
	(*s) = append((*s), _node{k, v})
}
func (s _nodeSlice) Get(k uint64) (v uint64) {
	sort.Stable(s)
	if i := BinSearch(s, k); i < len(s) && s[i].k == k {
		return s[i].v
	}
	return 0
}
func (s *_nodeSlice) Remove(k uint64) (v uint64) {
	sort.Stable((*s))
	if i := BinSearch(*s, k); i < len(*s) && (*s)[i].k == k {
		v = (*s)[i].v
		(*s) = append((*s)[:i], (*s)[i+1:]...)
		//(*s) = (*s)[:len(*s)-1]
		return
	}
	return 0
}
func BinSearch(a _nodeSlice, x uint64) int {
	return sort.Search(len(a), func(i int) bool { return a[i].k >= x })
}
func TestNewWithMaxLevel(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			if err.(string) != "maxLevel for a SkipList must be a positive integer <= 64" {
				t.Error(err)
			}
		}
	}()
	NewWithMaxLevel(65)
}
func TestBasicIntCRUD(t *testing.T) {
	list := New()
	var xs _nodeSlice
	list.Set(10, uint64(1))
	list.Set(60, uint64(2))
	node30_1 := list.Set(30, uint64(3))
	node20 := list.Set(20, uint64(4))
	list.Set(90, uint64(5))
	xs.Set(10, uint64(1))
	xs.Set(60, uint64(2))
	xs.Set(30, uint64(3))
	xs.Set(20, uint64(4))
	xs.Set(90, uint64(5))

	if err := checkSanity(list); err != nil {
		t.Error(err, "--:", xs)
	}
	if err := checkValue(list, xs); err != nil {
		t.Error(err, "--:", xs)
	}

	node30_2 := list.Set(30, uint64(9))
	xs.Set(30, uint64(9))
	if err := checkSanity(list); err != nil {
		t.Error(err, "--:", xs)
	}
	if err := checkValue(list, xs); err != nil {
		t.Error(err)
	}

	list.Remove(0)
	list.RemoveValue(20, node20.value)
	xs.Remove(0)
	xs.Remove(20)
	if err := checkSanity(list); err != nil {
		t.Error(err, "--:", xs)
	}
	if err := checkValue(list, xs); err != nil {
		t.Error(err)
	}

	v1 := list.Get(10)
	v2 := list.Get(60)
	v3 := list.Get(30)
	v4 := list.Get(20)
	v5 := list.Get(90)
	v6 := list.Get(0)

	if v1 == nil || v1.value.(uint64) != 1 || v1.key != 10 {
		t.Fatal(`wrong "10" value (expected "1")`, v1)
	}

	if v2 == nil || v2.value.(uint64) != 2 {
		t.Fatal(`wrong "60" value (expected "2")`)
	}

	if v3 == nil || v3.value.(uint64) != 9 {
		t.Fatal(`wrong "30" value (expected "9")`)
	}

	if v4 != nil {
		t.Fatal(`found value for key "20", which should have been deleted`)
	}

	if v5 == nil || v5.Value().(uint64) != 5 || v5.Key() != 90 {
		t.Fatal(`wrong "90" value`)
	}

	if v6 != nil {
		t.Fatal(`found value for key "0", which should have been deleted`)
	}

	v3 = list.Get(30)
	if v3 == nil || v3.value.(uint64) != 9 {
		t.Fatal(`wrong "30" value (expected "9")`)
	}
	list.RemoveValue(30, node30_2.value)
	v3 = list.Get(30)
	if v3 == nil || v3.value.(uint64) != 3 {
		t.Fatal(`wrong "30" value (expected "9")`, v3, "\n", xs)
	}
	list.RemoveValue(30, node30_1.value)
	v3 = list.Get(30)
	if v3 != nil {
		t.Fatal(`wrong "30" value (expected "9")`)
	}
}

func TestRenewKey(t *testing.T) {
	list := New()
	list.Set(10, uint64(1))
	list.Set(60, uint64(2))
	node30_1 := list.Set(30, uint64(3))
	node30_2 := list.Set(30, uint64(9))
	//node20 := list.Set(20, uint64(4))
	list.Set(90, uint64(5))

	v3 := list.Get(30)
	if v3 == nil || v3.value.(uint64) != 9 {
		t.Fatal(`wrong "30" value (expected "9")`)
	}
	list.RemoveValue(30, node30_1.value)
	v3 = list.Get(30)
	if v3 == nil || v3.value.(uint64) != 9 {
		t.Fatal(`wrong "30" value (expected "9")`)
	}

	list.RenewKey(35, 30, node30_2.value)
	v3 = list.Get(30)
	if v3 != nil {
		t.Fatal(`wrong "30" value (expected "9")`, v3)
	}
	v3 = list.Get(35)
	if v3 == nil || v3.value.(uint64) != 9 {
		t.Fatal(`wrong "30" value (expected "9")`, v3)
	}
}

func TestChangeLevel(t *testing.T) {
	var i uint64
	list := New()
	var xs _nodeSlice

	if list.maxLevel != defaultMaxLevel {
		t.Fatal("max level must equal default max value")
	}

	list = NewWithMaxLevel(4)
	if list.maxLevel != 4 {
		t.Fatal("wrong maxLevel (wanted 4)", list.maxLevel)
	}

	for i = 1; i <= 201; i++ {
		list.Set(i, uint64(i*10))
		xs.Set(i, uint64(i*10))
	}

	if err := checkSanity(list); err != nil {
		t.Error(err, "--:", xs)
	}
	if err := checkValue(list, xs); err != nil {
		t.Error(err)
	}

	if list.Length != 201 {
		t.Fatal("wrong list length", list.Length)
	}

	for c := list.Head(); c != nil; c = c.Next() {
		if c.key*10 != c.value.(uint64) {
			t.Fatal("wrong list element value")
		}
	}
}

func TestChangeProbability(t *testing.T) {
	list := New()

	if list.probability != defaultProbability {
		t.Fatal("new lists should have P value = DefaultProbability")
	}

	list.SetProbability(0.5)
	if list.probability != 0.5 {
		t.Fatal("failed to set new list probability value: expected 0.5, got", list.probability)
	}
}

func TestConcurrency(t *testing.T) {
	list := New()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 0; i < 100000; i++ {
			list.Set(uint64(i), i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 100000; i++ {
			list.Get(uint64(i))
		}
		wg.Done()
	}()

	wg.Wait()
	if list.Length != 100000 {
		t.Fail()
	}
}

//1/e
//BenchmarkIncSet-4   	 5000000	       280 ns/op	17814379.44 MB/s	      64 B/op	       2 allocs/op
//1/4
//BenchmarkRandSet-4   	 5000000	       335 ns/op	14889379.06 MB/s	      63 B/op	       2 allocs/op
//1/3
//BenchmarkRandSet-4   	 5000000	       276 ns/op	18086993.66 MB/s	      64 B/op	       2 allocs/op
//1/2
//BenchmarkRandSet-4   	 5000000	       285 ns/op	17517081.26 MB/s	      67 B/op	       2 allocs/op
func BenchmarkIncSet(b *testing.B) {
	b.StopTimer()
	benchList := New()
	var xs _nodeSlice
	for i := 0; i <= 100000; i++ {
		v := uint64(rand.Int63())
		benchList.Set(uint64(i), v)
		xs.Set(uint64(i), v)
	}
	inserS := make([]uint64, b.N)
	for i := 0; i < b.N; i++ {
		v := uint64(rand.Int63())
		xs.Set(uint64(i), v)
		inserS[i] = v
	}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		benchList.Set(uint64(i), inserS[i])
	}
	b.StopTimer()
	b.SetBytes(int64(b.N))

	if err := checkSanity(benchList); err != nil {
		b.Error(err, "--:", xs)
	}
	if err := checkValue(benchList, xs); err != nil {
		b.Error(err, "--:", xs)
	}
	b.StartTimer()
	//b.Log(list.probTable)
}

//100:    BenchmarkRandSet-4   	 2000000	       682 ns/op	2932082.49 MB/s	      66 B/op	       3 allocs/op
//1000:   BenchmarkRandSet-4   	 2000000	       940 ns/op	2126617.64 MB/s	      66 B/op	       3 allocs/op
//10000:  1704 ns/op	586847.88 MB/s	      66 B/op	       3 allocs/op
//50000:  1894 ns/op	527725.23 MB/s	      66 B/op	       3 allocs/op
//100000: 2144 ns/op	466322.23 MB/s	      65 B/op	       3 allocs/op
//200000: 2204 ns/op	453568.02 MB/s	      66 B/op	       3 allocs/op
//500000: 2331 ns/op	428879.79 MB/s	      65 B/op	       3 allocs/op
//1000000:2584 ns/op	193459.45 MB/s	      65 B/op	       3 allocs/op
//2000000:2776 ns/op	180087.34 MB/s	      65 B/op	       3 allocs/op
func BenchmarkRandSet(b *testing.B) {
	b.StopTimer()
	benchList := New()
	var xs _nodeSlice
	N := 2000000
	for i := 0; i <= N; i++ {
		v := uint64(rand.Int63())
		benchList.Set(v, v)
		xs.Set(v, v)
	}
	inserS := make([]uint64, N)
	for i := 0; i < N; i++ {
		v := uint64(rand.Int63())
		inserS[i] = v
	}
	for i := 0; i < b.N; i++ {
		xs.Set(inserS[i%N], inserS[i%N])
	}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		benchList.Set(inserS[i%N], inserS[i%N])
	}
	b.StopTimer()
	b.SetBytes(int64(b.N))

	if err := checkSanity(benchList); err != nil {
		b.Error(err, "--:")
	}
	if err := checkValue(benchList, xs); err != nil {
		b.Error(err, "--:")
	}
	b.StartTimer()
	//b.Log(list.probTable)
}

//1000000
//BenchmarkRandGet-4   	 1000000	      2244 ns/op	445490.54 MB/s	       0 B/op	       0 allocs/op
//BenchmarkRandGet-4   	 1000000	      2218 ns/op	450711.09 MB/s	       0 B/op	       0 allocs/op
//100000
//BenchmarkRandGet-4   	 2000000	       724 ns/op	2760198.76 MB/s	       0 B/op	       0 allocs/op
//BenchmarkRandGet-4   	 2000000	       817 ns/op	2447046.23 MB/s	       0 B/op	       0 allocs/op
//10000
//BenchmarkRandGet-4   	 5000000	       268 ns/op	18597862.51 MB/s	       0 B/op	       0 allocs/op
//1000
//BenchmarkRandGet-4   	20000000	       114 ns/op	174162906.67 MB/s	       0 B/op	       0 allocs/op
//100
//BenchmarkRandGet-4   	20000000	       101 ns/op	196672749.43 MB/s	       0 B/op	       0 allocs/op
func BenchmarkRandGet(b *testing.B) {
	b.StopTimer()
	benchList := New()
	var xs _nodeSlice
	N := 1000
	for i := 0; i <= N; i++ {
		v := uint64(rand.Int63())
		benchList.Set(v, v)
		xs.Set(v, v)
	}
	//100000, 排序：
	//BenchmarkRandGet-4   	 5000000	       215 ns/op	23161373.72 MB/s	       0 B/op	       0 allocs/op
	//100000, 不排序：(原因是高速缓存命中率差吗？)
	//BenchmarkRandGet-4   	 2000000	       686 ns/op	2912839.68 MB/s	       0 B/op	       0 allocs/op
	//sort.Sort(xs)
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		j := i % N
		v := benchList.Get(xs[j].k)
		if v == nil || v.value.(uint64) != xs[j].v {
			b.Error("---:", xs[j].v)
		}
	}
	b.SetBytes(int64(b.N))
}

//BenchmarkNextNodes-4   	10000000	       149 ns/op	66749280.03 MB/s	       0 B/op	       0 allocs/op
func BenchmarkNextNodes(b *testing.B) {
	b.StopTimer()
	benchList := New()
	var xs _nodeSlice
	N := 100000
	for i := 0; i <= N; i++ {
		v := uint64(rand.Int63())
		benchList.Set(v, v)
		xs.Set(v, v)
	}
	//排序：
	//BenchmarkRandGet-4   	 5000000	       215 ns/op	23161373.72 MB/s	       0 B/op	       0 allocs/op
	//不排序：(原因是高速缓存命中率差吗？)
	//BenchmarkRandGet-4   	 2000000	       686 ns/op	2912839.68 MB/s	       0 B/op	       0 allocs/op
	//sort.Sort(xs)
	b.ReportAllocs()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		benchList.nextNodes(xs[i%N].k)
	}

	b.SetBytes(int64(b.N))
}
