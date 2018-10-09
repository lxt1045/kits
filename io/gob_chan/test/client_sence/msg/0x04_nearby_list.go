package msg

import (
	"time"

	"github.com/lxt1045/geocache/cache"
	"github.com/lxt1045/geocache/services"
	"github.com/lxt1045/kits/log"
)

//RecvMsg ServiceInfo消息是服务器连接后第一个要发送的消息，接收到这个消息后，才能开始服务
func (m *NearbyList) RecvMsg(connID int, logT *log.TraceInfoST) {
	logT.Infof("MsgNotify.WsReceive, msg:", m)
	return
}

func init() {

	//所有列表更新工作放到一个协程中处理，利于采用无锁过程; 采用range还是channel通知在处理？
	renewNearbyLists()
}

func renewNearbyLists() {
	go cache.RenewNearbyLists(func(connID int, userID uint64, nearbyList []uint64) bool {
		msg := &NearbyList{
			UserID:      userID,
			NearbyUsers: nearbyList,
			Timestamp:   time.Now().UnixNano(),
		}
		services.Send([]int{connID}, msg)
		return true
	})
}
