package log

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/lxt1045/kits/kafka"
)

const (
	NIL_PARENT = "nil"
)

// 默认参数的打印，这样不会有业务ID，在公共的地方打印, 在有业务的地方不建议使用
var (
	localIP    string
	levelArray []string
	//levelMap   map[int]string
)

const (
	CRITICAL = iota
	ERROR
	WARNING
	INFO
	DEBUG

	LevelMax
)

var (
	topic    string // 写日志的kafka topic
	level    int    // 写日志的最大等级（含设置等级）， 4(debug)<3(info)<2(warning)<1(error)<0(critical)
	cmdPrint int    // 是否打印到控制台， 0不打印， 1打印
	timeZone *time.Location

	srvName string // 服务启动得到的服务ID,比如g_sence
)

func init() {
	localIP = getLocalIP()
	// levelMap = map[int]string{
	// 	//  CRITICAL: "CRITICAL",
	// 	//  ERROR:    "ERROR",
	// 	//  WARNING:  "WARNING",
	// 	//  INFO:     "INFO",
	// 	//  DEBUG:    "DEBUG",
	// 	CRITICAL: "\033[35m[CRITICAL]\033[0m",
	// 	ERROR:    "\033[31m[ERROR]\033[0m",
	// 	WARNING:  "\033[33m[WARNING]\033[0m",
	// 	INFO:     "\033[36m[INFO]\033[0m",
	// 	DEBUG:    "\033[32m[DEBUG]\033[0m",
	// }
	levelArray = make([]string, LevelMax)
	{
		levelArray[CRITICAL] = "\033[35m[CRITICAL]\033[0m"
		levelArray[ERROR] = "\033[35m[ERROR]\033[0m"
		levelArray[WARNING] = "\033[35m[WARNING]\033[0m"
		levelArray[INFO] = "\033[35m[INFO]\033[0m"
		levelArray[DEBUG] = "\033[35m[DEBUG]\033[0m"
	}
	level = DEBUG
	cmdPrint = 1
	srvName = "g_log_test"

	var err error
	timeZone, err = time.LoadLocation("Asia/Chongqing")
	if err != nil {
		timeZone = time.FixedZone("CST", 28800)
	}
}

type Config struct {
	Topic string         // kafka服务器地址
	Level int            // 发送超时时间， 单位毫秒
	Print int            // 重传次数， 消息系统里面快速失败，不保证数据的高可到达性
	TimeZ *time.Location //时区：time zone
}

func Init(srvID string, conf *Config, kconf *kafka.Config) (err error) {
	if conf == nil || kconf == nil {
		srvName = srvID
		return fmt.Errorf("conf:%v, kconf:%v", conf, kconf)
	}
	topic = conf.Topic
	level = conf.Level
	cmdPrint = conf.Print
	if conf.TimeZ != nil {
		timeZone = conf.TimeZ
	} else {
		//timeZone, err = time.LoadLocation("Asia/Chongqing")
	}
	srvName = srvID

	// err := KafkaInit(srv_id, kconf)
	// if nil != err {
	// 	return fmt.Errorf("kafka.client.Init(%+v).%v", kconf, err)
	// }
	kafkaClient, err = kafka.NewClient([]string{topic}, kconf, nil)
	if nil != err {
		return fmt.Errorf("kafka.client.Init(%+v).%v", kconf, err)
	}
	return nil
}

func lastFname(fname string) string {
	flen := len(fname)
	n := strings.LastIndex(fname, ".")
	if n+1 < flen {
		return fname[n+1:]
	}

	return fname
}

func opid() string {
	return "nil" //fmt.Sprintf("%s_%d", g_conf.srvName, atomic.AddUint64(&g_id, 1))
}

func getLocalIP() (ip string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "110.120.130.140"
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = fmt.Sprintf("%s&%s", ip, ipnet.IP.String())
			}
		}
	}
	return
}

//
// 格式化 log 层
//
func store(stackDepth, l int, trace *TraceInfoST, msg string) {
	if level < l {
		return
	}
	g_count.RecvLog()
	lmsg := logFormat(stackDepth, l, trace, msg)
	sendLog(lmsg)

	// 日志等级高于Err，直接打印，不在受条件控制
	if cmdPrint != 0 || l < WARNING {
		if lmsg[len(lmsg)-1] == '\n' {
			lmsg = string(lmsg[:len(lmsg)-1])
		}
		fmt.Println(lmsg)
	}
}

// 写日志就不返回错误了，直接控制台打印
func sendLog(lmsg string) int {
	result := Write(topic, lmsg)
	if WR_LOG_CODE_SUCC != result {
		g_count.BufErr()
	}

	return result
}

func logFormat(stackDepth, level int, trace *TraceInfoST, msg string) string {
	funcName := ""
	pc, file, line, ok := runtime.Caller(stackDepth) // 去掉两层，当前函数和日志的接口函数
	if ok {
		if f := runtime.FuncForPC(pc); f != nil {
			funcName = f.Name()
		}
	}

	tNow := time.Now().In(timeZone).Format("2006-01-02 15:04:05.000")
	//					t lev ip tid pid fl
	return fmt.Sprintf("%s %s %s [%d] %d %s %s %s:%d [%d,%d,%d] - %s",
		tNow,                   //时间：timestamp， YYYY-MM-DD HH:MM:SS.sss
		levelArray[level],      //日志级别：DEBUG、ERROR、INFO...////levelMap[level]
		localIP,                //本地IP：ip
		runtime.NumGoroutine(), //[线程ID]： thread, 这里显示当前goroutine数量
		trace.FlowID,           //PID:进程ID,pid, 这里用流程ID替代
		srcPath(file),          //调用位置：文件名; src/services/D_Go_V2.0_sence/kit/log/log_test.go
		srvName,                //服务名称：  ServerID..
		lastFname(funcName),    //调用的方法名：methodName; funcName:services/D_Go_V2.0_sence/kit/log_test.TestLog
		line,                   //调用位置：行号
		trace.TraceID,          //traceId
		trace.SpanID,           //spanId
		trace.ParentID,         //parentSpanId
		msg,                    //msg：消息主体
	)
}
func srcPath(fullPath string) string {
	i := strings.Index(fullPath, "/src/")
	return fullPath[i+1:]
	//return filepath.Base(fullPath) //调用位置：文件名 //"path/filepath"
}
