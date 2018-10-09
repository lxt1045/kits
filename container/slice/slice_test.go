package map_test

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//100:    2.95 ns/op	169246364892.73 MB/s	       0 B/op	       0 allocs/op
//1000:   2.76 ns/op	362657702155.10 MB/s	       0 B/op	       0 allocs/op
//10000:  2.54 ns/op	394389462828.54 MB/s	       0 B/op	       0 allocs/op
//100000: 3.08 ns/op	162325983299.38 MB/s	       0 B/op	       0 allocs/op
//1000000:9.37 ns/op	21334335762.66 MB/s	       0 B/op	       0 allocs/op
func BenchmarkMapSet(b *testing.B) {
	b.StopTimer()
	N := 1000
	m := make([]uint64, N)
	for i := 0; i < N; i++ {
		v := uint64(rand.Int63())
		m[i] = v
	}
	inserS := make([]uint64, N)
	for i := 0; i < N; i++ {
		v := uint64(rand.Int63())
		inserS[i] = v
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		v := inserS[i%N]
		k := int(v) % N
		m[k] = v
	}
	b.SetBytes(int64(b.N))
}

//10:     2.81 ns/op	177785590861.88 MB/s	       0 B/op	       0 allocs/op
//100:    2.93 ns/op	341022178513.79 MB/s	       0 B/op	       0 allocs/op
//1000:   2.51 ns/op	398790993321.37 MB/s	       0 B/op	       0 allocs/op
//10000:  2.62 ns/op	382288453611.86 MB/s	       0 B/op	       0 allocs/op
//100000: 2.56 ns/op	390541948813.70 MB/s	       0 B/op	       0 allocs/op
//1000000:2.66 ns/op	187934090161.45 MB/s	       0 B/op	       0 allocs/op
func BenchmarkMapGet(b *testing.B) {
	b.StopTimer()
	N := 10
	m := make([]uint64, N)
	for i := 0; i < N; i++ {
		v := uint64(rand.Int63())
		m[i] = v
	}
	inserS := make([]uint64, N)
	for i := 0; i < N; i++ {
		v := uint64(rand.Int63())
		inserS[i] = v
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		v := inserS[i%N]
		k := int(v) % N
		v = m[k]
	}
	b.SetBytes(int64(b.N))
}

//10:	  3.20 ns/op	156144475611.58 MB/s	       0 B/op	       0 allocs/op
//100:    38.7 ns/op	774959041.26 MB/s	       0 B/op	       0 allocs/op
//1000:   297 ns/op 	16799461.93 MB/s	       0 B/op	       0 allocs/op
//10000:  3017 ns/op	165676.11 MB/s	       0 B/op	       0 allocs/op
//100000: 30675 ns/op	1629.95 MB/s	       0 B/op	       0 allocs/op
//1000000:298822 ns/op	  16.73 MB/s	       0 B/op	       0 allocs/op
func BenchmarkSliceRange(b *testing.B) {
	b.StopTimer()
	N := 16
	//m := make(map[uint64]uint64, N) //Get和未初始化容量为N，所消耗的时间几乎没区别
	m := make([]uint64, N)
	for i := 0; i < N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		m[int(k)%N] = v
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for k, v := range m {
			_, _ = k, v
		}
	}
	b.SetBytes(int64(b.N))
}
