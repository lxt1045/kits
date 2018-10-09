package io

import (
	"context"
	"net"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/lxt1045/kits/conf"
	"github.com/lxt1045/kits/io/channel"
	"github.com/lxt1045/kits/log"
	"github.com/lxt1045/sence/msg"
)

//
// 使用sync.Poll来分配最活跃的msg，比如move消息
//
//

// Start udp接口启动的时候执行，inCh, outCh是输入、输出chan，为了和后端处理模块解耦
func Start(ctx context.Context, log *log.TraceInfoST, conf *conf.ConfigST) (
	recvCh, sendCh channel.IChanN, err error) {

	ioChanSize := conf.IOChanSize
	if ioChanSize <= 0 {
		ioChanSize = 1024
	}
	recvCh = channel.NewChanN(ioChanSize)
	sendCh = channel.NewChanN(ioChanSize)
	//判断ctx是否被取消了，如果是就退出
	go func(ctx context.Context) {
		<-ctx.Done()
		recvCh.Close()
		//close sendCh后，写入会panic，可以用select方式处理，
		//但是据测试，每个select的case要增加100ns的消耗，所以为了性能让它clash吧!
		sendCh.Close()
	}(ctx)

	udpAddr, err := net.ResolveUDPAddr("udp", conf.Listen)
	if err != nil {
		log.Errorf("error:%v", err)
		return
	}
	listener, err := net.ListenUDP("udp", udpAddr)
	if err != nil || listener == nil {
		log.Errorf("error:%v", err)
		return
	}
	defer listener.Close()
	if err != nil || listener == nil {
		log.Errorf("error:%v", err)
		return
	}

	//多线收发数据，可以提高数据收发速度
	//以下是发数据
	for i := 0; i < conf.CPUNUM; i++ {
		go func() {
			bufSend := make([]byte, 1024)
			msgLen, n := 0, 0
			var buf []byte
			for {
				var pkg Pkg
				iface, closed := sendCh.Recv(1, 1)
				//chan 已关闭
				if closed {
					break
				}
				pkg, ok := iface[0].(Pkg)
				if !ok {
					log.Errorf("msg type Error, type:[%v],value:[%v]", reflect.TypeOf(iface[0]), iface[0])
					continue
				}

				bytes, err := proto.Marshal(pkg.Msg)
				if err != nil {
					log.Errorf("proto.Marshal Error:%v,pkgIO:[%v]", err, pkg)
					continue
				}
				msgLen = len(bytes) + MSG_HEADER_LEN
				if len(bufSend) < msgLen {
					if msgLen > 0xffff { //消息体超大
						continue
					}
					bufSend = make([]byte, msgLen)
				}
				h := Header{
					Len:  uint16(msgLen - MSG_HEADER_LEN), // 消息体的长度
					Type: pkg.MsgType,                     // 消息类型
					ID:   1,                               // 消息ID
					Ver:  1,                               // 版本号
					Resv: 0,                               // 预留字段
				}
				buf, err = h.Serialize(bufSend)
				copy(buf, bytes)

				n, err = listener.WriteToUDP(bufSend[:msgLen], pkg.Addr)
				if err != nil || n != msgLen {
					log.Criticalf("error during read:%v,n:%d\n", err, n)
					break
				}
			}
		}()
	}
	//以下是接收数据
	for i := 0; i < conf.CPUNUM; i++ {
		go func() {
			n := 0
			bufRecv := make([]byte, 1024)
			//var buf []byte
			for {
				var pkg Pkg
				n, pkg.Addr, err = listener.ReadFromUDP(bufRecv)
				if err != nil || n <= 0 {
					log.Criticalf("error during read:%v, n:%d", err, n)
					break
				}
				var h MsgHeaderST
				_, err = h.Deserialize(bufRecv)
				pkg.Msg, err = newMsg(h.Type)
				pkg.MsgType = h.Type

				err = proto.Unmarshal(bufRecv[MSG_HEADER_LEN:n], pkg.Msg)
				_, closed := recvCh.Send(pkg)
				//chan 已关闭
				if closed {
					break
				}
			}
		}()
	}
	return
}
