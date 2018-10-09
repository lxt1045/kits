package main

import (
	"common/tool/kafka"
	"fmt"
	"log"
	"time"
)

const (
	topicTest = "kafka_cluster_topic_test_02"
)

func init() {
}

func main() {
	c := kafka.KafkaConfigST{
		ServerList: []string{"192.168.31.233:9092"},
		RetryTime:  3,
		OffSet:     -1,
		Group:      "kafka_cluster_group_test_02",
		Print:      1,
	}
	err := kafka.Init(topicTest, &c)
	if err != nil {
		log.Fatalln(err)
		return
	}

	str1 := time.Now().Format("2006-01-02 15:04:05")
	for i := 0; i < 10000; i++ {
		msg := fmt.Sprintf("%s test message index:%d", str1, i)
		//e := kafka.SendMsgCluster(topicTest, []byte(msg))
		e := kafka.SendMsg(topicTest, []byte(msg))
		if e != nil {
			log.Println(e)
		}
		log.Println("send msg:", string(msg))
	}
}
