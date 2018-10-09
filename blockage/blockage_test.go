/*
这里用goroutine协程池，处理会阻塞的操作，以避免主协程被阻塞影响业务性能
*/

package blockage

import (
	"testing"
	"time"

	"github.com/lxt1045/kits/conf"
	"github.com/lxt1045/kits/log"
	"github.com/lxt1045/sence"
)

func init() {
}

//
func TestCallback(t *testing.T) {
	logT := log.NewLogTrace(1, 2, 3)

	Callback([]interface{}{uint64(20201)}, func(ifaces []interface{}) {
		userID, ok := ifaces[0].(uint64)
		if !ok || userID == 0 {
			logT.Errorf("got an exception param")
			return
		}
		userInfo, err := sence.GetUserInfoByID(userID)
		t.Logf("userInfo:%v, err:%v", userInfo, err)
	})

	time.Sleep(time.Second * 3)
}
func BenchmarkRandSet(b *testing.B) {

	b.StopTimer()
	logT := log.NewLogTrace(1, 2, 3)

	Callback([]interface{}{uint64(20201)}, func(ifaces []interface{}) {
		userID, ok := ifaces[0].(uint64)
		if !ok || userID == 0 {
			logT.Errorf("got an exception param")
			return
		}
		userInfo, err := sence.GetUserInfoByID(userID)
		if err != nil {
			b.Logf("userInfo:%v, err:%v", userInfo, err)
		}
		b.Logf("userInfo:%v, err:%v", userInfo, err)
	})

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		//if i < 100 {
		Callback([]interface{}{uint64(20201)}, func(ifaces []interface{}) {
			userID, ok := ifaces[0].(uint64)
			if !ok || userID == 0 {
				logT.Errorf("got an exception param")
				return
			}
			userInfo, err := sence.GetUserInfoByID(userID)
			if err != nil || userInfo == nil {
				//b.Logf("userInfo:%v, err:%v", userInfo, err)
			}
		})
		//}
	}
	b.StopTimer()

	time.Sleep(time.Second * 3)
	b.StartTimer()
}
