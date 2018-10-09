package gob

import (
	"github.com/lxt1045/kits/log"
)

// IMsg 是路由层(消息分发层)和逻辑层(消息处理层)之间传递的数据结构,Msg:message的缩写
type IMsg interface {
	RecvMsg(logT *log.TraceInfoST) (msgs PkgRtn, err error)
	//MsgNO() uint16 //返回对应的消息编号，每种消息一个编号，保证唯一性
	IMsgRtn
}

// IMsgRtn 是路由层(消息分发层)和逻辑层(消息处理层)之间传递的数据结构,Msg:message的缩写
type IMsgRtn interface {
	MsgNO() uint16 //消息编号
}

//Pkg 每次收到消息，通过这个结构包装后写到channel中
type Pkg struct {
	Msg IMsg

	//gob的每个请求都有一个自动增加的针对连接唯一的编号，相当于是一个session，
	//当client收到数据时，client会根据编号调用回调函数，
	//所以，server端收到消息后，必须保存消息的编号，以便在回复消息的时候有目的地！
	Seq uint64
}

//MsgRtn 是RecvMsg()方法的返回值，Msg:message的缩写,Rtn:return的缩写
type PkgRtn struct {
	Msg IMsgRtn
	Seq uint64 //gob的每个请求都有一个自动增加的针对连接唯一的编号，相当于是一个session

	LogT *log.TraceInfoST
}
