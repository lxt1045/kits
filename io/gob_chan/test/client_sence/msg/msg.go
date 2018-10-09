package msg

import (
	"github.com/lxt1045/kits/io/gob"
	"github.com/lxt1045/kits/log"
)

func init() {
	imsgs := []gob.IMsg{
		&ServiceInfo{},
		&Ack{},
		&Move{},
		&NearbyList{},
	}
	for _, imsg := range imsgs {
		if err := gob.Register(imsg); err != nil {
			log.Criticalf("gob.Register(&Ack{}) got error:%v", err)
		}
	}
}

type ServiceInfo struct {
	ServiceID string
}

func (m *ServiceInfo) MsgNO() uint16 {
	return 1
}

type Ack struct {
	UserID    int64
	TimeStamp int64
}

func (m *Ack) MsgNO() uint16 {
	return 2
}

//Move 是移动消息，同时，还要给对方发送附近的人
type Move struct {
	Edge uint32 //Move 和 Edge 合成一个消息； 0：表示Move,1:表示Edge，即表示是边界外的位置，不需要给她发广播消息

	UserID    uint64
	Timestamp int64
	Geohash   uint64
	X, Y, Z   float32 //U3D 的原始坐标？
	Dir       float32 //direction 方向
	Dress     []byte
}

func (m *Move) MsgNO() uint16 {
	return 3
}

//NearbyList 是附近的人列表变化(靠近、离开)
type NearbyList struct {
	UserID      uint64
	NearbyUsers []uint64
	Timestamp   int64
}

func (m *NearbyList) MsgNO() uint16 {
	return 4
}
