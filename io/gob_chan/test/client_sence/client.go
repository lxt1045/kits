package main

import (
	"encoding/json"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lxt1045/geocache/msg"
	"github.com/lxt1045/kits/geo"
	"github.com/lxt1045/kits/io/channel"
	"github.com/lxt1045/kits/io/gob"
	"github.com/lxt1045/kits/log"
)

var (
	goN           int
	userNPerSence int
	SenceN        int
	userID        uint64
	goroutineMax  int
	exitFlags     int32

	sendPkgN           int64
	recvPkgN           int64
	errorN             int64
	timeRecvDeleyT     int64
	pkgRecvDelay100msN int32
	pkgRecvDelay1sN    int32

	queueSendSlice []Conn
)

type Conn struct {
	ch     channel.IChanN
	userid uint64
}

func init() {
	runtime.GOMAXPROCS(4)
	goN = runtime.GOMAXPROCS(0)
	userNPerSence = 3000
	SenceN = 4
	goroutineMax = 4
}

func main() {
	var wg sync.WaitGroup
	go statistics()

	for i := 0; i < SenceN; i++ {
		//conn, err := net.Dial("tcp", "127.0.0.1:8520")
		conn, err := net.Dial("tcp", "192.168.31.90:8520")
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
		go recv(recvCh, uint64(i))

		for j := 1; j <= userNPerSence; j++ {
			queueSendSlice = append(queueSendSlice, Conn{
				ch:     sendCh,
				userid: uint64(i)*uint64(userNPerSence) + uint64(j),
			})
		}

		if i%1 == 0 {
			log.Infof("start:%d", i)
		}
	}

	for i := 0; i < 1; i++ { // goroutineMax
		wg.Add(1)
		go send() //在单协程中发送
	}

	time.Sleep(time.Second)
	wg.Wait()
	//wg.Wait()
}

func recv(recvCh channel.IChanN, userid uint64) {
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
		//*
		imsg, ok := iface.(gob.IMsg)
		if !ok {
			atomic.AddInt64(&errorN, 1)
			log.Errorf("msg type Error, type:[%v],value:[%v]", reflect.TypeOf(imsg), imsg)
			continue
		} //*/

		//

		nearbyList, ok := iface.(*msg.NearbyList)
		if ok {
			s, _ := json.Marshal(nearbyList)
			//if nearbyList.UserID%1000 == 1 {
			if nearbyList.UserID == 1 {
				log.Infof("user[%d] get msg, type:[%v],value:[%v]", nearbyList.UserID, reflect.TypeOf(nearbyList), string(s))
			}
			atomic.AddInt64(&recvPkgN, 1)
			now := time.Now().UnixNano()
			deta := now - int64(nearbyList.Timestamp)
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
			continue
		}

		//性能测试时需要用到
		ack, ok := iface.(*msg.Move)
		if ok {
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
}

func send() {
	arg := msg.Move{
		UserID:    userID,
		Timestamp: time.Now().UnixNano(),
		Geohash:   geo.Coords2Geo(116.3906, 39.92324),
		X:         10.888,
		Y:         88.1,
	}
	lastMs := time.Now().UnixNano()
	for i := 0; ; i++ {
		if atomic.LoadInt32(&exitFlags) < 0 {
			break
		}
		arg.X += 0.001 + rand.Float32()/10.0
		arg.Y += 0.001 + rand.Float32()/10.0

		for _, q := range queueSendSlice {
			arg1 := arg
			arg1.Timestamp = time.Now().UnixNano()
			arg1.UserID = q.userid
			arg1.X += 0.02 + rand.Float32()/10.0
			arg1.Y += 0.03 + rand.Float32()/10.0
			full, closed := q.ch.Send(&arg1)
			if full || closed {
				if closed {
					return
				}
				log.Info("full")
				continue
			}
			atomic.AddInt64(&sendPkgN, 1)
		}
		/*
			if i == 10 {
				for _, q := range queueSendSlice {
					var arg msg.Logout
					arg.Timestamp = time.Now().UnixNano()
					arg.UserID = q.userid
					full, closed := q.ch.Send(&arg)
					if full || closed {
						if closed {
							return
						}
					}
					return
				}
				return
			}//*/

		if i%10 == 0 {
			nowMs := time.Now().UnixNano()
			//1s发10000个数据
			if delta := lastMs + int64(time.Second)*1 - nowMs; delta > 0 {
				log.Infof("............................\nleft:%dms\n", delta/int64(time.Millisecond))
				time.Sleep(time.Duration(delta))
			}
			lastMs = time.Now().UnixNano()
			//log.Info("............................\n\n")
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
		fpsRecv := float64(recvN-lastRecvN) / (float64(now-lastTime) / float64(time.Second))
		fpsSend := float64(sendN-lastSendN) / (float64(now-lastTime) / float64(time.Second))
		if lastTime == 0 {
			lastTime = now
			lastSendN = sendN
			lastRecvN = recvN
			lastDelayT = delayT
			<-time.After(time.Millisecond * 1000 * 3)
			continue
		}
		if recvN-lastRecvN > 0 {
			log.Infof("lost:%.3f, send fps:%.3fk, recv fps:%.3fk, recv:%dk, delay avg:%.3f,delay all avg:%.3f, >100ms:%.3f, >1s:%.3f, err:%d",
				float64((sendN-lastSendN)-(recvN-lastRecvN))/float64(sendN-lastSendN), fpsSend/1000.0, fpsRecv/1000.0,
				recvN/1000, float64((delayT-lastDelayT)/int64(time.Millisecond))/float64(recvN-lastRecvN),
				float64(delayT/int64(time.Millisecond))/float64(recvN),
				float64(delay100msN)/float64(recvN), float64(delay1sN)/float64(recvN), errN)
		}
		lastTime = now
		lastSendN = sendN
		lastRecvN = recvN
		lastDelayT = delayT

		<-time.After(time.Millisecond * 1000 * 3)
	}
}
