package mouselib

import (
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
)

// kafka生产者/消费者out-of-box

// ProducerProvider 生产者provider，不直接使用
type ProducerProvider struct {
	producers        []sarama.AsyncProducer
	producerProvider func() sarama.AsyncProducer

	l sync.Mutex
}

// NewProducerProvider 初始化生产者provider
//
// brokers kafka地址 init 初始大小 conf 生产者配置
func NewProducerProvider(brokers string, init int, conf func() *sarama.Config) *ProducerProvider {
	provider := &ProducerProvider{
		producers: make([]sarama.AsyncProducer, 0, init),
	}
	provider.producerProvider = func() sarama.AsyncProducer {
		config := conf()
		if config.Producer.Transaction.ID == "" {
			config.Producer.Transaction.ID = uuid.New().String()
		}
		producer, err := sarama.NewAsyncProducer(strings.Split(brokers, ","), config)
		if err != nil {
			return nil
		}
		return producer
	}
	return provider
}

// Borrow 获取一个生产者
func (p *ProducerProvider) Borrow() (producer sarama.AsyncProducer) {
	p.l.Lock()
	defer p.l.Unlock()

	if len(p.producers) == 0 {
		for {
			producer = p.producerProvider()
			if producer != nil {
				return
			}
		}
	}

	idx := len(p.producers) - 1
	producer = p.producers[idx]
	p.producers = p.producers[:idx]

	return
}

// Release 释放一个生产者
func (p *ProducerProvider) Release(producer sarama.AsyncProducer) {
	p.l.Lock()
	defer p.l.Unlock()

	if producer.TxnStatus()&sarama.ProducerTxnFlagInError != 0 {
		_ = producer.Close()
		return
	}
	p.producers = append(p.producers, producer)
}

// Clear 清理生产者provider
func (p *ProducerProvider) Clear() {
	p.l.Lock()
	defer p.l.Unlock()

	for _, producer := range p.producers {
		producer.Close()
	}

	p.producers = p.producers[:0]
}

// DefaultKafkaProducerConfig 默认producer配置，保证消息消费exact-once
func DefaultKafkaProducerConfig() *sarama.Config {
	conf := sarama.NewConfig()

	conf.Net.MaxOpenRequests = 1
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Producer.Idempotent = true
	conf.Producer.Transaction.ID = "sarama"

	return conf
}

// Consumer 消费者对象，直接创建
type Consumer struct {
	Ready   chan bool
	GroupId string
	Brokers []string
	Handler func(msg *sarama.ConsumerMessage) error
}

// Setup 在一个新session启动时运行此方法
func (consumer *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	// 清除consumer的ready状态
	close(consumer.Ready)
	return nil
}

// Cleanup 在session结束时调用，发生在ConsumeClaim的goroutine退出之后
func (consumer *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 循环接受消费的消息
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 下面的代码不能在goroutine运行，本身这个函数就是在goroutine中运行的
	for {
		select {
		case message := <-claim.Messages():
			func() {
				err := consumer.Handler(message)
				if err != nil {
					session.ResetOffset(message.Topic, message.Partition, message.Offset, "")
				}
				session.MarkMessage(message, "")
				session.Commit()
			}()
		// session.Context()结束后此方法需要返回，如果没有结束会抛出`ErrRebalanceInProgress`，当kafka重新平衡（rebalance）的时候会抛出`read tcp <ip>:<port>: i/o timeout`
		case <-session.Context().Done():
			return nil
		}
	}
}

// DefaultKafkaConsumerGroupConfig 默认消费者配置，手动提交消费消息
func DefaultKafkaConsumerGroupConfig() *sarama.Config {
	conf := sarama.NewConfig()

	conf.Consumer.IsolationLevel = sarama.ReadCommitted
	conf.Consumer.Offsets.AutoCommit.Enable = false
	conf.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	conf.Consumer.Offsets.Initial = sarama.OffsetNewest

	return conf
}

// NewConsumerClient 获取消费者client
func NewConsumerClient(brokers string, groupId string, conf func() *sarama.Config) (client sarama.ConsumerGroup, err error) {
	client, err = sarama.NewConsumerGroup(strings.Split(brokers, ","), groupId, conf())

	return
}
