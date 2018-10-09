package main

import (
	"common/tool/kafka"
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

	num := 0

	for {
		v, e := kafka.RecvMsg(topicTest)
		if e != nil {
			log.Fatalln(e)
		}
		log.Println("recv msg:", string(v))

		num += 1
		//		if num >= 20 {
		//			break
		//		}

	}

	time.Sleep(5 * time.Second)
}
