package algorithms

import (
//"fmt"
)

//BinSearch 区间二分查找，左必右开区间，已从小到大排序
func BinSearch(n int, f func(int) bool) int {
	left, right, mid := 0, n, 0
	for left < right-1 {
		mid = (left + right) / 2 // mid向下取整
		if f(mid) {              // k < v[mid]
			right = mid // 如果当前元素大于k，那么把right指针移到mid - 1的位置
		} else {
			left = mid // 如果当前元素小于k，那么把left指针移到mid + 1的位置
		}
	}
	return left //返回区间左下标
}

//BinSearchUint64Raw 性能比 BinSearchUint64 好，差不多两倍！
func BinSearchUint64Raw(v []uint64, k uint64) int {
	left, right, mid := 0, len(v), 0
	for left < right-1 {
		mid = (left + right) / 2 // mid向下取整
		if k < v[mid] {
			right = mid // 如果当前元素大于k，那么把right指针移到mid - 1的位置
		} else {
			left = mid // 如果当前元素小于k，那么把left指针移到mid + 1的位置
		}
	}
	return left //返回区间左下标
}

//BinSearch 区间二分查找，左必右开区间，已从小到大排序
func BinSearchUint64(v []uint64, k uint64) int {
	// left, right, mid := 0, len(v), 0
	// for left < right-1 {
	// 	mid = (left + right) / 2 // mid向下取整
	// 	if k < v[mid] {
	// 		right = mid // 如果当前元素大于k，那么把right指针移到mid - 1的位置
	// 	} else {
	// 		left = mid // 如果当前元素小于k，那么把left指针移到mid + 1的位置
	// 	}
	// }
	// return left //返回区间左下标
	return BinSearch(len(v), func(i int) bool { return k < v[i] })
}

//BinSearchFloat64 针对float64
func BinSearchFloat64(v []float64, k float64) int {
	left, right, mid := 0, len(v), 0
	for left < right-1 {
		mid = (left + right) / 2 // mid向下取整
		if k < v[mid] {
			right = mid // 如果当前元素大于k，那么把right指针移到mid - 1的位置
		} else {
			left = mid // 如果当前元素小于k，那么把left指针移到mid + 1的位置
		}
	}
	return left //返回区间左下标
}
