package quadtree

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//深，广度优先遍历
func quadtreePrint(q *Node, t *testing.T, level int) {
	if q == nil || level <= 2 {
		return
	}
	t.Log(q.son, ":", level)
	for i := 0; i < 4; i++ {
		if q.son[i] == nil {
			continue
		}
		quadtreePrint(q.son[i], t, level-2)
	}
}

//深，广度优先遍历
func quadtreePrint2(q *Node, t *testing.B, level int) {
	if q == nil || level <= 2 {
		return
	}
	t.Log(q.son, ":", level)
	for i := 0; i < 4; i++ {
		if q.son[i] == nil {
			continue
		}
		quadtreePrint2(q.son[i], t, level-2)
	}
}

//
func TestSet0(t *testing.T) {
	q := New()
	x := q.Set(7, 888)
	t.Log(x)
	y := q.Get(7)
	t.Log(y)

	x = q.Set(7, 111)
	t.Log(x)
	y = q.Get(7)
	t.Log(y)

	x = q.Set(7, 888)
	t.Log(x)
	y2 := q.Get(7)
	t.Log("y:", y2)

	//

	//读取空的值

	y3 := q.Get(12)
	t.Log("y:", y3)
}

//
func TestSet(t *testing.T) {
	N := 100000
	storeMap := make(map[uint64]uint64, N*2)
	quadtree := New()

	for i := 0; i < N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		quadtree.Set(k, v)
		storeMap[k] = v
	}
	for k, v := range storeMap {
		v2 := quadtree.Get(k)
		if v2 == nil || v2.value == nil {
			t.Error("v2.value ==nil")
		}
		if x, ok := v2.value.(uint64); !ok || x != v {
			t.Errorf("k:%v, v:%v <--> v2:%v", k, v, x)
		}
	}
}

//以下是sync.map，性能大概比原版map慢5倍
//100:    171 ns/op	58444240.05 MB/s	      32 B/op	       3 allocs/op
//1000:   178 ns/op	56149016.35 MB/s	      32 B/op	       3 allocs/op
//10000:  220 ns/op	45254160.03 MB/s	      32 B/op	       3 allocs/op
//50000:  251 ns/op	19862110.87 MB/s	      33 B/op	       3 allocs/op
//100000: 269 ns/op	18587338.48 MB/s	      34 B/op	       3 allocs/op
//200000: 322 ns/op	15498484.90 MB/s	      36 B/op	       3 allocs/op
//500000: 453 ns/op	6611631.93 MB/s	      56 B/op	       3 allocs/op
//1000000:513 ns/op	3897067.15 MB/s	     105 B/op	       4 allocs/op
//2000000:619 ns/op	8074339.47 MB/s	      91 B/op	       3 allocs/op
//以下是 QuadTree: 100w以下大概相差2倍，因为时间曲线不一样，所以更大数据不好比较！
//100:     173 ns/op	57471690.45 MB/s	       8 B/op	       1 allocs/op
//1000:    187 ns/op	53363857.46 MB/s	       8 B/op	       1 allocs/op
//10000:   243 ns/op	20520250.74 MB/s	       8 B/op	       1 allocs/op
//50000:   379 ns/op	7913180.80 MB/s	      14 B/op	       1 allocs/op
//100000:  435 ns/op	6884671.05 MB/s	      20 B/op	       1 allocs/op
//200000:  524 ns/op	3811726.26 MB/s	      44 B/op	       2 allocs/op
//500000:  796 ns/op	2512437.98 MB/s	      95 B/op	       3 allocs/op
//1000000: 1199 ns/op	833810.27 MB/s	     351 B/op
//2000000: 1899 ns/op	526586.59 MB/s	     678 B/op	      21 allocs/op
//以下是skip-list：20w以下数据量，和QuadTree相差4倍左右
//100:    682 ns/op	2932082.49 MB/s	      66 B/op	       3 allocs/op
//1000:   940 ns/op	2126617.64 MB/s	      66 B/op	       3 allocs/op
//10000:  1704 ns/op	586847.88 MB/s	      66 B/op	       3 allocs/op
//50000:  1894 ns/op	527725.23 MB/s	      66 B/op	       3 allocs/op
//100000: 2144 ns/op	466322.23 MB/s	      65 B/op	       3 allocs/op
//200000: 2204 ns/op	453568.02 MB/s	      66 B/op	       3 allocs/op
//500000: 2331 ns/op	428879.79 MB/s	      65 B/op	       3 allocs/op
//1000000:2584 ns/op	193459.45 MB/s	      65 B/op	       3 allocs/op
//2000000:2776 ns/op	180087.34 MB/s	      65 B/op	       3 allocs/op
func BenchmarkSet(b *testing.B) {
	b.StopTimer()
	N := 1000000
	baseSlice := make([]uint64, N*2)
	storeSlice := make([]uint64, b.N*2)
	quadtree := New()

	for i := 0; i < N*2; i += 2 {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		quadtree.Set(k, v)
		baseSlice[i] = k
		baseSlice[i+1] = v
	}

	for i := 0; i < b.N*2; i += 2 {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		//quadtree.Set(k, v)
		storeSlice[i] = k
		storeSlice[i+1] = v
	}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N*2; i += 2 {
		ii := i % N
		k := storeSlice[ii]
		v := storeSlice[ii+1]
		quadtree.Set(k, v)
	}
	b.StopTimer()
	//quadtreePrint2(&quadtree.Node, b, 64)
	//*
	for i := 0; i < b.N*2; i += 2 {
		ii := i % N
		k := storeSlice[ii]
		v := storeSlice[ii+1]
		//以上几步，耗时： 1.41 ns/op

		quadtree.Get(k)

		v2 := quadtree.Get(k)
		if v2 == nil || v2.value == nil {
			b.Errorf("v2:%v, i==%v, -->>k:%v <-->v:%v", v2, i, k, v)
			continue
		}
		if x, ok := v2.value.(uint64); !ok || x != v {
			b.Errorf("k:%v, v:%v <--> v2:%v", k, v, x)
		}
	} //*/
	b.SetBytes(int64(b.N))
	b.StartTimer()
}

