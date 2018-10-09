/*
日志BUF
	使用kafka发送日志，增加，日志buf，业务写buf，满了就丢日志
	kafka协程从buf中取日志，然后批量发送
	使用工作池的方式发送，一个topic一个协程， 暂定5条一发(平均一条日志200多字节，mtu 1460)，以后可以根据mtu调整
*/

package log

import (
	"sync"

	"github.com/lxt1045/kits/io/channel"
	"github.com/lxt1045/kits/kafka"
	//	"github.com/optiopay/kafka/proto"
)

const (
	WR_LOG_CODE_SUCC      = 0
	WR_LOG_CODE_INPUT_ERR = 1
	WR_LOG_CODE_FULL      = 2
)

const (
	MAX_BUF_LEN = 10240
)

var topicMapBuf sync.Map
var kafkaClient *kafka.Client

func Write(topic string, msg string) int {
	if "" == topic || 0 == len(msg) {
		return WR_LOG_CODE_INPUT_ERR
	}

	iface, ok := topicMapBuf.Load(topic)
	ch, ok1 := iface.(*channel.ChanN)
	if !ok || !ok1 {
		ch = channel.NewChanN(MAX_BUF_LEN)
		topicMapBuf.Store(topic, ch)
		go push2Kafka(topic, ch)
	}
	full, closed := ch.Send(msg)
	if full || closed {
		return WR_LOG_CODE_SUCC
	}

	return WR_LOG_CODE_FULL
}

const (
	BATCH_SEND_NUM = 5
)

func push2Kafka(topic string, ch channel.IChanN) {
	for {
		logs, closed := ch.Recv(BATCH_SEND_NUM, 1)
		if closed {
			return
		}
		if 0 != len(logs) && kafkaClient != nil {
			g_count.SendLog()
			for _, logstr := range logs {
				kafkaClient.SendMsg(topic, []byte(logstr.(string)))
				//SendMsgAsync(topic, []byte(logstr.(string)))
			}
		}
	}
}
