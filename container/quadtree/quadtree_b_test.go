package quadtree

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var testN []int

func init() {
	testN = []int{
		10, 100, 1000, 10000, 100000, 1000000,
	}
}
func format(br testing.BenchmarkResult) string {
	return fmt.Sprintf("%dns/op,\t%.3fs,\t%d,",
		int64(br.T)/int64(br.N), float64(br.T)/float64(time.Second), br.N)

}
func TestBench(t *testing.T) {
	logs := bytes.NewBufferString("\n")
	var avg float64
	for _, n := range testN {
		logs.WriteString(fmt.Sprintf("%7d:\t", n))
		logs.WriteString(
			format(
				testing.Benchmark(
					func(b *testing.B) {
						benchmarkGet2(b, n, &avg)
					},
				),
			),
		)
		logs.WriteString(fmt.Sprintf("\tavg:%f\n", avg))
	}
	t.Log(logs)
}

//Quad-tree Get:
//100:     68.1 ns/op	293635372.55 MB/s	       0 B/op	       0 allocs/op
//1000:    71.8 ns/op	278424418.49 MB/s	       0 B/op	       0 allocs/op
//10000:   90.9 ns/op	219906013.27 MB/s	       0 B/op	       0 allocs/op
//100000:  248 ns/op	20156413.77 MB/s	       0 B/op	       0 allocs/op
//1000000: 370 ns/op	13506644.84 MB/s	       0 B/op	       0 allocs/op
//Quad-tree Get2: 64
// 10:  	88ns/op,	1.761s,	20000000,	avg:1.000000
// 100: 	106ns/op,	2.136s,	20000000,	avg:1.000000
// 1000:	113ns/op,	1.133s,	10000000,	avg:1.000000
// 10000:	140ns/op,	1.400s,	10000000,	avg:1.000000
// 100000:	291ns/op,	1.455s,	5000000,	avg:1.000000
// 1000000:	451ns/op,	1.353s,	3000000,	avg:1.000000
//Quad-tree Get2: 60
// 10:  	136ns/op,	1.363s,	10000000,	avg:1.000000
// 100: 	162ns/op,	1.627s,	10000000,	avg:1.000000
// 1000:	176ns/op,	1.768s,	10000000,	avg:1.000000
// 10000:	207ns/op,	2.074s,	10000000,	avg:1.000000
// 100000:	382ns/op,	1.911s,	5000000,	avg:1.000000
// 1000000:	506ns/op,	1.519s,	3000000,	avg:1.000000
//Quad-tree Get2: 36
// 10:  	190ns/op,	1.906s,	10000000,  avg:1.000000
// 100: 	201ns/op,	2.016s,	10000000,  avg:1.000000
// 1000:	294ns/op,	1.474s,	5000000,  avg:1.000000
// 10000:	324ns/op,	1.624s,	5000000,  avg:1.000000
// 100000:	446ns/op,	1.338s,	3000000,  avg:1.000000
// 1000000:	574ns/op,	1.723s,	3000000,  avg:1.000026
//Quad-tree Get2: 16
// 10:  	263ns/op,	1.316s,	5000000,  avg:1.000000
// 100: 	285ns/op,	1.425s,	5000000,  avg:1.000000
// 1000:	430ns/op,	1.292s,	3000000,  avg:1.026000
// 10000:	594ns/op,	1.188s,	2000000,  avg:1.289600
// 100000:	2544ns/op,	1.272s,	500000,   avg:4.055840
// 1000000:	19918ns/op,	1.992s,	100000,   avg:31.513680
func benchmarkGet2(b *testing.B, N int, avg *float64) {
	b.StopTimer()
	storeSlice := make([]uint64, N*2)
	quadtree := New()

	for i := 0; i < N*2; i += 2 {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		quadtree.Set(k, v)
		storeSlice[i] = k
		storeSlice[i+1] = v
	}
	var all int
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N*2; i += 2 {
		ii := i % N
		k := storeSlice[ii]
		v := storeSlice[ii+1]
		_ = v
		qs := quadtree.Get2(k, 16)
		all += len(qs)
	}
	*avg = float64(all) / float64(b.N)
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
