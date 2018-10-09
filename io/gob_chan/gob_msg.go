package gob

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/lxt1045/kits/log"
)

// IMsg 是路由层(消息分发层)和逻辑层(消息处理层)之间传递的数据结构,Msg:message的缩写
type IMsg interface {
	RecvMsg(connID int, logT *log.TraceInfoST)
	MsgNO() uint16 //返回对应的消息编号，每种消息一个编号，保证唯一性
}

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
