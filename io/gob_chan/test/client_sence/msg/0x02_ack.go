package msg

import (
	"github.com/lxt1045/kits/log"
)

func (m *Ack) RecvMsg(srcID int, logT *log.TraceInfoST) {
	logT.Infof("MsgNotify.WsReceive, msg:", m)
	return
}
