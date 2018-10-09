package skiplist

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
	return fmt.Sprintf("%dns/op,\t%.3fs,\t%8d,",
		int64(br.T)/int64(br.N), float64(br.T)/float64(time.Second), br.N)

}

// 10:  	87ns/op,	1.750s,	20000000,	avg:1.000000
// 100: 	110ns/op,	2.215s,	20000000,	avg:1.020000
// 1000:	169ns/op,	1.693s,	10000000,	avg:1.036000
// 10000:	390ns/op,	1.952s,	5000000,	avg:1.301900
// 100000:	1135ns/op,	1.135s,	1000000,	avg:4.046670
// 1000000:	8610ns/op,	1.722s,	200000,	avg:31.512060
//  10: 	86ns/op,	1.723s,	20000000,	avg:1.000000
// 100: 	113ns/op,	2.277s,	20000000,	avg:1.000000
// 1000:	164ns/op,	1.645s,	10000000,	avg:1.000000
// 10000:	347ns/op,	1.740s,	5000000,	avg:1.000000
// 100000:	712ns/op,	1.425s,	2000000,	avg:1.000000
// 1000000:	2278ns/op,	2.279s,	1000000,	avg:1.000000
//  10: 	36ns/op,	1.831s,	50000000,	avg:0.000000
// 100: 	62ns/op,	1.889s,	30000000,	avg:0.000000
// 1000:	106ns/op,	2.121s,	20000000,	avg:0.000000
// 10000:	267ns/op,	1.335s,	5000000,	avg:0.000000
// 100000:	609ns/op,	1.830s,	3000000,	avg:0.000000
// 1000000:	2134ns/op,	2.135s,	1000000,	avg:0.000000
func aTestBench(t *testing.T) {
	logs := bytes.NewBufferString("\n")
	var avg float64
	for _, n := range testN {
		logs.WriteString(fmt.Sprintf("%7d:\t", n))
		logs.WriteString(
			format(
				testing.Benchmark(
					func(b *testing.B) {
						benchmarkRandGet2(b, n, &avg, 16)
					},
				),
			),
		)
		logs.WriteString(fmt.Sprintf("\tavg:%f\n", avg))
	}
	for _, n := range testN {
		logs.WriteString(fmt.Sprintf("%7d:\t", n))
		logs.WriteString(
			format(
				testing.Benchmark(
					func(b *testing.B) {
						benchmarkRandGet2(b, n, &avg, 64)
					},
				),
			),
		)
		logs.WriteString(fmt.Sprintf("\tavg:%f\n", avg))
	}
	avg = 0
	for _, n := range testN {
		logs.WriteString(fmt.Sprintf("%7d:\t", n))
		logs.WriteString(
			format(
				testing.Benchmark(
					func(b *testing.B) {
						benchmarkRandGet(b, n, &avg)
					},
				),
			),
		)
		logs.WriteString(fmt.Sprintf("\tavg:%f\n", avg))
	}
	t.Log(logs)
}

// 10:  	268ns/op,	1.342s,	5000000,	avg:0.000000
// 100: 	303ns/op,	1.518s,	5000000,	avg:0.000000
// 1000:	379ns/op,	1.898s,	5000000,	avg:0.000000
// 10000:	651ns/op,	1.302s,	2000000,	avg:0.000000
// 100000:	1202ns/op,	1.202s,	1000000,	avg:0.000000
// 1000000:	2708ns/op,	1.354s,	500000, 	avg:0.000000
//  10: 	242ns/op,	1.210s,	5000000,	avg:0.000000
// 100: 	302ns/op,	1.512s,	5000000,	avg:0.000000
// 1000:	378ns/op,	1.895s,	5000000,	avg:0.000000
// 10000:	654ns/op,	1.308s,	2000000,	avg:0.000000
// 100000:	1322ns/op,	1.322s,	1000000,	avg:0.000000
// 1000000:	4101ns/op,	2.051s,	500000, 	avg:0.000000
func TestBenchRenewKey(t *testing.T) {
	logs := bytes.NewBufferString("\n")
	var avg float64
	for _, n := range testN {
		logs.WriteString(fmt.Sprintf("%7d:\t", n))
		logs.WriteString(
			format(
				testing.Benchmark(
					func(b *testing.B) {
						benchmarkRenewKey(b, n, 16)
					},
				),
			),
		)
		logs.WriteString(fmt.Sprintf("\tavg:%f\n", avg))
	}
	for _, n := range testN {
		logs.WriteString(fmt.Sprintf("%7d:\t", n))
		logs.WriteString(
			format(
				testing.Benchmark(
					func(b *testing.B) {
						benchmarkRenewKey(b, n, 48)
					},
				),
			),
		)
		logs.WriteString(fmt.Sprintf("\tavg:%f\n", avg))
	}
	t.Log(logs)
}

func benchmarkRandGet2(b *testing.B, N int, avg *float64, bits uint64) {
	b.StopTimer()
	benchList := New()
	var xs _nodeSlice
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
	var all int
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		j := i % N
		es := benchList.Get2(xs[j].k, bits) //16
		all += len(es)
	}
	*avg = float64(all) / float64(b.N)
	b.SetBytes(int64(b.N))
}

func benchmarkRandGet(b *testing.B, N int, avg *float64) {
	b.StopTimer()
	benchList := New()
	var xs _nodeSlice
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
		benchList.Get(xs[j].k)
	}
	b.SetBytes(int64(b.N))
}

func TestRandGet2(t *testing.T) {
	N := 100000
	benchList := New()
	var xs _nodeSlice
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
	var all int
	es := benchList.Get2(xs[0].k, 16)
	all += len(es)

	t.Logf("len:%d, es:%v", all, es)
	t.Log(benchList.Get(xs[0].k))
}
func benchmarkRenewKey(b *testing.B, N int, add uint64) {
	b.StopTimer()
	benchList := New()
	var xs _nodeSlice
	for i := 0; i <= N; i++ {
		k := uint64(rand.Int63())
		v := uint64(rand.Int63())
		benchList.Set(k, v)
		xs.Set(k, v)
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		j := i % N
		//benchList.Get(xs[j].k)
		newkey := xs[j].k + 1<<add
		benchList.RenewKey(newkey, xs[j].k, xs[j].v)
		xs[j].k = newkey
	}
	b.SetBytes(int64(b.N))
}