//以下是sync.map
//100:    30.7 ns/op	1628444893.75 MB/s	       0 B/op	       0 allocs/op
//1000:   26.2 ns/op	1908055972.59 MB/s	       0 B/op	       0 allocs/op
//10000:  33.0 ns/op	1516938100.49 MB/s	       0 B/op	       0 allocs/op
//50000:  46.7 ns/op	1069894770.29 MB/s	       0 B/op	       0 allocs/op
//100000: 74.3 ns/op	269150386.43 MB/s	       0 B/op	       0 allocs/op
//200000: 60.3 ns/op	331589696.55 MB/s	       0 B/op	       0 allocs/op
//500000: 86.6 ns/op	231079899.92 MB/s	       0 B/op	       0 allocs/op
//1000000:97.7 ns/op	204616642.54 MB/s	       0 B/op	       0 allocs/op
//Quad-tree：sync.map是Q-tree的4倍
//100:     68.1 ns/op	293635372.55 MB/s	       0 B/op	       0 allocs/op
//1000:    71.8 ns/op	278424418.49 MB/s	       0 B/op	       0 allocs/op
//10000:   90.9 ns/op	219906013.27 MB/s	       0 B/op	       0 allocs/op
//50000:   205 ns/op	48665693.74 MB/s	       0 B/op	       0 allocs/op
//100000:  248 ns/op	20156413.77 MB/s	       0 B/op	       0 allocs/op
//500000:  323 ns/op	15479147.73 MB/s	       0 B/op	       0 allocs/op
//1000000: 370 ns/op	13506644.84 MB/s	       0 B/op	       0 allocs/op //32bit的key: 137 ns/op
//2000000: 420 ns/op	7129913.52 MB/s	       0 B/op	       0 allocs/op
//以下是skip-list: 10w数据量：3倍！
//1000000:	2244 ns/op	445490.54 MB/s	       0 B/op	       0 allocs/op
//100000:	724 ns/op	2760198.76 MB/s	       0 B/op	       0 allocs/op
//10000:	268 ns/op	18597862.51 MB/s	       0 B/op	       0 allocs/op
//1000:		114 ns/op	174162906.67 MB/s	       0 B/op	       0 allocs/op
//100:		101 ns/op	196672749.43 MB/s	       0 B/op	       0 allocs/op
func BenchmarkGet(b *testing.B) {
	b.StopTimer()

	N := 1000000
	storeSlice := make([]uint64, N*2)
	quadtree := New()

	for i := 0; i < N*2; i += 2 {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		quadtree.Set(k, v)
		storeSlice[i] = k
		storeSlice[i+1] = v
	}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N*2; i += 2 {
		ii := i % N
		k := storeSlice[ii]
		v := storeSlice[ii+1]
		_ = v
		quadtree.Get(k)
	}
	b.StopTimer()
	//quadtreePrint2(&quadtree.Node, b, 64)
	//*
	for i := 0; i < b.N*2; i += 2 {
		ii := i % N
		k := storeSlice[ii]
		v := storeSlice[ii+1]
		//以上几步，耗时： 1.41 ns/op

		quadtree.Get(k)

		v2 := quadtree.Get(k)
		if v2 == nil || v2.value == nil {
			b.Errorf("b.N:%d, v2:%v, i==%v, -->>k:%v <-->v:%v", b.N, v2, i, k, v)
			continue
		}
		if x, ok := v2.value.(uint64); !ok || x != v {
			b.Errorf("k:%v, v:%v <--> v2:%v", k, v, x)
		}
	} //*/
	b.SetBytes(int64(b.N))
	b.StartTimer()
}

