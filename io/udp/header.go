package udp

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sync"
	//"github.com/golang/protobuf/proto"
)

//
// 使用sync.Poll来分配最活跃的msg，比如move消息
//
//

var (
	arrayMsg     []reflect.Type
	registerLock sync.Mutex //arrayMsg的操作锁
)

func init() {
	arrayMsg = make([]reflect.Type, 64)
}
func typeHash(_t uint16) int {
	//return int(_t)/500 + int(_t)%500
	return int(_t)
}

func newMsg(_t uint16) (IMsg, error) {
	key := typeHash(_t)
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
		err = fmt.Errorf("the key has been taken, key:%d, msg.type:%v", k, reflect.TypeOf(msg))
		return
	}

	v := reflect.ValueOf(msg)
	t := reflect.Indirect(v).Type()
	arrayMsg[k] = t
	return
}

// ************ 消息头的相关处理 ************
/*
    0    1    2    3    4    5    6    7
 -----------------------------------------
 |Guar|Ver | MsgLen  | MsgType |  MsgId  |
 -----------------------------------------
 |      Body...                          |
 -----------------------------------------
*/
const MSG_HEADER_LEN = 8 //unsafe.Sizeof(p)

const Guar_YES = uint8(0xAA) //10101010
const Guar_NO = uint8(0x55)  //01010101

type Header struct {
	Guar uint8  // 是否保证到达,即是否通过KCP发送,guarantee缩写;01010101不可靠传输,10101010可靠传输
	Ver  uint8  // 版本号
	Len  uint16 // 消息体的长度
	Type uint16 // 消息类型
	ID   uint16 // 消息ID
}

// Serialize 在传入的切片上序列化本对象，返回未使用的切片
func (p *Header) Serialize(data []byte) ([]byte, error) {
	if MSG_HEADER_LEN > len(data) {
		return data, fmt.Errorf("input data(%v) err", data)
	}
	data = encode8u(data, p.Guar)
	data = encode8u(data, p.Ver)
	data = encode16u(data, p.Len)
	data = encode16u(data, p.Type)
	data = encode16u(data, p.ID)

	stats.PackageInc() //统计信息
	return data, nil
}
func (p *Header) Deserialize(data []byte) ([]byte, error) {
	if MSG_HEADER_LEN > len(data) {
		return data, fmt.Errorf("input data(%v) err", data)
	}
	data = decode8u(data, &p.Guar)
	data = decode8u(data, &p.Ver)
	data = decode16u(data, &p.Len)
	data = decode16u(data, &p.Type)
	data = decode16u(data, &p.ID)
	return data, nil
}

// ************ 消息头end ************
/* encode 8 bits unsigned int */
func encode8u(p []byte, c byte) []byte {
	p[0] = c
	return p[1:]
}

/* decode 8 bits unsigned int */
func decode8u(p []byte, c *byte) []byte {
	*c = p[0]
	return p[1:]
}

/* encode 16 bits unsigned int (lsb) */
func encode16u(p []byte, w uint16) []byte {
	binary.LittleEndian.PutUint16(p, w)
	return p[2:]
}

/* decode 16 bits unsigned int (lsb) */
func decode16u(p []byte, w *uint16) []byte {
	*w = binary.LittleEndian.Uint16(p)
	return p[2:]
}

/* encode 32 bits unsigned int (lsb) */
func encode32u(p []byte, l uint32) []byte {
	binary.LittleEndian.PutUint32(p, l)
	return p[4:]
}

/* decode 32 bits unsigned int (lsb) */
func decode32u(p []byte, l *uint32) []byte {
	*l = binary.LittleEndian.Uint32(p)
	return p[4:]
}
