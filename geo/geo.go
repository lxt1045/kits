package geo

import (
	//"fmt"
	"math"
	"sort"
)

/*
//赤道周长 40075.2km ,经线长度 40037/2 km；/12位geohash,即30bit 经度,30bit 纬度,
//矩形的长为:40075.2km * cos(latitude) / 2^30
//矩形的宽为:40037km /2 / 2^30  20018.5
//geohash位数:    12            11            10            9            8            7             6            5            4            3            2            1
//经线/纬线位数:    30/30        27/28        25/25        22/23        20/20        17/18          15/15        12/13        10/10        7/8            5/5            2/3
//纬度/经度
//60°时,长:                                                           0.0191         0.0764       0.6115        2.446        19.568        78.27
//45°时,长:                                                           0.0270         0.108        0.865         3.46         27.67        110.69
//30°时,长:                                                           0.0331         0.1324       1.059         4.237        33.89        135.57
//0°时(赤道),长:  0.0000373    0.0001493   0.0011943    0.0047773     0.0382187      0.15287      1.2230        4.8920       39.136       156.54        1252.3        5009.4
//宽:            0.0000186    0.0001491   0.0005960    0.0047728     0.0190911      0.15273      0.6109        4.8873       19.549       156.39        625.27        5004.6
//三亚18°左右,黑龙江50°左右:                                           0.03km/0.02    0.12km/0.08   0.9km/0.6     3.8/4.8      31/19        123/156
//误差估值：                                                            0.02km         0.1km         0.6km         4km          25km         140kn
//为了计算方便,取近似值:                                                                0.1km         0.6km         4km           25km         140km
//*/

const (
	earthClat = 40075.2     //地球赤道周长40075.2km
	earthClng = 40037.0 / 2 //地球经线长度40037.0/2km

	geoBinBitsCount = 60 //用30bit表示，精度：3.73cm
	base32          = "0123456789bcdefghjkmnpqrstuvwxyz"
)

var (
//geohash bits长度对应的长宽的大小
)

//Coords  坐标
type Coords struct {
	Lng float64 //   经度
	Lat float64 //纬度
}

//Area 一个闭合区域：收尾相接
type Area struct {
	area []Coords
}

//Coords2Geo 输入值:纬度,经度,精度(geohash的长度), 返回geohash
func Coords2Geo(longitude, latitude float64) uint64 {
	x, y := Coords2Uint32(longitude, latitude)
	return interleave64(x, y)
} //Coords2Geo 输入值:纬度,经度,精度(geohash的长度), 返回geohash
func Geo2Coords(geo uint64) (lng, lat float64) {
	//x, y := Coords2Uint32(longitude, latitude)
	x, y := deinterleave64(geo)
	_, _ = x, y
	return
}

//Coords2GeoStr 坐标直接转成geohash的显示格式
func Coords2GeoStr(longitude, latitude float64) string {
	return Geo2Str(Coords2Geo(longitude, latitude))
}

//Coords2Uint32 输入值:纬度,经度,精度(geohash的长度), 返回一个正方形区域, 该区域由4个坐标点表示
func Coords2Uint32(longitude, latitude float64) (uint32, uint32) {
	const deltaAngle = 0.0000000000001 //角度偏移,为了计算方便;//不用DELTAANGLE时,在（90/180）会有问题

	var reciprocalPrecisionLat = float64(1<<(geoBinBitsCount/2)) / (180.0 + deltaAngle) //纬度的刻度的倒数,每度有多少个刻度值
	var reciprocalPrecisionLng = float64(1<<(geoBinBitsCount/2)) / (360.0 + deltaAngle) //经度的刻度的倒数,每度有多少个刻度值

	valLat := (latitude + 90) * reciprocalPrecisionLat   //直接使用90,
	valLng := (longitude + 180) * reciprocalPrecisionLng //直接使用180,在（90.180）会有问题

	return uint32(valLng), uint32(valLat)
}

