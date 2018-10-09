package log

import (
	"fmt"
	"sync/atomic"
)

type CountST struct {
	recvLogNum uint64
	sendLogNum uint64
	sendErrNum uint64
	bufErrnum  uint64
}

var g_count CountST

// 获取日志的统计
func Count() string {
	return fmt.Sprintf("%+v", g_count)
}

func (my *CountST) RecvLog() {
	atomic.AddUint64(&my.recvLogNum, 1)
}

func (my *CountST) SendLog() {
	atomic.AddUint64(&my.sendLogNum, 1)
}

func (my *CountST) SendErr() {
	atomic.AddUint64(&my.sendErrNum, 1)
}

func (my *CountST) BufErr() {
	atomic.AddUint64(&my.bufErrnum, 1)
}
