package msg

import (
	"github.com/lxt1045/geocache/cache"
	"github.com/lxt1045/geocache/services"
	"github.com/lxt1045/kits/log"
)

//RecvMsg 如果收到Move消息，则会调用此函数
func (m *Move) RecvMsg(connID int, logT *log.TraceInfoST) {
	logT.Infof("Move.RecvMsg, msg:", m)

	//0、先取得Info数据结构
	info := cache.Key2Info(m.UserID)
	var geohash uint64
	info.Lock()
	geohash = info.Geohash

	info.ConnID = connID
	info.Edge = m.Edge
	info.Geohash = m.Geohash
	info.X = m.X
	info.Y = m.Y
	info.Z = m.Z
	info.Dir = m.Dir
	info.Timestamp = m.Timestamp
	if len(m.Dress) > 0 {
		info.Dress = m.Dress
	}

	//更新 Geo索引 数据表
	cache.Renew(geohash, info)
	info.Unlock()

	//给订阅自己的User广播自己的位置信息
	err := services.Send(*info.NearbySendConnIDs, m)
	if err != nil {
		logT.Error(err)
	}
	return
}
