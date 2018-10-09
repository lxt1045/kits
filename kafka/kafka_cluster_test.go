package kafka

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/lxt1045/kits/log"
)

const (
	topicTest = "kafka_cluster_test3344"
)

func init() {
}

func TestRecvCluster(t *testing.T) {
	var wg sync.WaitGroup
	c := Config{
		ServerList: []string{"192.168.31.233:9092"},
		RetryTime:  3,
		OffSet:     -2,
		Group:      "test5",
	}
	logT := log.NewLogTrace(0, 0, 0) //
	client, e := NewClient([]string{topicTest}, c, logT)
	if e != nil {
		t.Error(e)
		return
	}

	wg.Add(1)
	go func() {
		for i := 0; i < 10; i++ {
			v, e := client.RecvMsg()
			if e != nil {
				t.Error(e)
			}
			t.Log("recv msg:", string(v))
		}
		wg.Done()
	}()
	runtime.Gosched()
	wg.Add(1)
	go func() {
		tmf := time.Now().Format("2006-01-02 15:04:05")
		for i := 0; i < 10; i++ {
			msg := fmt.Sprintf("%s test message index:%d", tmf, i)
			e := client.SendMsg(topicTest, []byte(msg))
			if e != nil {
				t.Error(e)
			}
			t.Logf("send msg: %s\n", msg)
		}
		wg.Done()
	}()

	wg.Wait()
	t.Log("completed！")
}
