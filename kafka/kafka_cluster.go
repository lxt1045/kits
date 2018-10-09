package kafka

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	slog "log"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
)

type Config struct {
	ServerList  []string // kafka服务器地址
	SendTimeOut int      // 发送超时时间， 单位毫秒
	RetryTime   int      // 重传次数， 消息系统里面快速失败，不保证数据的高可到达性
	RetryWait   int      // 重传的等待时间, 单位毫秒
	Print       int      // 错误打印在串口
	OffSet      int64    // 启动时，接收消息的位置， -1 从最新的开始读， -2：从上次的问题开始读
	Group       string   // 读消息的组配置
}
type Client struct {
	serverList  []string // kafka服务器地址
	sendTimeOut int      // 发送超时时间， 单位毫秒
	retryTime   int      // 重传次数， 消息系统里面快速失败，不保证数据的高可到达性
	retryWait   int      // 重传的等待时间, 单位毫秒
	cmdPrint    int      // 错误打印在串口
	offSet      int64    // 启动时，接收消息的位置， -1 从最新的开始读， -2：从上次的问题开始读
	group       string   // 读消息的组配置
	topics      []string

	consumer      *cluster.Consumer //消费者
	producer      *sarama.SyncProducer
	producerAsync *sarama.AsyncProducer

	log sarama.StdLogger
}

// NewClient 新建Kafka需要通过config，所有必须通过此函数创建；
//有一个问题，logger无法在请求的时候返回，是否需要一个全局的logger？
func NewClient(topics []string, conf *Config, log sarama.StdLogger) (client *Client, e error) {
	if log == nil {
		log = slog.New(os.Stderr, "", slog.LstdFlags)
	}
	if 0 == len(conf.ServerList) {
		log.Print("kafka server is nil")
		return
	}
	client = &Client{
		serverList:  conf.ServerList,
		sendTimeOut: conf.SendTimeOut,
		retryTime:   conf.RetryTime,
		retryWait:   conf.RetryWait,
		cmdPrint:    conf.Print,
		offSet:      conf.OffSet,
		group:       conf.Group,
		topics:      topics,

		log: log,
		//consumer      *cluster.Consumer //消费者
		//producer      *sarama.SyncProducer
		//producerAsync *sarama.AsyncProducer

	}

	//初始化发送端
	sarama.Logger = log

	//监听退出事件，可以处理一下连接问题，避免kafka服务器重新负载延迟问题
	//这个在windows下无法工作，但是在linux下是有效的
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		v := <-signals
		log.Println("%s kafka cluster signal(%v) closed...", time.Now().Format("2006-01-02 15:04:05"), v)

		if client.producerAsync != nil {
			(*client.producerAsync).Close()
		}
		if client.producer != nil {
			(*client.producer).Close()
		}

		if client.consumer != nil {
			//consumer.CommitOffsets()
			client.consumer.CommitOffsets()
			client.consumer.Close()
		}
		os.Exit(0)
	}()

	//defer consumer.Close()
	return
}

func (p *Client) RecvMsg() ([]byte, error) {
	if p.consumer == nil {
		config := cluster.NewConfig()
		config.ClientID = "golang_kafa_client_recv"
		config.Consumer.Return.Errors = true
		config.Group.Return.Notifications = true
		config.Consumer.Offsets.CommitInterval = 1 * time.Second
		config.Consumer.Offsets.Initial = p.offSet
		//config.Consumer.MaxWaitTime = time.Duration(g_config.RetryTime) * time.Millisecond

		consumer, err := cluster.NewConsumer(p.serverList, p.group, p.topics, config)
		if err != nil {
			return []byte{}, fmt.Errorf("cluster.NewConsumer get error:%v\n", err)
		}
		// consume errors
		go func() {
			for err := range consumer.Errors() {
				if p.cmdPrint != 0 {
					p.log.Printf("%s kafka cluster,topic:%s, group:%s, got error: %s\n",
						time.Now().Format("2006-01-02 15:04:05"), p.topics, p.group, err.Error())
				}
			}
		}()

		// consume notifications
		go func() {
			for ntf := range consumer.Notifications() {
				if p.cmdPrint != 0 {
					p.log.Printf("%s kafka cluster,topic:%s, group:%s, rebalanced: %+v\n",
						time.Now().Format("2006-01-02 15:04:05"), p.topics, p.group, ntf)
				}
			}
		}()

		p.consumer = consumer
	}

	// consume messages, watch signals
	msg, ok := <-p.consumer.Messages()
	if ok {
		if p.cmdPrint != 0 {
			p.log.Printf("%s kafka cluster topic:%s partition:%d offset:%d key:%s value:%s\n",
				time.Now().Format("2006-01-02 15:04:05"), msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
		}
		p.consumer.MarkOffset(msg, "") ////MarkOffset并不是实时写入kafka，有可能在程序crash时丢掉未提交的offset

		return msg.Value, nil
	}

	return []byte{}, fmt.Errorf("kafaka cluster got error")
}

// 同步发送接口
func (p *Client) SendMsg(topic string, msgIn []byte) error {
	if p.producer == nil {
		config := sarama.NewConfig()
		config.ClientID = "golang_kafa_client_send"
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner //sarama.NewRandomPartitioner
		config.Producer.Return.Successes = true
		config.Producer.Return.Errors = true

		producer, err := sarama.NewSyncProducer(p.serverList, config)
		if err != nil {
			return fmt.Errorf("Failed to produce message: %s", err)
		}

		p.producer = &producer
	}

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: int32(-1),
		Key:       sarama.StringEncoder("key"),
		Value:     sarama.ByteEncoder(msgIn),
	}
	partition, offset, err := (*p.producer).SendMessage(msg)
	if err != nil {
		return fmt.Errorf("Failed to produce message: %v", err)
	}
	_, _ = partition, offset
	//fmt.Printf("partition=%d, offset=%d, message:%s\n", partition, offset, string(msgIn))

	return nil
}

// 同步发送接口
func (p *Client) SendMsgAsync(topic string, msgIn []byte) error {
	if p.producerAsync == nil {
		config := sarama.NewConfig()
		config.ClientID = "golang_kafa_client_send_async"
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner //sarama.NewRandomPartitioner
		config.Producer.Return.Successes = true
		config.Producer.Return.Errors = true

		producerAsync, err := sarama.NewAsyncProducer(p.serverList, config)
		if err != nil {
			return fmt.Errorf("Failed to produce message: %s", err)
		}
		//异步发送必须有这个匿名函数内容, 处理发送成功果实失败的结果
		go func(pp sarama.AsyncProducer) {
			errors := pp.Errors()
			success := pp.Successes()
			for {
				select {
				case err := <-errors:
					if err != nil {
						if p.cmdPrint != 0 {
							p.log.Printf("producer.%v\n", err)
						}
					}
				case <-success:
				}
			}
		}(producerAsync)
	}
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: int32(-1),
		Key:       sarama.StringEncoder("key"),
		Value:     sarama.ByteEncoder(msgIn),
	}
	(*p.producerAsync).Input() <- msg
	return nil
}