//Uint32ToCoords 输入值:纬度,经度,精度(geohash的长度), 返回一个正方形区域, 该区域由4个坐标点表示
func Uint32ToCoords(x uint32, y uint32) (lng, lat float64) {
	const deltaAngle = 0.0000000000001 //角度偏移,为了计算方便;//不用DELTAANGLE时,在（90/180）会有问题

	var precisionLat = (180.0 + deltaAngle) / float64(1<<(geoBinBitsCount/2)) //纬度的刻度,每个刻度值有多少度
	var precisionLng = (360.0 + deltaAngle) / float64(1<<(geoBinBitsCount/2)) //经度的刻度,每个刻度值有多少度

	lng = precisionLng*float64(x) - 180.0
	lat = precisionLat*float64(y) - 90.0

	return
}

//Geo2Str 转成字母显示格式，只有标准的5bit一组的数据才能转！
func Geo2Str(x uint64) string {
	buffer := make([]byte, 0, geoBinBitsCount/5)
	start := true
	for i := int(geoBinBitsCount) - 5; i >= 0; i -= 5 {
		charIndex := x >> uint(i) & 0x1F
		if !start && charIndex == 0 {
			continue
		}
		start = true
		buffer = append(buffer, base32[charIndex])
	}
	return string(buffer)
}

//交错和反交错

/* Interleave lower bits of x and y, so the bits of x
 * are in the even positions and bits from y in the odd;
 * x and y must initially be less than 2**32 (65536).
 * From:  https://graphics.stanford.edu/~seander/bithacks.html#InterleaveBMN
 */

var (
	B = []uint64{0x5555555555555555, 0x3333333333333333, 0x0F0F0F0F0F0F0F0F,
		0x00FF00FF00FF00FF, 0x0000FFFF0000FFFF, 0x00000000FFFFFFFF}
	S = []uint64{0, 1, 2, 4, 8, 16} //B、S放到函数内每次都生成需要额外5.4ns/op
)

//偶数位放经度，奇数位放纬度
func interleave64(xlo, ylo uint32) (b uint64) {
	x := uint64(xlo)
	y := uint64(ylo)

	x = (x | (x << S[5])) & B[4]
	y = (y | (y << S[5])) & B[4]

	x = (x | (x << S[4])) & B[3]
	y = (y | (y << S[4])) & B[3]

	x = (x | (x << S[3])) & B[2]
	y = (y | (y << S[3])) & B[2]

	x = (x | (x << S[2])) & B[1]
	y = (y | (y << S[2])) & B[1]

	x = (x | (x << S[1])) & B[0]
	y = (y | (y << S[1])) & B[0]

	return y | (x << 1)
}

/* reverse the interleave process
 * derived from http://stackoverflow.com/questions/4909263
 */
func deinterleave64(interleaved uint64) (uint32, uint32) { // uint64 {
	x := interleaved >> 1
	y := interleaved

	x = (x | (x >> S[0])) & B[0]
	y = (y | (y >> S[0])) & B[0]

	x = (x | (x >> S[1])) & B[1]
	y = (y | (y >> S[1])) & B[1]

	x = (x | (x >> S[2])) & B[2]
	y = (y | (y >> S[2])) & B[2]

	x = (x | (x >> S[3])) & B[3]
	y = (y | (y >> S[3])) & B[3]

	x = (x | (x >> S[4])) & B[4]
	y = (y | (y >> S[4])) & B[4]

	x = (x | (x >> S[5])) & B[5]
	y = (y | (y >> S[5])) & B[5]

	return uint32(x), uint32(y)
}

