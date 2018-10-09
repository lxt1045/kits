package log

import (
	"fmt"
	"time"
)

type TraceInfoST struct {
	ParentID int64 `json:"parent_id"` //用于标识同一个服务的所有log，便于同一个服务名下的不同服务，IP不安全
	SpanID   int64 `json:"span_id"`   //自己的调用链ID，比如和connID对应
	TraceID  int64 `json:"trace_id"`  //用于标识本次请求，比如一个消息一个traceID
	FlowID   int64 `json:"flow_id"`   //处理流程模块的ID，便于分模块检查日志，比如io:1, logic：2
	//Tag      string `json:"tag"`
}

//NewLogTrace ParentID :用于标识同一条连接的所有请求, SpanID : 自己的调用链ID, TraceID : 用于标识本次请求
func NewLogTrace(traceid, spanid, parentid int64) (trace *TraceInfoST) {
	if 0 == traceid {
		traceid = time.Now().UnixNano() / int64(time.Millisecond)
	}
	trace = &TraceInfoST{
		ParentID: parentid,
		SpanID:   spanid,
		TraceID:  traceid,
	}
	// runtime.SetFinalizer(trace, func(d *TraceInfoST) {
	// 	fmt.Printf("a %p final.\n", d)
	// 	topicMapBuf.Range(func(k, v interface{}) bool {
	// 		fmt.Println(k, v)
	//
	// 		ch, ok := v.(*channel.ChanN)
	// 		if !ok {
	// 			return true
	// 		}
	// 		ch.Close()
	// 		return true
	// 	})
	// })
	return
}

func (p *TraceInfoST) Debugf(format string, args ...interface{}) {
	store(stackDepthLogTrace, DEBUG, p, fmt.Sprintf(format, args...))
}
func (p *TraceInfoST) Debug(args ...interface{}) {
	store(stackDepthLogTrace, DEBUG, p, fmt.Sprint(args...))
}

func (p *TraceInfoST) Infof(format string, args ...interface{}) {
	store(stackDepthLogTrace, INFO, p, fmt.Sprintf(format, args...))
}
func (p *TraceInfoST) Info(args ...interface{}) {
	store(stackDepthLogTrace, INFO, p, fmt.Sprint(args...))
}

func (p *TraceInfoST) Warningf(format string, args ...interface{}) {
	store(stackDepthLogTrace, WARNING, p, fmt.Sprintf(format, args...))
}
func (p *TraceInfoST) Warning(args ...interface{}) {
	store(stackDepthLogTrace, WARNING, p, fmt.Sprint(args...))
}

func (p *TraceInfoST) Errorf(format string, args ...interface{}) {
	store(stackDepthLogTrace, ERROR, p, fmt.Sprintf(format, args...))
}
func (p *TraceInfoST) Error(args ...interface{}) {
	store(stackDepthLogTrace, ERROR, p, fmt.Sprint(args...))
}

func (p *TraceInfoST) Criticalf(format string, args ...interface{}) {
	store(stackDepthLogTrace, CRITICAL, p, fmt.Sprintf(format, args...))
}
func (p *TraceInfoST) Critical(args ...interface{}) {
	store(stackDepthLogTrace, CRITICAL, p, fmt.Sprint(args...))
}
func (p *TraceInfoST) Print(args ...interface{}) {
	store(stackDepthDefault, DEBUG, p, fmt.Sprint(args...))
}
func (p *TraceInfoST) Println(args ...interface{}) {
	store(stackDepthDefault, DEBUG, p, fmt.Sprint(args...))
}
func (p *TraceInfoST) Printf(format string, args ...interface{}) {
	store(stackDepthDefault, DEBUG, p, fmt.Sprintf(format, args...))
}

const stackDepthDefault = 3
const stackDepthLogTrace = 3

var defaultTrace TraceInfoST

func Debugf(format string, args ...interface{}) {
	store(stackDepthDefault, DEBUG, &defaultTrace, fmt.Sprintf(format, args...))
}
func Debug(args ...interface{}) {
	store(stackDepthDefault, DEBUG, &defaultTrace, fmt.Sprint(args...))
}

func Infof(format string, args ...interface{}) {
	store(stackDepthDefault, INFO, &defaultTrace, fmt.Sprintf(format, args...))
}
func Info(args ...interface{}) {
	store(stackDepthDefault, INFO, &defaultTrace, fmt.Sprint(args...))
}

func Warningf(format string, args ...interface{}) {
	store(stackDepthDefault, WARNING, &defaultTrace, fmt.Sprintf(format, args...))
}
func Warning(args ...interface{}) {
	store(stackDepthDefault, WARNING, &defaultTrace, fmt.Sprint(args...))
}

func Errorf(format string, args ...interface{}) {
	store(stackDepthDefault, ERROR, &defaultTrace, fmt.Sprintf(format, args...))
}
func Error(args ...interface{}) {
	store(stackDepthDefault, ERROR, &defaultTrace, fmt.Sprint(args...))
}

func Criticalf(format string, args ...interface{}) {
	store(stackDepthDefault, CRITICAL, &defaultTrace, fmt.Sprintf(format, args...))
}
func Critical(args ...interface{}) {
	store(stackDepthDefault, CRITICAL, &defaultTrace, fmt.Sprint(args...))
}

func Print(args ...interface{}) {
	store(stackDepthDefault, DEBUG, &defaultTrace, fmt.Sprint(args...))
}
func Println(args ...interface{}) {
	store(stackDepthDefault, DEBUG, &defaultTrace, fmt.Sprint(args...))
}
func Printf(format string, args ...interface{}) {
	store(stackDepthDefault, DEBUG, &defaultTrace, fmt.Sprintf(format, args...))
}
