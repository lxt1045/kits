package udp

import (
	//"github.com/golang/protobuf/proto"
	//"services/D_Go_V2.0_sence/msg"
	"github.com/lxt1045/kits/log"
	"github.com/lxt1045/sence/session"
)

/*
  IMsg.Receive(t uint16, addr *net.UDPAddr, log *log.TraceInfoST) (IMsg, error)
	逻辑层(消息处理层)
		/\
		|| type IMsg interface { Receive(), Reset(), String(), ProtoMessage() }
		\/
	路由层(消息分发层)
		/\
		|| type Pkg struct{ MsgType uint16, Addr *net.UDPAddr, Msg IMsg }
		\/
	IO层(消息收发层)
*/

// IMsg 是路由层(消息分发层)和逻辑层(消息处理层)之间传递的数据结构,Msg:message的缩写
type IMsg interface {
	RecvMsg(pSess *session.Session, log *log.TraceInfoST) (msgs []PkgRtn, err error)
	IMsgRtn
	// MsgNO() uint16 //消息编号

	// //以下是 proto.Message 接口的所有方法，所以IWsRequest是可以转化为 proto.Message 接口的
	// Reset()
	// String() string
	// ProtoMessage()
}

// IMsgRtn 是路由层(消息分发层)和逻辑层(消息处理层)之间传递的数据结构,Msg:message的缩写
type IMsgRtn interface {
	MsgNO() uint16 //消息编号

	//以下是 proto.Message 接口的所有方法，所以IWsRequest是可以转化为 proto.Message 接口的
	Reset()
	String() string
	ProtoMessage()
}

// Pkg 是用于在IO层(消息收发层)和路由层(消息分发层)之间交互的消息
type Pkg struct {
	//MsgType uint16
	Guar   bool   //是否保证到达，false表示不保证到达
	ConnID uint64 //UserID uint64 // 用于获取UDPAddr;
	Msg    IMsg

	LogT *log.TraceInfoST
}

//MsgRtn 是RecvMsg()方法的返回值，Msg:message的缩写,Rtn:return的缩写
type PkgRtn struct {
	//MsgType uint16
	Guar   bool   //是否保证到达，false表示不保证到达
	ConnID uint64 //用于获取UDPAddr; 有可能要给其他用户发消息,比如单用户多设备登陆踢人的时候，还要跨服踢人！放cache服务器里设个标志，“惰性踢人”！
	Msg    IMsgRtn

	LogT *log.TraceInfoST
}
