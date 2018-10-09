package gob

import (
	"errors"
	"reflect"
)

type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var ErrShutdown = errors.New("connection is shut down")
var ErrNotSuportType = errors.New("type is not suported")
var ErrPendingTooLong = errors.New("call is pending too long")
var ErrSendError = errors.New("got error when send to network")

const (
	CallTypeNone   = 0
	CallTypeLoop   = 1
	CallTypeOnce   = 2
	CallTypeCancel = 3
)

type Call struct {
	Args      interface{}  // Args作为请求的body
	ReplyType reflect.Type // 需要的返回值类型 (*struct)

	MethodNO uint32                             //请求类型
	Done     func(reply interface{}, err error) //在收到返回值或未能正确调用的时候调用
	Type     uint32                             //调用类型，只发不等待请求数据，none:0；一次请求循环多次接收，loop:1； 一次请求一次接收，once:2；取消调用：cancel:3
	TimeOut  int64                              //超时时间，超市部分赶回，则代表调用失败
}

type Request struct {
	MethodNO uint32   //
	Seq      uint64   // sequence number chosen by client
	next     *Request // for free list in Server
}

type Response struct {
	MethodNO uint32    // echoes that of the Request
	Seq      uint64    // echoes that of the request
	Error    string    // error, if any.
	next     *Response // for free list in Server
}

type ClientCodec interface {
	// WriteRequest must be safe for concurrent use by multiple goroutines.
	WriteRequest(*Request, interface{}) error
	ReadResponseHeader(*Response) error
	ReadResponseBody(interface{}) error

	Close() error
}

//
//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//server

type ServerCodec interface {
	ReadRequestHeader(*Request) error
	ReadRequestBody(interface{}) error
	// WriteResponse must be safe for concurrent use by multiple goroutines.
	WriteResponse(*Response, interface{}) error

	Close() error
}