//Quad-tree Get:
//100:     68.1 ns/op	293635372.55 MB/s	       0 B/op	       0 allocs/op
//1000:    71.8 ns/op	278424418.49 MB/s	       0 B/op	       0 allocs/op
//10000:   90.9 ns/op	219906013.27 MB/s	       0 B/op	       0 allocs/op
//50000:   205 ns/op	48665693.74 MB/s	       0 B/op	       0 allocs/op
//100000:  248 ns/op	20156413.77 MB/s	       0 B/op	       0 allocs/op
//500000:  323 ns/op	15479147.73 MB/s	       0 B/op	       0 allocs/op
//1000000: 370 ns/op	13506644.84 MB/s	       0 B/op	       0 allocs/op
//2000000: 420 ns/op	7129913.52 MB/s	       	   0 B/op	       0 allocs/op
//Quad-tree Get2:
//100:     68.1 ns/op	293635372.55 MB/s	       0 B/op	       0 allocs/op
//1000:    71.8 ns/op	278424418.49 MB/s	       0 B/op	       0 allocs/op
//10000:   90.9 ns/op	219906013.27 MB/s	       0 B/op	       0 allocs/op
//50000:   205 ns/op	48665693.74 MB/s	       0 B/op	       0 allocs/op
//100000:  248 ns/op	20156413.77 MB/s	       0 B/op	       0 allocs/op
//500000:  323 ns/op	15479147.73 MB/s	       0 B/op	       0 allocs/op
//1000000: 1649 ns/op	606413.39 MB/s	     144 B/op	      18 allocs/op
//2000000: 1943 ns/op	514435.50 MB/s	     144 B/op	      18 allocs/op
func BenchmarkGet2(b *testing.B) {
	b.StopTimer()

	N := 2000000
	storeSlice := make([]uint64, N*2)
	quadtree := New()

	for i := 0; i < N*2; i += 2 {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		quadtree.Set(k, v)
		storeSlice[i] = k
		storeSlice[i+1] = v
	}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N*2; i += 2 {
		ii := i % N
		k := storeSlice[ii]
		v := storeSlice[ii+1]
		_ = v
		quadtree.Get2(k, 30)
	}
	b.StopTimer()
	//quadtreePrint2(&quadtree.Node, b, 64)
	//*
	for i := 0; i < b.N*2; i += 2 {
		ii := i % N
		k := storeSlice[ii]
		v := storeSlice[ii+1]
		//以上几步，耗时： 1.41 ns/op

		quadtree.Get(k)

		v2 := quadtree.Get(k)
		if v2 == nil || v2.value == nil {
			b.Errorf("b.N:%d, v2:%v, i==%v, -->>k:%v <-->v:%v", b.N, v2, i, k, v)
			continue
		}
		if x, ok := v2.value.(uint64); !ok || x != v {
			b.Errorf("k:%v, v:%v <--> v2:%v", k, v, x)
		}
	} //*/
	b.SetBytes(int64(b.N))
	b.StartTimer()
}
func TestGet2(t *testing.T) {
	N := 100
	storeSlice := make([]uint64, N*2)
	quadtree := New()

	for i := 0; i < N*2; i += 2 {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		quadtree.Set(k, v)
		storeSlice[i] = k
		storeSlice[i+1] = v
	}
	t.Log(*quadtree.Get(storeSlice[0]))
	qs := quadtree.Get2(storeSlice[0], 30)
	for _, v := range qs {
		t.Log(*v)
	}
}

//BenchmarkSliceSet-4   	2000000000	         1.41 ns/op	1421772562952.98 MB/s	       0 B/op	       0 allocs/op
func BenchmarkSliceSet(b *testing.B) {
	b.StopTimer()

	N := 1000
	storeSlice := make([]uint64, N*2)
	//quadtree := New()

	for i := 0; i < N*2; i += 2 {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		//quadtree.Set(k, v)
		storeSlice[i] = k
		storeSlice[i+1] = v
	}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N*2; i += 2 {
		ii := i % N
		k := storeSlice[ii]
		v := storeSlice[ii+1]
		//quadtree.Set(k, v)
		_, _ = k, v
	}
	b.StopTimer()
	b.SetBytes(int64(b.N))
	b.StartTimer()
}
