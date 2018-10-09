package gob

import (
	"bufio"
	"encoding/gob"
	"io"
	"reflect"
	//"log"
	"net"
	"sync"
	"sync/atomic"

	//"github.com/lxt1045/kits/io/channel"
	"github.com/lxt1045/kits/log"
)

type Client struct {
	codec ClientCodec

	sync.Mutex // protects following

	pending sync.Map //得到返回结果后，会找到发送者的call，调用其Done()函数
	seq     uint64   //请求编号，唯一，递增
	status  uint32   //0:正常；closing:1, shutdown:2
}

func NewClient(conn io.ReadWriteCloser) *Client {
	encBuf := bufio.NewWriter(conn)
	client := &Client{
		codec: &gobClientCodec{
			rwc:    conn,                   //   io.ReadWriteCloser
			dec:    gob.NewDecoder(conn),   // *gob.Decoder
			enc:    gob.NewEncoder(encBuf), //  *gob.Encoder
			encBuf: encBuf,                 // *bufio.Writer
		},
	}
	go client.recv() //必须保证单协程操作，否则会因为header和body乱序而出错；而send()是线程安全的
	return client
}

//SendReply 外部发rpc的时候调用
func (client *Client) GetReply(args IMsg, // Args作为请求的body
	reply IMsg, // 需要的返回值类型 (*struct)
	_type uint32, //调用类型，只发不等待请求数据，none:0；一次请求循环多次接收，loop:1； 一次请求一次接收，once:2；取消调用：cancel:3
	timeOut int64,
	doneFunc func(reply interface{}, err error), //在收到返回值或未能正确调用的时候调用 )
) {
	client.SendCall(&Call{
		Args:      args,
		ReplyType: reflect.Indirect(reflect.ValueOf(reply)).Type(),
		MethodNO:  uint32(args.MsgNO()),
		Done:      doneFunc,
		Type:      _type, //只发送，不接收none:0；一次请求循环多次接收，loop:1； 一次请求一次接收，once:2；取消调用：cancel:3
		TimeOut:   timeOut,
	})
}

//SendCall 外部发rpc的时候调用
func (client *Client) SendCall(call *Call) {
	if atomic.LoadUint32(&client.status) != 0 {
		call.Done(nil, ErrShutdown)
		return
	}
	if call.Type == 0 {
		call.Done(nil, ErrNotSuportType)
		return
	}

	req := Request{ //应该放到sync.pool中
		Seq:      atomic.AddUint64(&client.seq, 1),
		MethodNO: call.MethodNO,
	}
	I, loaded := client.pending.LoadOrStore(req.Seq, call) //为了强一致性，用LoadOrStore()
	if loaded {
		//同一个ID，表示走完了一个循环，可能性极低！！！
		if p, ok := I.(*Call); ok {
			p.Done(nil, ErrPendingTooLong)      //旧的返回
			client.pending.Store(req.Seq, call) //新的接上
		}
	}
	//codec 非线程安全，需要加锁！！
	client.Lock()
	err := client.codec.WriteRequest(&req, call.Args) //request作为hedaer, Args作为body
	client.Unlock()
	if err != nil {
		call.Done(nil, ErrSendError)
		client.pending.Delete(req.Seq)
	}
}

//Send 外部发rpc的时候调用
func (client *Client) Send(args IMsg) (err error) {
	if atomic.LoadUint32(&client.status) != 0 {
		err = ErrShutdown
		return
	}

	req := Request{ //应该放到sync.pool中
		Seq:      atomic.AddUint64(&client.seq, 1),
		MethodNO: uint32(args.MsgNO()),
	}

	//codec 非线程安全，需要加锁！！
	client.Lock()
	err = client.codec.WriteRequest(&req, args) //request作为hedaer, Args作为body
	client.Unlock()
	if err != nil {
		return
	}
	return
}

func (client *Client) recv() {
	var err error
	var response Response
	for err == nil {
		response = Response{}
		err = client.codec.ReadResponseHeader(&response)
		if err != nil {
			continue
		}
		I, ok := client.pending.Load(response.Seq) //为了强一致性，用LoadOrStore()
		if !ok {
			log.Errorf("response has no callback function, resp:%v", response)
			continue
		}
		call, ok := I.(*Call)
		if !ok {
			log.Errorf("response has an error callback function, resp:%v, call:%v", response, I)
			client.pending.Delete(response.Seq)
			continue
		}
		replay := reflect.New(call.ReplyType).Interface()
		//log.Infof("reflect.New:%v, type:%v, call:%v", replay, call.ReplyType, call)

		if response.Error != "" {
			err = client.codec.ReadResponseBody(nil) //header找不到回调，也要把body给取出来
			if err != nil {
				log.Errorf("reading error body: %v", err)
			}
			goto doneErr
		}
		err = client.codec.ReadResponseBody(replay)
		if err != nil {
			log.Errorf("reading error body: %v", err)
			goto doneErr
		}
		call.Done(replay, nil)
		if call.Type == CallTypeOnce {
			client.pending.Delete(response.Seq)
		}

		continue
	doneErr:
		call.Done(nil, ServerError(response.Error))
		if call.Type == CallTypeOnce {
			client.pending.Delete(response.Seq)
		}
	}
	if err == io.EOF {
		if atomic.LoadUint32(&client.status) != 0 {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	client.pending.Range(func(k, v interface{}) bool {
		if call, ok := v.(*Call); ok {
			call.Done(nil, err)
		}
		return true
	})

	log.Errorf("rpc: client protocol error:", err)
}

func (client *Client) Close() error {
	atomic.StoreUint32(&client.status, 2)
	return client.codec.Close()
}

type gobClientCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func (c *gobClientCodec) WriteRequest(r *Request, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *gobClientCodec) ReadResponseHeader(r *Response) error {
	return c.dec.Decode(r)
}

func (c *gobClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *gobClientCodec) Close() error {
	return c.rwc.Close()
}

// Dial connects to an RPC server at the specified network address.
func Dial(network, address string) (*Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}
