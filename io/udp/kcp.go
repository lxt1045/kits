package udp

import (
	"container/list"
	"net"
	"sync"
	"sync/atomic"
	"time"

	//kcp "github.com/xtaci/kcp-go"
	"github.com/lxt1045/kits/log"
	"github.com/lxt1045/sence/conf"
)

var (
	kcpUpdataStatus int32 //用于保证全局之调用一次 kcpUpdata()

	updataList  *list.List //用于存放kcp，然后循环调用kcp.Updata(); updata()处理较快，单goroutine即可
	updataAddCh chan *Conn //单协程操作list，不需要加锁，通过chan同步(加入、删除)
	updataDelCh chan *list.Element
)

func init() {
	updataList = list.New()
	updataAddCh = make(chan *Conn, ioChanLenDefault)
	updataDelCh = make(chan *list.Element, ioChanLenDefault)
}

//kcpUpdataRun保证kcpUpdata()全局以goroutine形式启动一次
func kcpUpdata(logT *log.TraceInfoST, wg *sync.WaitGroup, conf *conf.Config) {
	if atomic.AddInt32(&kcpUpdataStatus, 1) > 1 {
		atomic.AddInt32(&kcpUpdataStatus, -1)
		logT.Critical(" io.Start() Can only be called once")
		return
	}
	if conf.KcpUpdataTime == 0 {
		conf.KcpUpdataTime = 10
	}
	if conf.MaxIdleTime == 0 { //< 10 {
		conf.MaxIdleTime = 10 //最少10s
	}
	for {
		deadTime := time.Now().Unix() - conf.MaxIdleTime //认定僵死状态的时间
		select {
		case <-time.After(time.Millisecond * time.Duration(conf.KcpUpdataTime)):
			for e := updataList.Front(); e != nil; e = e.Next() {
				if e.Value == nil {
					continue
				}
				conn, ok := e.Value.(*Conn)
				if !ok {
					continue
				}
				if atomic.LoadInt64(&conn.refleshTime) < deadTime {
					removeConn(conn) //该连接已经凉了
					continue
				}
				conn.Lock() //kcp.Check() 和 kcp.Update() 线程不安全，，，
				if conn.kcp != nil && conn.kcp.Check() <= uint32(time.Now().UnixNano()/int64(time.Millisecond)) {
					conn.kcp.Update() //kcp.Check()比kcp.Update()轻量
				}
				conn.Unlock()
			}
		case conn := <-updataAddCh:
			conn.Lock()
			if conn.listE != nil {
				log.Errorf("conn.listE is not nil, remove before PushBack, conn:%v", conn)
				updataList.Remove(conn.listE)
			}
			conn.listE = updataList.PushBack(conn)
			conn.Unlock()
		case ekcp := <-updataDelCh:
			updataList.Remove(ekcp)
		}
	}
	wg.Done()
	log.Critical("udp.kcpUpdata exit")
}

//kcp send()之后,会在flush()的时候调用这里的函数,将数据包通过udp 发送出去
func kcpOutoput(listener *net.UDPConn, addr *net.UDPAddr, logT *log.TraceInfoST) (f func([]byte, int)) {
	var bufSend []byte
	f = func(buf []byte, size int) {
		//发送数据成熟时的异步回调函数，成熟数据为：buf[:size]，在这里把数据通过UDP发送出去
		//func copy(dst, src []Type) int //The source and destination may overlap.
		if len(buf) > size {
			bufSend = buf
		} else {
			bufSend = make([]byte, size+1)
		}
		copy(bufSend[1:], buf[:size]) //copy的src和dst可以重叠，所以直接使用
		encode8u(bufSend, Guar_YES)   //加个header信息
		n, err := listener.WriteToUDP(bufSend[:size+1], addr)
		if err != nil || n != size+1 {
			logT.Criticalf("error during kcp send:%v,n:%d\n", err, n)
		}
	}
	return
}
