package algorithms

import (
	"fmt"
	//"fmt"
	"testing"
)

var xs []uint64
var testN []int

const N = 10e7

func init() {
	testN = []int{
		10, 10e2, 10e3, 10e4, 10e5, 10e6, 10e7,
	}
	xs = make([]uint64, N)
	for i := range xs {
		xs[i] = uint64(i)
	}
}
func TestSearch(t *testing.T) {
	a := []uint64{1, 2, 5, 7, 15, 25, 30, 36, 39, 51, 67, 78, 80, 82, 85, 91, 92, 97}
	k := BinSearchUint64(a, 38)
	t.Log("--:", k)
}

//    Search []int(10):				Total used:1.4171986s, Per operation:28ns, 50000000ns/ops
// 	  Search []int(1000):			Total used:1.8640647s, Per operation:93ns, 20000000ns/ops
// 	  Search []int(10000):			Total used:1.9368901s, Per operation:96ns, 20000000ns/ops
// 	  Search []int(100000):			Total used:2.2117424s, Per operation:110ns, 20000000ns/ops
// 	  Search []int(1000000):		Total used:1.2742707s, Per operation:127ns, 10000000ns/ops
// 	  Search []int(10000000):		Total used:1.3872075s, Per operation:138ns, 10000000ns/ops
// 	  Search []int(100000000):		Total used:1.4601741s, Per operation:146ns, 10000000ns/ops
//
// 	  SearchRaw []int(10):			Total used:2.3196605s, Per operation:23ns, 100000000ns/ops
// 	  SearchRaw []int(1000):		Total used:1.4961381s, Per operation:49ns, 30000000ns/ops
// 	  SearchRaw []int(10000):		Total used:1.6100844s, Per operation:53ns, 30000000ns/ops
// 	  SearchRaw []int(100000):		Total used:1.7729753s, Per operation:59ns, 30000000ns/ops
// 	  SearchRaw []int(1000000):		Total used:1.3282403s, Per operation:66ns, 20000000ns/ops
// 	  SearchRaw []int(10000000):	Total used:1.4371684s, Per operation:71ns, 20000000ns/ops
// 	  SearchRaw []int(100000000):	Total used:1.5930878s, Per operation:79ns, 20000000ns/ops
func TestSearchN(t *testing.T) {
	for _, n := range testN {
		printBR(t, fmt.Sprintf("Search []int(%d)", n),
			testing.Benchmark(func(b *testing.B) {
				benchmarkSearch(b, n)
			}))
	}
	for _, n := range testN {
		printBR(t, fmt.Sprintf("SearchRaw []int(%d)", n),
			testing.Benchmark(func(b *testing.B) {
				benchmarkSearchRaw(b, n)
			}))
	}
}

func printBR(t *testing.T, name string, br testing.BenchmarkResult) {
	t.Logf("%s:\tTotal used:%v, Per operation:%dns, %dns/ops",
		name, br.T, int64(br.T)/int64(br.N), br.N)
}

func benchmarkSearch(b *testing.B, l int) {
	N := l
	xs := xs[:l]
	errN := 0
	for i := 0; i < b.N; i++ { //use b.N for looping
		M := i % N
		k := BinSearchUint64(xs, uint64(M))
		if k != M {
			errN++
		}
	}
	if errN != 0 {
		b.Error(errN)
	}
}
func benchmarkSearchRaw(b *testing.B, l int) {
	N := l
	xs := xs[:l]
	errN := 0
	for i := 0; i < b.N; i++ { //use b.N for looping
		M := i % N
		k := BinSearchUint64Raw(xs, uint64(M))
		if k != M {
			errN++
		}
	}
	if errN != 0 {
		b.Error(errN)
	}
}

//10e3 * 10e3
//BenchmarkFor-4   	       1	2955309700 ns/op	       0 B/op	       0 allocs/op
//BenchmarkFor-4   	       1	3007288200 ns/op	       0 B/op	       0 allocs/op
//10e1 * 10e1
//BenchmarkFor-4   	  300000	      5636 ns/op	       0 B/op	       0 allocs/op
//BenchmarkFor-4   	  200000	      6168 ns/op	       0 B/op	       0 allocs/op
//10 * 10 = 100
//BenchmarkFor-4   	20000000	        77.6 ns/op	       0 B/op	       0 allocs/op
//BenchmarkFor-4   	20000000	        75.1 ns/op	       0 B/op	       0 allocs/op
//32*32 = 1024
//BenchmarkFor-4   	 2000000	       627 ns/op	       0 B/op	       0 allocs/op
//BenchmarkFor-4   	 2000000	       672 ns/op	       0 B/op	       0 allocs/op
//100*100 =1000
//BenchmarkFor-4   	  300000	      5816 ns/op	       0 B/op	       0 allocs/op
func BenchmarkFor(b *testing.B) {
	X := 1
	for i := 0; i < b.N; i++ { //use b.N for looping
		for j := 0; j < 100; j++ {
			for k := 0; k < 100; k++ {
				_ = X
			}
		}
	}
}

//10e6
//BenchmarkFor2-4   	       5	 294231800 ns/op	       0 B/op	       0 allocs/op
//BenchmarkFor2-4   	       5	 304961860 ns/op	       0 B/op	       0 allocs/op
//10e2
//BenchmarkFor2-4   	 5000000	       323 ns/op	       0 B/op	       0 allocs/op
//BenchmarkFor2-4   	 5000000	       333 ns/op	       0 B/op	       0 allocs/op
//100
//BenchmarkFor2-4   	30000000	        57.0 ns/op	       0 B/op	       0 allocs/op
//BenchmarkFor2-4   	20000000	        59.0 ns/op	       0 B/op	       0 allocs/op
//32*32=1024
//BenchmarkFor2-4   	 5000000	       326 ns/op	       0 B/op	       0 allocs/op
//BenchmarkFor2-4   	 5000000	       342 ns/op	       0 B/op	       0 allocs/op
//10000
//BenchmarkFor2-4   	  500000	      3020 ns/op	       0 B/op	       0 allocs/op
func BenchmarkFor2(b *testing.B) {
	X := 1
	for i := 0; i < b.N; i++ { //use b.N for looping
		for k := 0; k < 10000; k++ {
			_ = X
		}
	}
}
