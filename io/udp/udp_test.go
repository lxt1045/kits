package udp_test

import (
	"context"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lxt1045/kits/io/udp"
	"github.com/lxt1045/kits/log"
	"github.com/lxt1045/sence/conf"
	"github.com/lxt1045/sence/msg"
)

func TestStart(t *testing.T) {
	conf.Conf.CPUNUM = runtime.GOMAXPROCS(0)
	conf.Conf.IOChanSize = 64
	conf.Conf.Listen = ":18080"

	ctx, cancel := context.WithCancel(context.Background())
	logTracer := log.NewLogTrace(0, 0, 0) //
	recvCh, sendCh, err := udp.Start(ctx, logTracer, &conf.Conf)
	if err != nil {
		t.Error(err)
	}

	for {
		ifaces, closed := sendCh.Recv(16, 1) //一次获取多个，可以减少锁竞争
		if closed {
			break //chan 已关闭,则退出
		}

		if len(ifaces) > 0 {
			n, closed := recvCh.SendN(ifaces)
			if closed || n != len(ifaces) {
				if closed {
					t.Logf("error during read, recvCh is closed, exit") //chan 已关闭
					return
				}
				//chan写满，可能需要做一些其它处理，比如：通知server已经忙不过来，请关闭稍后再试
				t.Logf("error during read, recvCh is full, pkgs:%v", ifaces[n:])
			}
		}
	}
	//cctx, ccancel := context.WithCancel(context.Background())
	//cgw := client(cctx, 1)
	time.Sleep(time.Second * 30)
	//ccancel()
	//cgw.Wait()
	cancel()
}

//client gN 操作并行协程数量
func client(ctx context.Context, gN int) (wg *sync.WaitGroup) {
	wg = new(sync.WaitGroup)
	for i := 0; i < gN; i++ {
		conn0, err := net.Dial("udp", "127.0.0.1:18080")
		if err != nil {
			log.Infof("out, get error:%v", err)
			break
		}
		go func(gIdex int, conn net.Conn) {
			exit := make(chan int, 1)
			msgO := msg.LoginReq{
				UserId:   0,
				AppId:    []byte("u_007_8"),
				DeviceId: []byte("XXX*****"),
				UseLast:  false,
				Token:    []byte("DGhhkjhjkhjh7889354354"),
				Sign:     []byte("hjalhljdahdasdasj;d"),
				Name:     []byte("vision"),
				Stamp:    time.Now().UnixNano(),
				Imid:     []byte("test"),
				Version:  []byte("2.002"),
			}
			msgEq := func(m1, m2 *msg.LoginReq) bool {
				if m1.UserId == m1.UserId &&
					m1.UseLast == m1.UseLast &&
					//m1.Stamp == m1.Stamp &&
					string(m1.AppId) == string(m1.AppId) &&
					string(m1.DeviceId) == string(m1.DeviceId) &&
					string(m1.Token) == string(m1.Token) &&
					string(m1.Sign) == string(m1.Sign) &&
					string(m1.Name) == string(m1.Name) &&
					string(m1.Imid) == string(m1.Imid) &&
					string(m1.Version) == string(m1.Version) {
					return true
				}
				return false
			}
			_ = msgEq
			//recv
			go func() {
				wg.Add(1)
				bufRecv, n := make([]byte, 1024), 0
				var err error
				for {
					n, err = conn.Read(bufRecv)
					if err != nil || n == 0 {
						log.Errorf("read error:%v, n:%d,gIdex:%d", err, n, gIdex)
						wg.Done()
						break
					}
					var h udp.Header
					_, err = h.Deserialize(bufRecv)
					if err != nil {
						log.Errorf("read error:%v", err)
						continue
					}
					//msg, err = newMsg(h.Type)
					if uint16(msg.MSG_TYPE_LOGIN_REQ) != h.Type {
						log.Errorf("read error:%v, h:%v", err, h)
					}
					msgO1 := new(msg.LoginReq)
					if err != nil {
						log.Errorf("read error:%v", err)
						continue
					}
					if int(h.Len) > n-udp.MSG_HEADER_LEN {
						log.Errorf("error during read,h.Len > n-MSG_HEADER_LEN, h:%v, n:%d", h, n)
						continue
					}
					err = proto.Unmarshal(bufRecv[udp.MSG_HEADER_LEN:n], msgO1)
					log.Debug(msgO1)
					select {
					case <-exit:
						wg.Done()
						return
					default:
					}
				}
			}()

			bufSend, buf := make([]byte, 1024), make([]byte, 0)
			msgLen, n := 0, 0
			wg.Add(1)
			for {
				msgO.Stamp = time.Now().UnixNano()
				bytes, err := proto.Marshal(&msgO)
				if err != nil {
					log.Errorf("proto.Marshal Error:%v,pkgIO:[%v]", err, msgO)
					continue
				}
				msgLen = len(bytes) + udp.MSG_HEADER_LEN
				if len(bufSend) < msgLen {
					if msgLen > 0xffff { //消息体超大
						log.Errorf("send Error:%v,pkgIO:[%v]", "body is too big", msgO)
						continue
					}
					bufSend = make([]byte, msgLen)
				}
				h := udp.Header{
					Len:  uint16(msgLen - udp.MSG_HEADER_LEN), // 消息体的长度
					Type: uint16(msg.MSG_TYPE_LOGIN_REQ),      // 消息类型
					ID:   1,                                   // 消息ID
					Ver:  1,                                   // 版本号
					Guar: 0,                                   // 预留字段
				}
				buf, err = h.Serialize(bufSend)
				copy(buf, bytes)

				n, err = conn.Write(bufSend[:msgLen])
				if err != nil || n != msgLen {
					log.Errorf("error during read:%v,n:%d != msgLen:%d\n", err, n, msgLen)
					continue
				}
				time.Sleep(time.Millisecond * 1000)
				select {
				case <-ctx.Done():
					conn.Close()
					exit <- 1
					wg.Done()
					return
				default:
				}
			}
		}(i, conn0)
	}
	return
}
