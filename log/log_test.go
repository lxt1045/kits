package log_test

import (
	"testing"
	"time"

	"github.com/lxt1045/kits/kafka"
	"github.com/lxt1045/kits/log"
)

func init() {
	log.Init(
		"g_log_test",
		&log.Config{
			Topic: "logs",
			Level: 5,
			Print: 1,
		},
		&kafka.Config{
			ServerList:  []string{"192.168.31.233:9092"},
			SendTimeOut: 500,
			RetryTime:   1,
			RetryWait:   100,
			Print:       1,
			OffSet:      -1, // 启动时，接收消息的位置， -1 从最新的开始读， -2：从上次的问题开始读
			Group:       "g_log_test",
		},
	)
}

func TestLog(t *testing.T) {
	logT := log.NewLogTrace(0, 0, 0) //
	logT.Debug("*****************")
	logT.Debug("AAAAAAAAAAAAAAA")
	logT.Debug("BBBBBBBBBBBBBBBBBBB")
	t.Log("completed！")
	time.Sleep(time.Second * 1)
	time.Sleep(time.Second * 1)
	time.Sleep(time.Second * 3)
}