//Coords2Area 根据中心点和半径获取涵盖该区域内的所有geo块的geohash值
//抽象为二维平面来计算,横轴x,对应经度；纵轴y,对应纬度
//长度:x --> longitude,y --> latitude,r=radius;
//bits 表示需要多少bit表示一个geo方块
func Area2Geos(lngL, lngR, latU, latD float64, bits uint32) (geohashS []uint64) {
	if lngL > 180 || lngL < -180 ||
		lngR > 180 || lngR < -180 ||
		latU > 90 || latU < -90 ||
		latD > 90 || latD < -90 {
		return
	}
	geoLngL, geoLatU := Coords2Uint32(lngL, latU) //计算相关的uint64坐标
	geoLngR, geoLatD := Coords2Uint32(lngR, latD)
	deltaLng := geoLngR - geoLngL
	deltaLat := geoLatU - geoLatD
	if deltaLng < 0 { //在本初子午线交界处
		deltaLng = 0xFFFFFFFF + deltaLng
	}
	n := ((deltaLng >> bits) + 1) * ((deltaLat >> bits) + 1)
	if n > 1024*1024 {
		//需要的空间太大了，不处理
		return
	}
	geohashS = make([]uint64, 0, n) //本行申请空间,耗时和申请大小相关,50~250ns
	for y := geoLatD >> bits; y <= geoLatU>>bits; y++ {
		//如果在这里合并区间，效率更高，也避免申请多余的空间
		for x := geoLngL >> bits; x <= geoLngR>>bits; x++ {
			geohash := interleave64(x, y)
			geohashS = append(geohashS, geohash<<(2*bits))
		}
	}
	return geohashS
}

type uint64Slice []uint64

