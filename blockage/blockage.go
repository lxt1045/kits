/*
这里用goroutine协程池，处理会阻塞的操作，以避免主协程被阻塞影响业务性能
*/

package blockage

import (
	"fmt"
	"sync"
	"sync/atomic"

	"services/D_Go_V2.0_kit/io/channel"
	"services/D_Go_V2.0_kit/log"
)

var (
	queue       channel.IChanN //Queue 就是阻塞调度的输入队列
	routineN    int32          //用于统计当前的协程数量，因为sync.Pool会被GC回收，无法统计
	routineNMax int32
	pool        *sync.Pool
	logT        *log.TraceInfoST

	blockageGoroutineN int
)

func init() {
	blockageGoroutineN = 1024

	pool = &sync.Pool{
		New: newRoutine,
	}
}

//Routine 就是一个阻塞的调用过程
type Routine struct {
	Params []interface{}
	Func   func(params []interface{})
}

func newRoutine() interface{} {
	if atomic.AddInt32(&routineN, 1) > routineNMax {
		atomic.AddInt32(&routineN, -1)
		return nil
	}
	return &Routine{}
}

//Callback 只是把相关回调放到队列中
func Callback(Params []interface{}, Func func(params []interface{})) (err error) {
	if queue == nil {
		routineNMax = int32(blockageGoroutineN)
		if routineNMax == 0 {
			routineNMax = 1024 //65536
		}
		queue = channel.NewChanN(int(routineNMax)) //不能在init()里初始化，因为拿不到conf还没初始化
		logT = log.NewLogTrace(0, 0, 0)
	}
	routine := pool.Get()
	if routine == nil {
		full, closed := queue.Send(&Routine{
			Params: Params,
			Func:   Func,
		})
		if full || closed {
			if closed {
				return fmt.Errorf("Callback.Queues has been closed")
			}
			return fmt.Errorf("Callback.Queues are too long")
		}
		logT.Debugf("callback.queue goroutine is too much, put in the queue, params:%v", Params)
		return nil
	}
	r0, ok := routine.(*Routine)
	if !ok {
		logT.Errorf("callback.queue got an exception data:%v", routine)
	}
	r0.Params = Params
	r0.Func = Func
	go r0.work()
	logT.Debugf("callback.queue goroutine is enough, work in running, params:%v", Params)
	return
}

func (r *Routine) work() {
	r0, ok := r, true
	for {
		if ok {
			logT.Debugf("callback.work before run, params:%v", r.Params)
			r0.Func(r0.Params)
			logT.Debugf("callback.work after run, params:%v", r.Params)
		}
		//处理完成后，检查队列中是否有等待数据
		ifaces, closed := queue.Recv(1, 0)
		if closed || len(ifaces) == 0 {
			break //chan 已关闭或队列无数据，则退出
		}
		r0, ok = ifaces[0].(*Routine)
		if !ok {
			logT.Errorf("callback.queue got an exception data:%v", ifaces[0])
		}
	}
	pool.Put(r)
	atomic.AddInt32(&routineN, -1)
}
