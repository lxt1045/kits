package main

import (
	//"log"
	"net"
	"os"
	//"reflect"
	//"net/rpc"

	_ "github.com/lxt1045/geocache/msg"
	"github.com/lxt1045/kits/io/channel"
	"github.com/lxt1045/kits/io/gob"
	"github.com/lxt1045/kits/log"
)

func chkError(err error) {
	if err != nil {
		//log.Fatal(err)
		log.Error(err)
		os.Exit(1)
	}
}

func main() {
	tcpaddr, err := net.ResolveTCPAddr("tcp4", ":8080")
	chkError(err)
	//监听端口
	tcplisten, err2 := net.ListenTCP("tcp", tcpaddr)
	chkError(err2)
	//死循环处理连接请求
	for {
		conn, err3 := tcplisten.Accept()
		if err3 != nil {
			continue
		}
		//使用goroutine单独处理rpc连接请求
		recvCh, sendCh, err := gob.NewPeer().ServeConn(conn)
		if err != nil {
			log.Error("ServeConn(conn) get err:%v", err)
			continue
		}
		go do(recvCh, sendCh)
	}
}

func do(recvCh, sendCh channel.IChanN) {
	for {
		iface, closed := recvCh.Recv(32, 1)
		if closed {
			log.Infof("error during read, recvCh is closed, exit")
			break //chan 已关闭,则退出
		}
		if len(iface) == 0 {
			log.Infof("error during read, sendCh get len==0")
		}
		//		{
		//			pkg, ok := iface[0].(gob.Pkg)
		//			if !ok {
		//				log.Errorf("msg type Error, type:[%v],value:[%v]", reflect.TypeOf(iface[0]), iface[0])
		//				continue
		//			}
		//			y := pkg.Msg
		//			//log.Debugf("get msg:%s", y)
		//			//y.Receive()
		//			_ = y
		//		}
		n, closed := sendCh.SendN(iface)
		if closed || n != len(iface) {
			if closed {
				log.Infof("error during read, sendCh is closed, exit") //chan 已关闭
				break
			}
			//chan写满，可能需要做一些其它处理，比如：通知server已经忙不过来，请关闭稍后再试
			//log.Errorf("error during read, sendCh is full, pkg:%v", iface)
		}
	}
	recvCh.Close()
	sendCh.Close()
	log.Errorf("do() out!")
}
