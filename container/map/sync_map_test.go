package map_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//以下是原版map:
//100:    21.4 ns/op	4683314726.52 MB/s	       0 B/op	       0 allocs/op
//1000:   31.4 ns/op	1593064231.65 MB/s	       0 B/op	       0 allocs/op
//10000:  51.9 ns/op	385101019.22 MB/s	       0 B/op	       0 allocs/op
//50000:  73.0 ns/op	273941699.86 MB/s	       0 B/op	       0 allocs/op
//100000: 54.4 ns/op	367854179.66 MB/s	       0 B/op	       0 allocs/op
//200000: 89.9 ns/op	222472936.86 MB/s	       0 B/op	       0 allocs/op
//500000: 101 ns/op	98860523.72 MB/s	       4 B/op	       0 allocs/op
//1000000:106 ns/op	93912312.76 MB/s	       8 B/op	       0 allocs/op
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
func BenchmarkSyncMapSet(b *testing.B) {
	b.StopTimer()
	N := 2000000
	m := sync.Map{}
	for i := 0; i <= N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		m.Store(k, v)
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
		m.Store(k, v)
	}
	b.SetBytes(int64(b.N))
}

//以下是sync.map的Renew，性能和set差不多
func BenchmarkSyncMapRenew(b *testing.B) {
	b.StopTimer()
	N := 100000
	m := sync.Map{}
	inserS := make([]uint64, N)
	for i := 0; i < N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		v2 := uint64(rand.Int63())

		v2 = v
		inserS[i] = v2
		m.Store(k, v)
	}
	for i := 0; i < N; i++ {
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		k := inserS[i%N]
		v := inserS[i%N]
		m.Store(k, v)
	}
	//b.Log(len(m))
	b.SetBytes(int64(b.N))
}

//以下是原版map：
//100:    19.1 ns/op	5241362104.22 MB/s	       0 B/op	       0 allocs/op
//1000:   22.2 ns/op	2255571193.18 MB/s	       0 B/op	       0 allocs/op
//10000:  22.0 ns/op	2271927760.69 MB/s	       0 B/op	       0 allocs/op
//50000:  35.0 ns/op	1426935345.27 MB/s	       0 B/op	       0 allocs/op
//100000: 30.9 ns/op	1620444460.69 MB/s	       0 B/op	       0 allocs/op
//200000: 63.0 ns/op	476211515.16 MB/s	       0 B/op	       0 allocs/op
//500000: 69.7 ns/op	286900023.02 MB/s	       0 B/op	       0 allocs/op
//1000000:61.9 ns/op	323288490.53 MB/s	       0 B/op	       0 allocs/op
//以下是sync.map
//100:    30.7 ns/op	1628444893.75 MB/s	       0 B/op	       0 allocs/op
//1000:   26.2 ns/op	1908055972.59 MB/s	       0 B/op	       0 allocs/op
//10000:  33.0 ns/op	1516938100.49 MB/s	       0 B/op	       0 allocs/op
//50000:  46.7 ns/op	1069894770.29 MB/s	       0 B/op	       0 allocs/op
//100000: 74.3 ns/op	269150386.43 MB/s	       0 B/op	       0 allocs/op
//200000: 60.3 ns/op	331589696.55 MB/s	       0 B/op	       0 allocs/op
//500000: 86.6 ns/op	231079899.92 MB/s	       0 B/op	       0 allocs/op
//1000000:97.7 ns/op	204616642.54 MB/s	       0 B/op	       0 allocs/op
func BenchmarkSyncMapGet(b *testing.B) {
	b.StopTimer()
	N := 1000000
	m := sync.Map{}
	for i := 0; i <= N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		m.Store(k, v)
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
		v, _ := m.Load(k)
		_ = v
	}
	b.SetBytes(int64(b.N))
}

//以下是slice:
//10:	  3.20 ns/op	156144475611.58 MB/s       0 B/op	       0 allocs/op
//100:    38.7 ns/op	774959041.26 MB/s	       0 B/op	       0 allocs/op
//1000:   297 ns/op 	16799461.93 MB/s	       0 B/op	       0 allocs/op
//10000:  3017 ns/op	165676.11 MB/s    	       0 B/op	       0 allocs/op
//100000: 30675 ns/op	1629.95 MB/s    	       0 B/op	       0 allocs/op
//1000000:298822 ns/op	  16.73 MB/s    	       0 B/op	       0 allocs/op
//以下是map： 50倍
//10:     14.9 ns/op		66926985.87 MB/s	   0 B/op	       0 allocs/op		len==10
//100:    12.43 ns/op   	804288.66 MB/s	       0 B/op	       0 allocs/op		len==100
//1000:   15.301 ns/op  	6535.38 MB/s	       0 B/op	       0 allocs/op		len==1000
//10000:  14.2281 ns/op	    70.28 MB/s  	       0 B/op	       0 allocs/op		len==10000
//100000: 12.83281 ns/op    0.78 MB/s   	       0 B/op	       0 allocs/op		len==100000
//1000000:15.411748 ns/op   0.01 MB/s   	       0 B/op	       0 allocs/op		len==1000000
//以下是sync.map： 50倍
//10:     14.9 ns/op		66926985.87 MB/s	   0 B/op	       0 allocs/op		len==10
//100:    12.43 ns/op   	804288.66 MB/s	       0 B/op	       0 allocs/op		len==100
//1000:   27.912 ns/op  	1791.34 MB/s	       0 B/op	       0 allocs/op
//10000:  25.7250 ns/op 	  19.44 MB/s	       0 B/op	       0 allocs/op
//100000: 45.64944 ns/op	   0.07 MB/s	       0 B/op	       0 allocs/op
//1000000:121.701350 ns/op	   0.00 MB/s	       1 B/op	       0 allocs/op
func BenchmarkSyncMapRange(b *testing.B) {
	b.StopTimer()
	N := 1000
	m := sync.Map{}
	for i := 0; i < N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		m.Store(k, v)
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Range(func(key, value interface{}) bool {
			_, _ = key, value
			return true
		})
	}
	b.SetBytes(int64(b.N))
}