func (p uint64Slice) Len() int           { return len(p) }
func (p uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

//mergeGeos 把连续的块合并起来
//deviation: 允许的偏差，即为了连续(比如减少查询次数)，适当允许多查询的块(里面的内容都不合格),
//虽然过滤元素变多,但是查询次数变少,可以适当允许
func mergeGeos(in []uint64, deviation uint64, bits uint32) (outL, outR []uint64) {
	if len(in) <= 0 {
		return
	}
	if deviation < 1 {
		deviation = 1
	}
	deviation = deviation << bits //偏差值
	//deviationOne := (uint64(1) << bits) - 1 //这个是左闭右闭区间,所以最后一个，，，
	deviationOne := uint64(1) << bits //所有区间都是左闭右开区间,所以最后一个，，，

	sort.Sort(uint64Slice(in))
	outR = make([]uint64, len(in))
	j := 0
	outR[0] = in[0] + deviationOne
	for i := 1; i < len(in); i++ {
		if in[i]-in[i-1] <= deviation {
			outR[j] = in[i] + deviationOne
			continue
		}
		j++
		in[j] = in[i]
		outR[j] = in[i] + deviationOne
	}
	outL = in[:j+1]
	outR = outR[:j+1]
	return
}

//Area2Geos2 坐标 + 半径 --> 范围
func Area2Geos2(area Area) (geohashS []uint64) {
	var lngL, lngR, latU, latD float64

	//先找到最左、最右、最上、最下4个点
	lngL, latD = area.area[0].Lng, area.area[0].Lat
	for _, v := range area.area {
		if lngL > v.Lng {
			lngL = v.Lng
		}
		if lngR < v.Lng {
			lngR = v.Lng
		}
		if latD > v.Lat {
			latD = v.Lat
		}
		if latU > v.Lat {
			latU = v.Lat
		}
	}
	return Area2Geos(lngL, lngR, latU, latD, 15)
}

//R2MergeGeos2 (radius:半径) 坐标 + 半径 --> 范围 --> 合并范围
func R2MergeGeos2(geohash uint64, r float64, bits uint32) (outL, outR []uint64) {
	x, y := deinterleave64(geohash)
	lng, lat := Uint32ToCoords(x, y)
	lngL, lngR, latU, latD := R2Area(lng, lat, r)
	geohashS := Area2Geos(lngL, lngR, latU, latD, 32-bits)
	return mergeGeos(geohashS, 3, 64-bits)
}

//R2MergeGeos (radius:半径) 坐标 + 半径 --> 范围 --> 合并范围
func R2MergeGeos(lng, lat, r float64, bits uint32) (outL, outR []uint64) {
	lngL, lngR, latU, latD := R2Area(lng, lat, r)
	geohashS := Area2Geos(lngL, lngR, latU, latD, 32-bits)
	return mergeGeos(geohashS, 3, 64-bits)
}

//R2Geos2 (radius:半径) 坐标 + 半径 --> 范围
func R2Geos2(lng, lat, r float64) (geohashS []uint64) {
	lngL, lngR, latU, latD := R2Area(lng, lat, r)
	return Area2Geos(lngL, lngR, latU, latD, 15)
}

//R2Area 长度变成角度差; 经度+-180°处，需要特殊考虑
func R2Area(lng, lat, r float64) (lngL, lngR, latU, latD float64) {
	cosY := math.Cos(math.Pi * lat / 180) //该纬度上的纬线圈对应的角度的cos()
	C := earthClat * cosY                 //earthClat 是赤道的长度，C是纬线圈的长度(周长和半径成正比)
	deltaLng := 360.0 * r / C
	deltaLat := 180.0 * r / earthClng
	lngL = lng - deltaLng
	if lngL < -180.0 {
		lngL += 180.0
	}
	lngR = lng + deltaLng
	if lngL > 180.0 {
		lngL -= 180.0
	}
	latD = lat - deltaLat
	if latD < -90.0 {
		latD = -90.0
	}
	latU = lat + deltaLat
	if latU > 90.0 {
		latU = 90.0
	}
	return
}

//Dist 计算两坐标间的距离(distance)
func Dist(p1, p2 Coords) (l float64) {
	calculateY := func(p1, p2 Coords) float64 {
		deltaY := p1.Lat - p2.Lat
		if deltaY < 0 {
			deltaY = -deltaY
		}
		return deltaY / 180.0 * earthClng //180°时，即经线长度: earthClng
	}
	calculateX := func(p1, p2 Coords) float64 {
		alphaY := p1.Lat
		deltaX := p1.Lng - p2.Lng
		if deltaX < 0 {
			deltaX = -deltaX
		}
		cosY := math.Cos(math.Pi * alphaY / 180) //该纬度上的纬线圈对应的角度的cos()
		C := earthClat * cosY                    //earthClat 是赤道的长度，C是纬线圈的长度(周长和半径成正比)
		return deltaX / 360.0 * C
	}

	x := calculateX(p1, p2)
	y := calculateY(p1, p2)

	return math.Sqrt(x*x + y*y)
}

//Dist2 计算两坐标间的距离(distance)
func Dist2(p1, p2 Coords) (l float64) {

	deltaY := p1.Lat - p2.Lat
	// if deltaY < 0 {
	// 	deltaY = -deltaY
	// }
	y := deltaY / 180.0 * earthClng //180°时，即经线长度: earthClng

	deltaX := p1.Lng - p2.Lng
	// if deltaX < 0 {
	// 	deltaX = -deltaX
	// }
	cosY := math.Cos(math.Pi * p1.Lat / 180.0) //该纬度上的纬线圈对应的角度的cos()
	C := earthClat * cosY                      //earthClat 是赤道的长度，C是纬线圈的长度(周长和半径成正比)
	x := deltaX / 360.0 * C

	return math.Sqrt(x*x + y*y)
}

//	earthClat = 40075.2     //地球赤道周长40075.2km
//	earthClng = 40037.0 / 2 //地球经线长度40037.0/2km
//Len2Bits 根据长度，获取能包含改长度的Geohash的bit长度
//  只在初始化的时候调用一次，对性能要求不高; l 单位ikm
func Len2Bits(l float64) (bits int) {
	bits = 1
	a := earthClat / 2.0       //最初的geohash块的长(这是赤道的长,纬度60°的时候,和b大小差不多)
	b := earthClng / 2.0 / 2.0 //最初的geohash块的宽
	for b > l {
		a = a / 2.0
		b = b / 2.0
		bits++
	}
	bits-- //保证大于
	return
}
