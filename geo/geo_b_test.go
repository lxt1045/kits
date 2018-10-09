package geo

import (
	"testing"
)

//*
func BenchmarkEncodeMy(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		Coords2Geo(39.92324, 116.3906)
	}
}
func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		Encode(39.92324, 116.3906, 12)
	}
}

func BenchmarkGeoMerge(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		interleave64(23651, 26978)
	}
}

//BenchmarkDist-4   	50000000	        28.6 ns/op	       0 B/op	       0 allocs/op
//BenchmarkDist-4   	50000000	        27.6 ns/op	       0 B/op	       0 allocs/op
func BenchmarkDist(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		Dist(Coords{116.3906, 39.92324}, Coords{118.3906, 38.92324})
	}
}

//*/
//BenchmarkDist2-4   	50000000	        25.5 ns/op	       0 B/op	       0 allocs/op
func BenchmarkDist2(b *testing.B) {
	c1, c2 := Coords{116.3906, 39.92324}, Coords{118.3906, 38.92324}
	for i := 0; i < b.N; i++ { //use b.N for looping
		Dist2(c1, c2)
	}
}

//240 ns/op	      96 B/op	       3 allocs/op
func BenchmarkR2MergeGeos(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		R2MergeGeos(116.3363, 39.91350, 0.55, 17)
	}
}

//256 ns/op	      96 B/op	       3 allocs/op
func BenchmarkR2MergeGeos2(b *testing.B) {
	geohash := Coords2Geo(116.3363, 39.91350)
	for i := 0; i < b.N; i++ { //use b.N for looping
		R2MergeGeos2(geohash, 0.55, 17)
	}
}
