package mouselib

import (
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

// 提供一个默认的exactly-once的生产者和消费者，并且提供自定义配置来生成生产者盒消费者
//
// 提供修改kafka集群配置，topic操作的方便api

type KafkaErr struct {
	msg string
	err error
}

func (e KafkaErr) Error() string {
	return e.msg + " => " + e.err.Error()
}

func KafkaError(msg string, err error) KafkaErr {
	return KafkaErr{msg: msg, err: err}
}

func DefaultProducerConfig() *sarama.Config {
	conf := sarama.NewConfig()
	conf.Producer.Idempotent = true
	conf.Net.MaxOpenRequests = 1
	conf.Producer.Return.Successes = true
	conf.Producer.Retry.Max = 3
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	return conf
}

func DefaultProducer(brokers string) (sarama.SyncProducer, error) {
	addrs := strings.Split(brokers, ",")
	prd, err := sarama.NewSyncProducer(addrs, DefaultProducerConfig())
	if err != nil {
		return nil, KafkaError("创建生产者失败", err)
	}
	return prd, nil
}

func DefaultConsumerConfig() *sarama.Config {
	conf := sarama.NewConfig()
	conf.Consumer.Offsets.AutoCommit.Enable = false
	conf.Consumer.Offsets.Initial = sarama.OffsetOldest
	return conf
}

func DefaultConsumer(brokers string) (sarama.Consumer, error) {
	addrs := strings.Split(brokers, ",")
	prd, err := sarama.NewConsumer(addrs, DefaultProducerConfig())
	if err != nil {
		return nil, KafkaError("创建消费者失败", err)
	}
	return prd, nil
}

// CreateTopic 创建topic
func CreateTopic(brokers string, topic string, partition int32, replica int16) error {
	addrs := strings.Split(brokers, ",")
	cli, err := sarama.NewClient(addrs, sarama.NewConfig())
	if err != nil {
		return KafkaError("初始化Kafka客户端失败", err)
	}
	defer cli.Close()

	ctrlr, err := cli.Controller()
	if err != nil {
		return KafkaError("", err)
	}

	rsp, err := ctrlr.CreateTopics(&sarama.CreateTopicsRequest{
		TopicDetails: map[string]*sarama.TopicDetail{
			topic: {
				NumPartitions:     partition,
				ReplicationFactor: replica,
			},
		},
		Timeout: 3 * time.Second,
	})
	if err != nil {
		return KafkaError("创建Topic请求发送失败", err)
	}
	err = nil
	for t, e := range rsp.TopicErrors {
		if e != nil {
			err = KafkaError("topic "+t+"创建失败", e)
		}
	}
	return err
}
