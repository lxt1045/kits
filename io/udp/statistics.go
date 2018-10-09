package udp

import (
	"sync/atomic"
	//"time"
)

var stats statistic

// 服务器的统计信息
type statistic struct {
	StartTime string // 启动时间
	conn      uint64 // 连接请求数

	packageSum uint64 // 所有包统计
	delayCount uint64 // 延时包的数量
	delaySum   uint64 // 延时总时长，用于计算平均值; 拿到发出包的时候，会有一个参数表示何时收到包

}

func (p *statistic) ConnInc() {
	atomic.AddUint64(&p.conn, 1)
}
func (p *statistic) PackageInc() {
	atomic.AddUint64(&p.packageSum, 1)
}

func String() string {

	return ""
}
