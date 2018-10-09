package msg

import (
	"github.com/lxt1045/geocache/services"
	"github.com/lxt1045/kits/log"
)

//RecvMsg ServiceInfo消息是服务器连接后第一个要发送的消息，接收到这个消息后，才能开始服务
func (m *ServiceInfo) RecvMsg(connID int, logT *log.TraceInfoST) {
	logT.Infof("ServiceInfo.RecvMsg, msg:", m)

	//
	err := services.SetServiceID(connID, m.ServiceID)
	if err != nil {
		logT.Errorf("get error:%v", err)
	}

	return
}
