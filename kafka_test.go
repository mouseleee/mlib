package mouselib_test

import (
	"fmt"
	"sync"
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
	csm, err := mouselib.DefaultConsumer(brokers)
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

	parts, err := csm.Partitions(topic)
	if err != nil {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	for _, part := range parts {
		wg.Add(1)
		go func(p int32) {
			defer wg.Done()
			pcsm, err := csm.ConsumePartition(topic, p, sarama.OffsetNewest)
			if err != nil {
				t.Error(err)
			}
			defer pcsm.Close()

			for {
				select {
				case msg := <-pcsm.Messages():
					println(string(msg.Value))
				case err := <-pcsm.Errors():
					t.Error(err)
				}
			}
		}(part)
	}
	wg.Wait()
}
