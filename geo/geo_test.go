package geo

import (
	"bytes"
	"math"
	"testing"
)

var X [][2]float64

func init() {
	const N = 1000 + 7
	X = make([][2]float64, 0, N)
	X = append(X, [2]float64{39.92324, 116.3906})
	for i := 1; i < N; i++ {
		x := X[i-1][0] + 0.00123
		y := X[i-1][1] + 0.00456
		X = append(X, [2]float64{x, y})
	}
	X = append(X, [2]float64{0.92324, 0.3906})
	X = append(X, [2]float64{90, 180})
	X = append(X, [2]float64{-0.92324, -0.3906})
	X = append(X, [2]float64{-90, -180})
	X = append(X, [2]float64{89.999999, 179.99999})
	X = append(X, [2]float64{0.000001, 0.000001})
	X = append(X, [2]float64{-89.999999999999, -179.999999999999})
}

//输入坐标：(39.92324, 116.3906, 0)； 预计返回 ： wx4g0ec19x3d
func TestCoords2Geo1(t *testing.T) {
	x1 := Coords2GeoStr(116.3906, 39.92324)
	x, _ := Encode(39.92324, 116.3906, int(geoBinBitsCount/5))
	if x1 != x {
		t.Errorf("my:%s <--> his:%s", x1, x)
	}
}
func TestCoords2Geo(t *testing.T) {
	//getArray()
	for _, val := range X {
		x1 := Coords2GeoStr(val[1], val[0])
		x, _ := Encode(val[0], val[1], int(geoBinBitsCount/5))
		if x1 != x {
			t.Errorf("my:%s <--> his:%s", x1, x)
		}
	}
}

func TestGetAreas(t *testing.T) {
	calculateY := func(crd1, crd2 Coords) float64 {
		deltaY := crd1.Lat - crd2.Lat
		if deltaY < 0 {
			deltaY = -deltaY
		}
		C := 40037.0 / 2
		return deltaY / 180.0 * C
	}
	calculateX := func(crd1, crd2 Coords) float64 {
		alphaY := crd1.Lat
		deltaX := crd1.Lng - crd2.Lng
		if deltaX < 0 {
			deltaX = -deltaX
		}
		cosY := math.Cos(math.Pi * alphaY / 180)
		C := 40075.2 * cosY
		return deltaX / 360.0 * C
	}

	const N = 10000
	const M = 0.03
	for i := 1; i < N; i++ {
		R := float64(i) * M
		lngL, lngR, latU, latD := R2Area(116.3906, 39.92324, R)
		//t.Logf("lngL=%f, lngR=%f, latU=%f, latD=%f", lngL, lngR, latU, latD)
		r1 := calculateX(Coords{lngL, latU}, Coords{lngR, latU}) / 2
		r2 := calculateY(Coords{lngL, latU}, Coords{lngL, latD}) / 2
		if r1-R > R || r2-R > R {
			t.Errorf("R2Area(116.3906, 39.92324, R), r1=%f, r2=%f,R=%f", r1, r2, R)
		}
		lngL, lngR, latU, latD = R2Area(108.329518, 22.824924, R)
		r1 = calculateX(Coords{lngL, latU}, Coords{lngR, latU}) / 2
		r2 = calculateY(Coords{lngL, latU}, Coords{lngL, latD}) / 2
		if r1-R > R || r2-R > R {
			t.Errorf("R2Area(108.329518, 22.824924, R), r1=%f, r2=%f,R=%f", r1, r2, R)
		}
	}
	geohashS := R2Geos2(116.3906, 39.92324, 2)
	x1, x2 := mergeGeos(geohashS, 2, 64-32)
	t.Log(x1)
	t.Log(x2)
}

