package algorithms

import (
//"fmt"
)

//BinSearch 区间二分查找，左必右开区间，已从小到大排序
func BinSearch(v []uint64, k uint64) int {
	left, right, mid := 1, len(v), 0
	for left < right-1 {
		mid = (left + right) / 2 // mid向下取整
		if k < v[mid] {
			right = mid - 1 // 如果当前元素大于k，那么把right指针移到mid - 1的位置
		} else if k >= v[mid] {
			left = mid + 1 // 如果当前元素小于k，那么把left指针移到mid + 1的位置
		}
	}
	return left //返回区间左下标
}
