package mouselib_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/mouseleee/mouselib"
)

var brokers = "localhost:9093"

func TestCreateTopic(t *testing.T) {
	t.FailNow()
	err := mouselib.CreateTopic(brokers, "mousetest", 1, 1)
	if err != nil {
		t.Error(err)
	}
}

func TestKafkaProducerConsumer(t *testing.T) {
	topic := "mousetest"
	prd, err := mouselib.DefaultProducer(brokers)
	if err != nil {
		t.Error(err)
	}
	defer prd.Close()
	csm, err := mouselib.DefaultConsumer(brokers, "mousegroup")
	if err != nil {
		t.Error(err)
	}
	defer csm.Close()

	count := 100

	for count > 0 {
		_, _, err := prd.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(fmt.Sprintf("这是第%d条测试消息", count)),
		})
		if err != nil {
			t.Error(err)
		}
		count--
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		err := csm.Consume(ctx, []string{topic}, &MyConsumer{})
		if err != nil {
			t.Error(err)
		}
	}
}

type MyConsumer struct {
}

func (c *MyConsumer) Setup(session sarama.ConsumerGroupSession) error {
	println("init...")
	return nil
}

func (c *MyConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *MyConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg := <-claim.Messages():
			println(string(msg.Value))
			session.Commit()
		case <-session.Context().Done():
			return nil
		}
	}
}
