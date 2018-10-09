package gob

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/lxt1045/kits/io/channel"
	"github.com/lxt1045/kits/log"
)

var (
	arrayMsg     []reflect.Type
	registerLock sync.Mutex //arrayMsg的操作锁
)

func init() {
	arrayMsg = make([]reflect.Type, 64)
}

//Register msg要先注册后，才能正确接收该类型的数据
func Register(msg IMsg) (err error) {
	registerLock.Lock()
	defer registerLock.Unlock()

	k := typeHash(uint16(msg.MsgNO()))
	if k >= len(arrayMsg) {
		arrayNew := make([]reflect.Type, k+1)
		copy(arrayNew, arrayMsg)
		arrayMsg = arrayNew
	}
	if arrayMsg[k] != nil {
		err = fmt.Errorf("the key has been taken, key:%d, meg:%v", k, msg)
		return
	}

	v := reflect.ValueOf(msg)
	t := reflect.Indirect(v).Type()
	arrayMsg[k] = t
	return
}
func typeHash(_t uint16) int {
	return int(_t) //return int(_t)/500 + int(_t)%500
}
func newMsg(_t uint32) (IMsg, error) {
	key := typeHash(uint16(_t))
	if len(arrayMsg) <= int(key) || key < 0 {
		return nil, fmt.Errorf("key error:%d", key)
	}
	t := arrayMsg[key]
	if t == nil {
		return nil, fmt.Errorf("type error:%v", _t)
	}
	vo := reflect.New(t)
	msg, ok := vo.Interface().(IMsg)
	if !ok {
		return nil, fmt.Errorf("Create Object Error,type:[%d]%v, val:%v", _t, t, msg)
	}
	return msg, nil
}

// Server 一条连接一个Server，连接的相关的信息可以存这里
type Server struct {
	serviceMap sync.Map // map[string]*service

	userIDs []uint64 //该连接对应的需求的所有userID，也就是该client订阅的所有user
	logT    *log.TraceInfoST

	//reqFreeList *Request //Request的空闲对象列表，暂时不用，不好管理，什么时候回收呢？
	//r sync.Mutex //锁，用于多线程发，多线程收
	//w sync.Mutex
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{
		logT: log.NewLogTrace(0, 0, 0),
	}
}

func (server *Server) ServeConn(conn io.ReadWriteCloser) (channel.IChanN, channel.IChanN, error) {
	encBuf := bufio.NewWriter(conn)
	codec := &gobServerCodec{ //每条连接一个codec，为了避免争用加锁
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
		enc:    gob.NewEncoder(encBuf),
		encBuf: encBuf,
	}
	sendCh, err := server.send(codec)
	if err != nil {
		return nil, nil, err
	}
	recvCh, err := server.recv(codec)
	if err != nil {
		sendCh.Close()
		return nil, nil, err
	}
	return recvCh, sendCh, nil
}

var invalidRequest = struct{}{}

func (server *Server) send(codec ServerCodec) (sendCh channel.IChanN, err error) {
	sendCh = channel.NewChanN(10000)
	go func(codec ServerCodec, sendCh channel.IChanN) {
		var ifaces []interface{}
		i, closed := 0, false
		resp := &Response{}
		for {
			if i >= len(ifaces) {
				i = 0
				ifaces, closed = sendCh.Recv(16, 1) //一次获取多个，可以减少锁竞争
				if closed {
					break //chan 已关闭,则退出
				}
			}
			iface := ifaces[i]
			i++

			pkg, ok := iface.(Pkg)
			if !ok {
				log.Errorf("msg type Error, type:[%v],value:[%v]", reflect.TypeOf(iface), iface)
				continue
			}
			// Encode the response header
			resp.MethodNO = uint32(pkg.Msg.MsgNO())
			resp.Seq = pkg.Seq
			//server.w.Lock()
			err := codec.WriteResponse(resp, pkg.Msg)
			if err != nil {
				log.Println("rpc: writing response:", err)
				break
			}
			//server.w.Unlock()

		}
		codec.Close()
		sendCh.Close()
		log.Infof("exit")
	}(codec, sendCh)
	return
}
func (server *Server) recv(codec ServerCodec) (recvCh channel.IChanN, err error) {
	recvCh = channel.NewChanN(10000) //100万,1M
	go func(codec ServerCodec, recvCh channel.IChanN) {
		//var err error
		//var req Request
		//server.r.Lock()
		for {
			req := Request{}
			err := codec.ReadRequestHeader(&req)
			if err != nil {
				server.logT.Error(err)
				break
			}
			imsg, err := newMsg(req.MethodNO)
			if err != nil {
				server.logT.Error(err)
				err = codec.ReadRequestBody(nil) //即使出错，也要把body给取出来
				if err != nil {
					server.logT.Errorf("reading error body: %v", err)
					break
				}
				continue
			}
			err = codec.ReadRequestBody(imsg) //即使出错，也要把body给取出来
			//server.r.Unlock()
			if err != nil {
				server.logT.Errorf("reading error body: %v", err)
			}
			full, closed := recvCh.Send(Pkg{Msg: imsg, Seq: req.Seq})
			if full || closed {
				if closed {
					server.logT.Error("recv error, recvCh is closed")
					//server.r.Lock()
					break
				}
				server.logT.Errorf("recv error, recvCh is full, msg:%v", Pkg{Msg: imsg, Seq: req.Seq})
				continue
			}
			//server.r.Lock()
		}
		//server.r.Unlock()
		codec.Close()
		recvCh.Close()
		log.Infof("exit")
	}(codec, recvCh)
	return
}

type gobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool //为了避免调用Close()时多次释放
}

func (c *gobServerCodec) ReadRequestHeader(r *Request) error {
	return c.dec.Decode(r)
}

func (c *gobServerCodec) ReadRequestBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *gobServerCodec) WriteResponse(r *Response, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Gob couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a gob problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *gobServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}
