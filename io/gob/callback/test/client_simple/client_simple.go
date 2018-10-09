package main

import (
	//"log"
	//"net/rpc"
	"encoding/json"
	"time"

	"github.com/lxt1045/geocache/msg"
	"github.com/lxt1045/kits/io/gob"
	"github.com/lxt1045/kits/log"
)

func main() {
	client, err := gob.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Error(err)
	}

	arg := msg.Ack{UserID: 1001, TimeStamp: time.Now().UnixNano()}
	client.GetReply(arg, &msg.Ack{}, 2, time.Now().UnixNano()+int64(time.Second*10),
		func(reply interface{}, err error) {
			if err != nil {
				log.Error(err)
			}
			js, _ := json.Marshal(reply)
			log.Info(string(js))
		},
	)
	time.Sleep(time.Second * 10)
}
