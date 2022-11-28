package mkafka

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
)

// 提供一个开箱即用的kafka生产者/消费者实例，保证exactly-once的消费逻辑
//
// 提供kafka客户端的部分功能，包括topic和consumer-group相关

var logger zerolog.Logger = zerolog.New(os.Stdout).With().Time("time", time.Now()).Caller().Logger()

func SetLogger(out zerolog.Logger, saraLog bool) {
	logger = out
	if saraLog {
		sarama.Logger = log.New(os.Stdout, "sarama->", log.Flags())
	} else {
		sarama.Logger = log.New(io.Discard, "[Sarama] ", log.LstdFlags)
	}
}

// DefaultProducerConfig 默认生产者配置
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

// DefaultProducer 默认生产者
func DefaultProducer(brokers string) (sarama.SyncProducer, error) {
	addrs := strings.Split(brokers, ",")
	prd, err := sarama.NewSyncProducer(addrs, DefaultProducerConfig())
	if err != nil {
		logger.Err(err).Msg("创建默认生产者失败")
		return nil, err
	}
	return prd, nil
}

// DefaultConsumerConfig 默认消费者组配置
func DefaultConsumerConfig() *sarama.Config {
	conf := sarama.NewConfig()
	conf.Consumer.Offsets.AutoCommit.Enable = false
	conf.Consumer.Offsets.Initial = sarama.OffsetOldest
	return conf
}

// DefaultConsumerGroup 默认消费者组
func DefaultConsumerGroup(brokers string, group string) (sarama.ConsumerGroup, error) {
	addrs := strings.Split(brokers, ",")
	csm, err := sarama.NewConsumerGroup(addrs, group, DefaultProducerConfig())
	if err != nil {
		logger.Err(err).Msg("创建默认消费者组失败")
		return nil, err
	}
	return csm, nil
}

// CustomConsumerGroup 根据传入的config参数创建消费者组
func CustomConsumerGroup(brokers string, group string, config *sarama.Config) (sarama.ConsumerGroup, error) {
	addrs := strings.Split(brokers, ",")
	csm, err := sarama.NewConsumerGroup(addrs, group, config)
	if err != nil {
		logger.Err(err).Msg("创建自定义消费者组失败")
		return nil, err
	}
	return csm, nil
}

// CreateKafkaClient 创建kafka客户端，使用后需关闭
func CreateKafkaClient(brokers string, config *sarama.Config) (sarama.Client, error) {
	addr := strings.Split(brokers, ",")
	return sarama.NewClient(addr, config)
}

// CreateTopic 创建topic
func CreateTopic(client sarama.Client, topic string, partition int32, replica int16) error {
	ctrlr, err := client.RefreshController()
	if err != nil {
		logger.Err(err).Msg("获取Controller失败")
		return err
	}
	defer ctrlr.Close()

	rsp, err := ctrlr.CreateTopics(&sarama.CreateTopicsRequest{
		TopicDetails: map[string]*sarama.TopicDetail{
			topic: {
				NumPartitions:     partition,
				ReplicationFactor: replica,
			},
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		logger.Err(err).Msg("发送创建Topic请求失败")
		return err
	}

	for t, e := range rsp.TopicErrors {
		if e != nil {
			logger.Err(e).Str("topic", t).Msg("创建Topic失败")
			return e
		}
	}
	return nil
}

// ResetConsumerGroupOffset 通过消费者组删除对某个topic的某些partition的offset
func ResetConsumerGroupOffset(client sarama.Client, group string, allTopic bool, partitions map[string][]int32) error {
	err := client.RefreshCoordinator(group)
	if err != nil {
		logger.Err(err).Msg("刷新coordinator失败")
		return err
	}
	crdntr, err := client.Coordinator(group)
	if err != nil {
		logger.Err(err).Msg("获取coordinator失败")
		return err
	}
	defer crdntr.Close()

	req := new(sarama.DeleteOffsetsRequest)
	req.Group = group
	for k, v := range partitions {
		for _, sv := range v {
			req.AddPartition(k, sv)
		}
	}
	res, err := crdntr.DeleteOffsets(req)
	if err != nil {
		logger.Err(err).Msg("发送删除指定topic的offset请求失败")
		return err
	}

	for k, e := range res.Errors {
		for p, se := range e {
			if se != 0 {
				logger.Err(se).Str("topic", k).Int32("partition", p).Msg("删除offset失败")
				return se
			}
		}
	}

	return nil
}

func DeleteConsumerGroup(client sarama.Client, groups []string) error {
	req := new(sarama.DeleteGroupsRequest)
	req.Groups = groups

	ctrlr, err := client.RefreshController()
	if err != nil {
		logger.Err(err).Msg("获取controller失败")
		return err
	}

	rsp, err := ctrlr.DeleteGroups(req)
	if err != nil {
		logger.Err(err).Msg("发送删除消费者组请求失败")
		return err
	}

	for k, e := range rsp.GroupErrorCodes {
		if e != 0 {
			logger.Err(e).Str("group", k).Msg("删除消费者组失败")
			return e
		}
	}

	return nil
}
