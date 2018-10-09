package gob

import (
	"bufio"
	"encoding/gob"
	"io"
	"net"

	"github.com/lxt1045/kits/log"
)

// Peer 一条连接一个Peer，连接的相关的信息可以存这里
type Peer struct {

	//userIDs []uint64 //该连接对应的需求的所有userID，也就是该client订阅的所有user
	logT *log.TraceInfoST

	//reqFreeList *Header //Hearder的空闲对象列表，暂时不用，不好管理，什么时候回收呢？
	//r sync.Mutex //锁，用于多线程发，多线程收
	//w sync.Mutex
}

// NewPeer returns a new Peer.
func NewPeer() *Peer {
	return &Peer{
		logT: log.NewLogTrace(0, 0, 0),
	}
}

func (p *Peer) ServeConn(conn io.ReadWriteCloser) (chan IMsg, chan IMsg, error) {
	encBuf := bufio.NewWriter(conn)
	codec := &gobCodec{ //每条连接一个codec，为了避免争用加锁
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
		enc:    gob.NewEncoder(encBuf),
		encBuf: encBuf,
	}
	sendCh, err := p.send(codec)
	if err != nil {
		return nil, nil, err
	}
	recvCh, err := p.recv(codec)
	if err != nil {
		close(sendCh)
		return nil, nil, err
	}
	return recvCh, sendCh, nil
}

func (p *Peer) send(codec Codec) (sendCh chan IMsg, err error) {
	sendCh = make(chan IMsg, 10000)
	go func(codec Codec, sendCh chan IMsg) {
		resp := &Header{}
		for {
			msg, ok := <-sendCh
			if !ok {
				break //chan 已关闭,则退出
			}

			// Encode the response header
			resp.MsgNO = uint32(msg.MsgNO())
			//p.w.Lock()
			err := codec.Write(resp, msg)
			if err != nil {
				log.Println("rpc: writing response:", err)
				break
			}
			//p.w.Unlock()

		}
		codec.Close()
		close(sendCh)
		log.Infof("exit")
	}(codec, sendCh)
	return
}
func (p *Peer) recv(codec Codec) (recvCh chan IMsg, err error) {
	recvCh = make(chan IMsg, 10000)
	go func(codec Codec, recvCh chan IMsg) {
		//var err error
		//var req Header
		//p.r.Lock()
		for {
			req := Header{}
			err := codec.ReadHeader(&req)
			if err != nil {
				p.logT.Error(err)
				break
			}
			imsg, err := newMsg(req.MsgNO)
			if err != nil {
				p.logT.Error(err)
				err = codec.ReadBody(nil) //即使出错，也要把body给取出来
				if err != nil {
					p.logT.Errorf("reading error body: %v", err)
					break
				}
				continue
			}
			err = codec.ReadBody(imsg) //即使出错，也要把body给取出来
			//p.r.Unlock()
			if err != nil {
				p.logT.Errorf("reading error body: %v", err)
			}
			//p.logT.Infof("reading msg: %v", imsg)
			recvCh <- imsg
			//p.r.Lock()
		}
		//p.r.Unlock()
		codec.Close()
		close(recvCh)
		log.Infof("exit")
	}(codec, recvCh)
	return
}

type gobCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool //为了避免调用Close()时多次释放
}

func (c *gobCodec) ReadHeader(r *Header) error {
	return c.dec.Decode(r)
}

func (c *gobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *gobCodec) Write(r *Header, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Peer couldn't encode the header. Should not happen, so if it does,
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

func (c *gobCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

func Dial(network, address string) (chan IMsg, chan IMsg, error) {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		return nil, nil, err
	}
	return NewPeer().ServeConn(conn)
}
