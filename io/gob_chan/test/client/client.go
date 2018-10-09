package main

import (
	"github.com/lxt1045/kits/io/channel"
	"net"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lxt1045/geocache/invoke"
	"github.com/lxt1045/geocache/invoke/msg"
	"github.com/lxt1045/kits/io/gob"
	"github.com/lxt1045/kits/log"
)

var (
	goN          int
	goroutineMax int
	exitFlags    int32

	sendPkgN           int64
	recvPkgN           int64
	errorN             int64
	timeRecvDeleyT     int64
	pkgRecvDelay100msN int32
	pkgRecvDelay1sN    int32

	queueSendSlice []channel.IChanN
)

func init() {
	goroutineMax = 10
	runtime.GOMAXPROCS(4)
	goN = runtime.GOMAXPROCS(0)
}

func main() {
	var wg sync.WaitGroup
	go statistics()

	for i := 0; i < goroutineMax; i++ {
		// "192.168.31.90:8080")
		conn, err := net.Dial("tcp", "127.0.0.1:8080")
		if err != nil {
			log.Error(err)
			continue
		}
		wg.Add(1)
		//使用goroutine单独处理rpc连接请求
		recvCh, sendCh, err := gob.NewPeer().ServeConn(conn)
		if err != nil {
			log.Error("ServeConn(conn) get err:%v", err)
			continue
		}
		go recv(recvCh)
		queueSendSlice = append(queueSendSlice, sendCh)

		if i%100 == 0 {
			log.Infof("start:%d", i)
		}
	}

	for i := 0; i < goN/4; i++ {
		wg.Add(1)
		go send() //在单协程中发送
	}

	time.Sleep(time.Second)
	wg.Wait()
	//wg.Wait()
}

func recv(recvCh channel.IChanN) {
	//func(reply interface{}, err error) {
	//recv
	i := 0
	closed := false
	var ifaces []interface{}
	for {
		if i >= len(ifaces) {
			i = 0
			ifaces, closed = recvCh.Recv(16, 1) //一次获取多个，可以减少锁竞争
			if closed {
				break //chan 已关闭,则退出
			}
		}
		iface := ifaces[i]
		i++

		ack, ok := iface.(*invoke.Ack)
		if !ok {
			atomic.AddInt64(&errorN, 1)
			log.Errorf("msg type Error, type:[%v],value:[%v]", reflect.TypeOf(iface), iface)
			continue
		}
		//
		if ack.UserID != 8881 {
			atomic.AddInt64(&errorN, 1)
			log.Info(ack)
			continue
		}

		atomic.AddInt64(&recvPkgN, 1)
		now := time.Now().UnixNano()
		deta := now - int64(ack.Timestamp)
		if deta < 0 {
			atomic.AddInt64(&errorN, 1)
			return
		}
		if deta > int64(time.Millisecond*100) {
			atomic.AddInt32(&pkgRecvDelay100msN, 1)
		}
		if deta > int64(time.Second) {
			atomic.AddInt32(&pkgRecvDelay1sN, 1)
		}
		atomic.AddInt64(&timeRecvDeleyT, deta)
	}
}

func send() {
	arg := invoke.Ack{msg.Ack{UserID: 8881, Timestamp: time.Now().UnixNano()}}
	lastMs := time.Now().UnixNano()
	for i := 0; ; {
		if atomic.LoadInt32(&exitFlags) < 0 {
			break
		}

		for _, q := range queueSendSlice {
			arg.Timestamp = time.Now().UnixNano()
			full, closed := q.Send(&arg)
			if full || closed {
				if closed {
					return
				}
				//log.Info("full")
			}
			i++
			atomic.AddInt64(&sendPkgN, 1)

			if i%100000 == 0 {
				nowMs := time.Now().UnixNano()
				//1s发10000个数据
				if delta := lastMs + int64(time.Second) - nowMs; delta > 0 {
					time.Sleep(time.Duration(delta))
				}
				lastMs = time.Now().UnixNano()
			}
		}
	}
	//wg.Done()
}
func statistics() {
	var lastSendN, lastRecvN, lastDelayT int64
	var lastTime int64

	for {
		recvN := atomic.LoadInt64(&recvPkgN)
		sendN := atomic.LoadInt64(&sendPkgN)
		delayT := atomic.LoadInt64(&timeRecvDeleyT)
		delay100msN := atomic.LoadInt32(&pkgRecvDelay100msN)
		delay1sN := atomic.LoadInt32(&pkgRecvDelay1sN)
		errN := atomic.LoadInt64(&errorN)

		now := time.Now().UnixNano()
		fps := float64(recvN-lastRecvN) / (float64(now-lastTime) / float64(time.Second))
		fpsSend := float64(sendN-lastSendN) / (float64(now-lastTime) / float64(time.Second))
		if lastTime == 0 {
			lastTime = now
			lastSendN = sendN
			lastRecvN = recvN
			lastDelayT = delayT
			<-time.After(time.Millisecond * 1000 * 3)
			continue
		}
		log.Infof("lost:%.3f, send fps:%.3fk, recv fps:%.3fk, recv:%dk, delay avg:%.3f,delay all avg:%.3f, >100ms:%.3f, >1s:%.3f, err:%d",
			float64((sendN-lastSendN)-(recvN-lastRecvN))/float64(sendN-lastSendN), fpsSend/1000, fps/1000.0,
			recvN/1000, float64((delayT-lastDelayT)/int64(time.Millisecond))/float64(recvN-lastRecvN),
			float64(delayT/int64(time.Millisecond))/float64(recvN),
			float64(delay100msN)/float64(recvN), float64(delay1sN)/float64(recvN), errN)
		lastTime = now
		lastSendN = sendN
		lastRecvN = recvN
		lastDelayT = delayT

		<-time.After(time.Millisecond * 1000 * 3)
	}
}
