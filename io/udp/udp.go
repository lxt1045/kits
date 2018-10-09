package udp

import (
	"context"
	"encoding/json"
	"fmt"
	synclog "log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lxt1045/kits/io/channel"
	"github.com/lxt1045/kits/log"
	"github.com/lxt1045/sence/conf"
	kcp "github.com/xtaci/kcp-go"
)

//IP协议规定路由器最少能转发：512数据+60IP首部最大+4预留=576字节,即最少可以转发512-8=504字节UDP数据
//内网一般1500字节,UDP：1500-IP(20)-UDP(8)=1472字节数据
const (
	udpBufLen        = 1472    //UDP发送、接收，kcp发送、接收缓冲的大小
	ioChanLenDefault = 1024000 //当config没有设置时，使用此默认值
)

var (
	startStatus         int32 //用于保证只启动一次
	recvN, lostN, sendN int64 //用于统计信息
)

// Start udp接口启动的时候执行，inCh, outCh是输入、输出chan，为了和后端处理模块解耦
func Start(ctx context.Context, wg *sync.WaitGroup, logT *log.TraceInfoST, conf *conf.Config) (
	recvCh, sendCh channel.IChanN, err error) {
	if atomic.AddInt32(&startStatus, 1) > 1 {
		atomic.AddInt32(&startStatus, -1)
		logT.Critical(" io.Start() Can only be called once")
		return
	}
	//go statistics()          //用于性能统计
	go kcpUpdata(logT, wg, conf) //每一个kcp连接都需要定期巧用updata()以驱动kcp循环
	wg.Add(1)

	ioChanSize := conf.IOChanSize
	if ioChanSize <= 0 {
		ioChanSize = ioChanLenDefault
	}
	recvCh = channel.NewChanN(ioChanSize)
	sendCh = channel.NewChanN(ioChanSize)
	//conf.Listen = "127.0.0.1:8501"
	udpAddr, err := net.ResolveUDPAddr("udp", conf.Listen)
	if err != nil {
		logT.Errorf("error:%v", err)
		return
	}
	listener, err := net.ListenUDP("udp", udpAddr)
	if err != nil || listener == nil {
		logT.Errorf("error:%v", err)
		return
	}
	logT.Infof("udp listening on: %s", udpAddr)

	go func(ctx context.Context) { //判断ctx是否被取消了，如果是就退出
		<-ctx.Done()
		listener.Close()
		recvCh.Close()
		sendCh.Close()
		wg.Done()
	}(ctx)
	wg.Add(1)

	//多线收发数据，可以提高数据收发速度
	for i := 0; i < conf.CPUNUM; i++ {
		go send(sendCh, listener, wg, logT) //发送数据
		go recv(recvCh, listener, wg, logT) //接收数据
		wg.Add(2)
	}
	return
}

