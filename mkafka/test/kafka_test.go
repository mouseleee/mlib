package mkafka_test

import (
	"os"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/mouseleee/mlib/mkafka"
	"github.com/rs/zerolog"
)

var brokers = "localhost:9093"

func TestSetLogger(t *testing.T) {
	tl := zerolog.New(os.Stdout)
	mkafka.SetLogger(tl, true)
	mkafka.SetLogger(tl, false)
}

func TestDefaultProducerConfig(t *testing.T) {
	mkafka.DefaultProducerConfig()
}

func TestDefaultProducer(t *testing.T) {
	_, err := mkafka.DefaultProducer(brokers)
	if err != nil {
		t.Error(err)
	}
}

func TestDefaultConsumerConfig(t *testing.T) {
	mkafka.DefaultConsumerConfig()
}

func TestDefaultConsumerGroup(t *testing.T) {
	_, err := mkafka.DefaultConsumerGroup(brokers, "test.group")
	if err != nil {
		t.Error(err)
	}
}

func TestCustomConsumerGroup(t *testing.T) {
	_, err := mkafka.CustomConsumerGroup(brokers, "test.group", sarama.NewConfig())
	if err != nil {
		t.Error(err)
	}
}

func TestTopicOp(t *testing.T) {
	cli, err := mkafka.CreateKafkaClient(brokers, sarama.NewConfig())
	if err != nil {
		t.Error(err)
	}
	defer cli.Close()

	err = mkafka.CreateTopic(cli, "TEST_TOPIC", 1, 1)
	if err != nil {
		t.Error(err)
	}

	err = mkafka.RemoveTopic(cli, []string{"TEST_TOPIC"})
	if err != nil {
		t.Error(err)
	}
}

// func TestKafkaProduceMsg(t *testing.T) {
// 	sarama.Logger = log.New(os.Stdout, "sarama=>", 0)
// 	topic := "mousetest"
// 	prd, err := mkafka.DefaultProducer(brokers)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	defer prd.Close()

// 	count := 100

// 	for count > 0 {
// 		_, _, err := prd.SendMessage(&sarama.ProducerMessage{
// 			Topic: topic,
// 			Value: sarama.StringEncoder(fmt.Sprintf("这是第%d条测试消息", count)),
// 		})
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		count--
// 	}
// }

// func TestKafkaConsumeMsg(t *testing.T) {
// 	topic := "mousetest"
// 	csm, err := mkafka.DefaultConsumer(brokers, "mousegroup")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	defer csm.Close()

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	for {
// 		err := csm.Consume(ctx, []string{topic}, &MyConsumer{})
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	}
// }

// type MyConsumer struct {
// }

// func (c *MyConsumer) Setup(session sarama.ConsumerGroupSession) error {
// 	println("init...")
// 	return nil
// }

// func (c *MyConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
// 	return nil
// }

// func (c *MyConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
// 	for {
// 		select {
// 		case msg := <-claim.Messages():
// 			println(string(msg.Value))
// 			session.Commit()
// 		case <-session.Context().Done():
// 			return nil
// 		}
// 	}
// }
