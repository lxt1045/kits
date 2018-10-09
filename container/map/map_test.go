package map_test

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//100:    21.4 ns/op	4683314726.52 MB/s	       0 B/op	       0 allocs/op
//1000:   31.4 ns/op	1593064231.65 MB/s	       0 B/op	       0 allocs/op
//10000:  51.9 ns/op	385101019.22 MB/s	       0 B/op	       0 allocs/op
//50000:  73.0 ns/op	273941699.86 MB/s	       0 B/op	       0 allocs/op
//100000: 54.4 ns/op	367854179.66 MB/s	       0 B/op	       0 allocs/op
//200000: 89.9 ns/op	222472936.86 MB/s	       0 B/op	       0 allocs/op
//500000: 101 ns/op	98860523.72 MB/s	       4 B/op	       0 allocs/op
//1000000:106 ns/op	93912312.76 MB/s	       8 B/op	       0 allocs/op
func BenchmarkMapSet(b *testing.B) {
	b.StopTimer()
	N := 100000
	//Get和未初始化容量为N，所消耗的时间几乎没区别;
	//N := 1000000时，初始化80ns/op, 未初始化85ns/op
	//m := make(map[uint64]uint64, N*2)
	m := make(map[uint64]uint64)
	for i := 0; i <= N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		m[k] = v
	}
	inserS := make([]uint64, N)
	for i := 0; i < N; i++ {
		v := uint64(rand.Int63())
		inserS[i] = v
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		k := inserS[i%N]
		v := inserS[i%N]
		m[k] = v
	}
	b.Log(len(m))
	b.SetBytes(int64(b.N))
}

//100:    19.1 ns/op	5241362104.22 MB/s	       0 B/op	       0 allocs/op
//1000:   22.2 ns/op	2255571193.18 MB/s	       0 B/op	       0 allocs/op
//10000:  22.0 ns/op	2271927760.69 MB/s	       0 B/op	       0 allocs/op
//50000:  35.0 ns/op	1426935345.27 MB/s	       0 B/op	       0 allocs/op
//100000: 30.9 ns/op	1620444460.69 MB/s	       0 B/op	       0 allocs/op
//200000: 63.0 ns/op	476211515.16 MB/s	       0 B/op	       0 allocs/op
//500000: 69.7 ns/op	286900023.02 MB/s	       0 B/op	       0 allocs/op
//1000000:61.9 ns/op	323288490.53 MB/s	       0 B/op	       0 allocs/op
func BenchmarkMapGet(b *testing.B) {
	b.StopTimer()
	N := 10
	//m := make(map[uint64]uint64, N) //Get和未初始化容量为N，所消耗的时间几乎没区别
	m := make(map[uint64]uint64)
	for i := 0; i <= N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		m[k] = v
	}
	inserS := make([]uint64, N)
	for i := 0; i < N; i++ {
		v := uint64(rand.Int63())
		inserS[i] = v
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		k := inserS[i%N]
		v := m[k]
		_ = v
	}
	b.SetBytes(int64(b.N))
}

//以下是slice:
//10:	  3.20 ns/op	156144475611.58 MB/s	       0 B/op	       0 allocs/op
//100:    38.7 ns/op	774959041.26 MB/s	       0 B/op	       0 allocs/op
//1000:   297 ns/op	16799461.93 MB/s	       0 B/op	       0 allocs/op
//10000:  3017 ns/op	165676.11 MB/s	       0 B/op	       0 allocs/op
//100000: 30675 ns/op	1629.95 MB/s	       0 B/op	       0 allocs/op
//1000000:298822 ns/op	  16.73 MB/s	       0 B/op	       0 allocs/op
//以下是map： 50倍
//10:     149 ns/op		66926985.87 MB/s	   0 B/op	       0 allocs/op		len==10
//100:    1243 ns/op	804288.66 MB/s	       0 B/op	       0 allocs/op		len==100
//1000:   15301 ns/op	6535.38 MB/s	       0 B/op	       0 allocs/op		len==1000
//10000:  142281 ns/op	  70.28 MB/s	       0 B/op	       0 allocs/op		len==10000
//100000: 1283281 ns/op	   0.78 MB/s	       0 B/op	       0 allocs/op		len==100000
//1000000:15411748 ns/op   0.01 MB/s	       0 B/op	       0 allocs/op		len==1000000
func BenchmarkMapRange(b *testing.B) {
	b.StopTimer()
	N := 1000000
	//m := make(map[uint64]uint64, N) //Get和未初始化容量为N，所消耗的时间几乎没区别
	m := make(map[uint64]uint64)
	for i := 0; i < N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		m[k] = v
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for k, v := range m {
			_, _ = k, v
		}
	}
	//b.Log(len(m), "--", cap(m))
	b.SetBytes(int64(b.N))
}

func TestLenCap(t *testing.T) {
	{
		ch := make(chan int, 100)
		t.Logf("len(chan)==%d", len(ch))
		t.Logf("cap(chan)==%d", cap(ch))
	}
	{
		ch := make(chan int)
		t.Logf("len(chan)==%d", len(ch))
		t.Logf("cap(chan)==%d", cap(ch))
	}
	{
		ch := make(map[int]int)
		t.Logf("len(chan)==%d", len(ch))
		//t.Logf("cap(chan)==%d", cap(ch))
	}
	{
		ch := make(map[int]int, 100)
		t.Logf("len(chan)==%d", len(ch))
		//t.Logf("cap(chan)==%d", cap(ch))
	}
	{
		ch := make(map[int]int, 100)
		ch[1] = 2
		t.Logf("len(chan)==%d", len(ch))
		//t.Logf("cap(chan)==%d", cap(ch))
	}
}