func recv(recvCh channel.IChanN, listener *net.UDPConn, wg *sync.WaitGroup, logT *log.TraceInfoST) {
	parseMsg := func(buf []byte, n int, connID uint64, logT *log.TraceInfoST) (pkg Pkg, err error) {
		var h Header
		_, err = h.Deserialize(buf)
		pkg.Msg, err = newMsg(h.Type)
		if err != nil {
			logT.Errorf("newMsg got error:%v, header:%v", err, h)
			err = fmt.Errorf("newMsg got error:%v, header:%v", err, h)
			return
		}
		pkg.ConnID = connID
		if int(h.Len) > n-MSG_HEADER_LEN {
			atomic.AddInt64(&lostN, 1)
			err = fmt.Errorf("error during read,h.Len > n-MSG_HEADER_LEN, h:%v, n:%d", h, n)
			return
		}
		err = proto.Unmarshal(buf[MSG_HEADER_LEN:n], pkg.Msg)

		pkg.LogT = logT //log.NewLogTrace(0, connID, 0)

		js, _ := json.MarshalIndent(&h, "", "\t")
		synclog.Printf("\n\n+++++++++++++++++++++++++\nheader:%v, \nmsg:%v, \n\nbody:%x\n\n+++++++++++++++++++++++++\n\n",
			string(js), pkg.Msg, buf[MSG_HEADER_LEN:n])
		return
	}
	var (
		addr    *net.UDPAddr
		err     error
		n       = 0
		bufRecv = make([]byte, udpBufLen)
		pkgs    = make([]interface{}, 0, 8)
	)
	for {
		pkgs = pkgs[:0]
		n, addr, err = listener.ReadFromUDP(bufRecv)
		if err != nil || n <= 0 {
			logT.Criticalf("error during read:%v, n:%d", err, n)
			continue
			//break
		}
		atomic.AddInt64(&recvN, 1)
		conn := addr2Conn(addr)
		logMsg := log.NewLogTrace(0, int64(conn.connID), 0)     //为了方便追踪，每个Msg都创建一个LogTrace
		atomic.StoreInt64(&conn.refleshTime, time.Now().Unix()) //这个不需要线程安全，并发更新谁成功了都是可以接受的,但是检查的时候，需要最新的值

		var guar uint8
		decode8u(bufRecv, &guar) //先确认是否是可靠传输的消息

		fmt.Printf("\n\nget UDP msg,  guar == Guar_YES:%v\n", guar == Guar_YES)

		if guar == Guar_YES {
			conn.Lock()
			if conn.kcp == nil {
				var conv uint32
				decode32u(bufRecv[1:], &conv) //获取数据包的conv
				conn.kcp = kcp.NewKCP(conv, kcpOutoput(listener, addr, logMsg))
				conn.kcp.WndSize(128, 128) //设置最大收发窗口为128
				// 第1个参数 nodelay-启用以后若干常规加速将启动
				// 第2个参数 interval为内部处理时钟，默认设置为 10ms
				// 第3个参数 resend为快速重传指标，设置为2
				// 第4个参数 为是否禁用常规流控，这里禁止
				conn.kcp.NoDelay(0, 10, 0, 0) // 默认模式
				//conn.kcp.NoDelay(0, 10, 0, 1) // 普通模式，关闭流控等
				//conn.kcp.NoDelay(1, 10, 2, 1) // 启动快速模式
				updataAddCh <- conn //通知updata()协程增加kcp
			}
			//以下操作会将bufRecv的数据复制到kcp的底层缓存，所以bufRecv可以快速重用
			m := conn.kcp.Input(bufRecv[1:n], true, false) //bufRecv[1:], true, false：数据，正常包，ack延时发送
			conn.Unlock()
			for m >= 0 {
				//这里要确认一下，kcp.Recv()是否要经过update()驱动，如果要驱动，则不能在这里处理
				conn.Lock()
				m = conn.kcp.Recv(bufRecv)
				conn.Unlock()
				if m <= 0 {
					if m == -3 {
						bufRecv = make([]byte, len(bufRecv)*2)
						m = 0
						continue
					}
					break
				}
				pkg, err := parseMsg(bufRecv, m, conn.connID, logMsg)
				if err != nil {
					logT.Error(err)
					continue
				}
				pkg.Guar = true
				pkgs = append(pkgs, pkg)
			}
		} else if guar == Guar_NO {
			pkg, err := parseMsg(bufRecv, n, conn.connID, logMsg)
			if err != nil {
				if pkg.LogT != nil {
					pkg.LogT.Error(err)
				} else {
					logT.Error(err)
				}
				continue
			}
			pkgs = append(pkgs, pkg)
		} else {
			var h Header
			h.Deserialize(bufRecv)
			logMsg.Errorf("error during read, Header.Guar == %x, Header:%v", h.Guar, h)
		}
		if len(pkgs) > 0 {
			n, closed := recvCh.SendN(pkgs)
			if closed || n != len(pkgs) {
				if closed {
					logMsg.Infof("error during read, recvCh is closed, exit") //chan 已关闭
					time.Sleep(time.Second * 3)
					return
				}
				//chan写满，可能需要做一些其它处理，比如：通知server已经忙不过来，请关闭稍后再试
				logMsg.Errorf("error during read, recvCh is full, pkgs:%v", pkgs[n:])
				atomic.AddInt64(&lostN, 1)
			}
		}
	}
	wg.Done()
}

