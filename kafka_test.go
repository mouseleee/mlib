package mouselib_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/mouseleee/mouselib"
)

func TestKafkaProducer(t *testing.T) {
	// t.FailNow()
	brokers := "localhost:9093"
	topic := "mewo"

	provider := mouselib.NewProducerProvider(brokers, 1, mouselib.DefaultKafkaProducerConfig)

	producer := provider.Borrow()
	defer provider.Clear()
	defer provider.Release(producer)

	for i := 0; i < 100; i++ {
		// BeginTxn must be called before any messages.
		err := producer.BeginTxn()
		if err != nil {
			t.Logf("Message consumer: unable to start transaction: %+v", err)
			t.Error(err)
		}
		// Produce current record in producer transaction.
		producer.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(fmt.Sprintf("%d fix message", i)),
		}

		// Commit producer transaction.
		err = producer.CommitTxn()
		if err != nil {
			t.Log("error on CommitTxn")
			t.Error(err)
		}
	}
}

func TestKafkaConsumer(t *testing.T) {
	t.FailNow()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		brokers := "localhost:9093"
		groupId := "test"

		client, err := mouselib.NewConsumerClient(brokers, groupId, mouselib.DefaultKafkaConsumerGroupConfig)
		if err != nil {
			t.Error(err)
		}

		consumer1 := mouselib.DefaultConsumer{
			Ready:   make(chan bool),
			GroupId: "test",
		}

		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			for {
				// `Consume` should be called inside an infinite loop, when a
				// server-side rebalance happens, the consumer session will need to be
				// recreated to get the new claims
				if err := client.Consume(ctx, strings.Split("mewo", ","), &consumer1); err != nil {
					t.Errorf("Error from consumer: %v", err)
				}
				// check if context was cancelled, signaling that the consumer should stop
				if ctx.Err() != nil {
					return
				}
				consumer1.Ready = make(chan bool)
			}
		}()

		<-consumer1.Ready
		wg.Wait()

		cancel()
	}()

	<-sig
}