func TestR2MergeGeos(t *testing.T) {
	bits := Len2Bits(50.0 / 1000.0)
	t.Log(bits)
	bits = Len2Bits(150.0 / 1000.0)
	t.Log(bits)
	bits = Len2Bits(300.0 / 1000.0)
	t.Log(bits)
	x1, x2 := R2MergeGeos(116.3363, 39.91350, 0.55, uint32(bits)+1) //e748f0340000000 e748f0300000000 e748da900000000
	t.Logf("%x", x1)
	t.Logf("%x", x2)

	x1, x2 = R2MergeGeos2(Coords2Geo(116.3363, 39.91350), 0.55, uint32(bits)+1) //e748f0340000000 e748f0300000000 e748da900000000
	t.Logf("%x", x1)
	t.Logf("%x", x2)
}
func TestUint32ToCoords(t *testing.T) {
	x0, y0 := 116.3363, 39.91350
	t.Log(x0, y0)
	x1, y1 := Coords2Uint32(x0, y0)
	x2, y2 := Uint32ToCoords(x1, y1)
	t.Log(x2, y2) //32bits: 116.3362999353559 39.91349997930237 //30bits: 116.33629985153686 39.91349989548334
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////
//以下代码来自"github.com/gansidui/geohash"，LICENSE类型：MIT
//本测试用一下代码做对比

const (
	//BASE32                = "0123456789bcdefghjkmnpqrstuvwxyz"
	MAX_LATITUDE  float64 = 90
	MIN_LATITUDE  float64 = -90
	MAX_LONGITUDE float64 = 180
	MIN_LONGITUDE float64 = -180
)

var (
	bits = []int{16, 8, 4, 2, 1}
	//base32 = []byte(BASE32)
)

type Box struct {
	MinLat, MaxLat float64 // 纬度
	MinLng, MaxLng float64 // 经度
}

func (this *Box) Width() float64 {
	return this.MaxLng - this.MinLng
}

func (this *Box) Height() float64 {
	return this.MaxLat - this.MinLat
}

// 输入值：纬度，经度，精度(geohash的长度)
// 返回geohash, 以及该点所在的区域
func Encode(latitude, longitude float64, precision int) (string, *Box) {
	var geohash bytes.Buffer
	var minLat, maxLat float64 = MIN_LATITUDE, MAX_LATITUDE
	var minLng, maxLng float64 = MIN_LONGITUDE, MAX_LONGITUDE
	var mid float64 = 0

	bit, ch, length, isEven := 0, 0, 0, true
	for length < precision {
		if isEven {
			if mid = (minLng + maxLng) / 2; mid < longitude {
				ch |= bits[bit]
				minLng = mid
			} else {
				maxLng = mid
			}
		} else {
			if mid = (minLat + maxLat) / 2; mid < latitude {
				ch |= bits[bit]
				minLat = mid
			} else {
				maxLat = mid
			}
		}

		isEven = !isEven
		if bit < 4 {
			bit++
		} else {
			geohash.WriteByte(base32[ch])
			length, bit, ch = length+1, 0, 0
		}
	}

	b := &Box{
		MinLat: minLat,
		MaxLat: maxLat,
		MinLng: minLng,
		MaxLng: maxLng,
	}

	return geohash.String(), b
}

// 计算该点（latitude, longitude）在精度precision下的邻居 -- 周围8个区域+本身所在区域
// 返回这些区域的geohash值，总共9个
func GetNeighbors(latitude, longitude float64, precision int) []string {
	geohashs := make([]string, 9)

	// 本身
	geohash, b := Encode(latitude, longitude, precision)
	geohashs[0] = geohash

	// 上下左右
	geohashUp, _ := Encode((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2, precision)
	geohashDown, _ := Encode((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2, precision)
	geohashLeft, _ := Encode((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2-b.Width(), precision)
	geohashRight, _ := Encode((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2+b.Width(), precision)

	// 四个角
	geohashLeftUp, _ := Encode((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
	geohashLeftDown, _ := Encode((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
	geohashRightUp, _ := Encode((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)
	geohashRightDown, _ := Encode((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)

	geohashs[1], geohashs[2], geohashs[3], geohashs[4] = geohashUp, geohashDown, geohashLeft, geohashRight
	geohashs[5], geohashs[6], geohashs[7], geohashs[8] = geohashLeftUp, geohashLeftDown, geohashRightUp, geohashRightDown

	return geohashs
}