func send(sendCh channel.IChanN, listener *net.UDPConn, wg *sync.WaitGroup, logT *log.TraceInfoST) {
	bufSend := make([]byte, udpBufLen)
	msgLen, n := 0, 0
	var buf []byte
	var ifaces []interface{}
	i, closed := 0, false
	for {
		if i >= len(ifaces) {
			i = 0
			ifaces, closed = sendCh.Recv(16, 1) //一次获取多个，可以减少锁竞争
			if closed {
				break //chan 已关闭,则退出
			}
		}
		atomic.AddInt64(&sendN, 1)
		iface := ifaces[i]
		i++

		pkg, ok := iface.(PkgRtn)
		if !ok {
			log.Errorf("msg type Error, type:[%v],value:[%v]", reflect.TypeOf(iface), iface)
			continue
		}
		bytes, err := proto.Marshal(pkg.Msg) //序列化msg
		if err != nil {
			logT.Errorf("proto.Marshal Error:%v,pkgIO:[%v]", err, pkg)
			continue
		}
		msgLen = len(bytes) + MSG_HEADER_LEN
		if len(bufSend) < msgLen {
			if msgLen > udpBufLen /*0xffff*/ { //消息体超大
				pkg.LogT.Errorf("send error, msg too large ,pkg:[%v]", pkg)
				continue
			}
			bufSend = make([]byte, msgLen)
		}
		guar := Guar_NO
		if pkg.Guar {
			guar = Guar_YES
		}
		h := Header{
			Len:  uint16(msgLen - MSG_HEADER_LEN), // 消息体的长度
			Type: pkg.Msg.MsgNO(),                 // 消息类型
			ID:   1,                               // 消息ID
			Ver:  1,                               // 版本号
			Guar: guar,                            // 保证到达
		}
		buf, err = h.Serialize(bufSend)
		copy(buf, bytes)

		conn := connID2Conn(pkg.ConnID)
		if conn == nil {
			pkg.LogT.Errorf("send message error,cannot get conn, connID:%d, pkg:%v", pkg.ConnID, pkg)
			continue
		}
		//以下处理KCP消息
		if pkg.Guar {
			conn.Lock()
			if conn.kcp != nil {
				if m := conn.kcp.Send(bufSend[:msgLen]); m < 0 {
					pkg.LogT.Errorf("kcp.Send error, pkg:%v", pkg)
				}
			} else {
				pkg.LogT.Errorf("conn.kcp == nil, pkg:%v", pkg)
			}
			conn.Unlock()
		} else {
			n, err = listener.WriteToUDP(bufSend[:msgLen], conn.addr)
			if err != nil || n != msgLen {
				pkg.LogT.Criticalf("error during read:%v,n:%d\n", err, n)
				break
			}
		}
	}
	log.Critical("udp.send exit")
	wg.Done()
}
func statistics() {
	var lastRecvN int64
	var lastSendN int64
	var lastLostN int64
	var lastTime int64
	for {
		lost := atomic.LoadInt64(&lostN)
		recv := atomic.LoadInt64(&recvN)
		send := atomic.LoadInt64(&sendN)
		if recv-lastRecvN > 0 {
			now := time.Now().UnixNano()
			fpsRecv := float64(recv-lastRecvN) / (float64(now-lastTime) / float64(time.Second))
			fpsSend := float64(send-lastSendN) / (float64(now-lastTime) / float64(time.Second))
			if lastTime == 0 {
				lastTime = now
				lastLostN = lost
				lastRecvN = recv
				lastSendN = send
				<-time.After(time.Millisecond * 1000 * 3)
				continue
			}
			log.Infof("io lost:%f, app lost:%.3f, recv fps:%.3fk/s, send fps:%.3fk/s, recv:%dk",
				float64(lost-lastLostN)/float64(recv-lastRecvN),
				float64(recv-send)/float64(recv),
				fpsRecv/1000.0, fpsSend/1000.0, recv/1000)

			lastTime = now
			lastLostN = lost
			lastRecvN = recv
			lastSendN = send
		}
		<-time.After(time.Millisecond * 1000 * 3)
	}
}
