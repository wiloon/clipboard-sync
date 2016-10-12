package main

import (
	"github.com/atotto/clipboard"
	"fmt"
	"time"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"log"
)

var (
	producer sarama.SyncProducer
)

func init() {

}

func main() {
	//start goroutine to send the content of local clipboard to mq
	var wg sync.WaitGroup
	wg.Add(2)
	go localClipboardChangeMonitor(&wg)

	go clipboardSyncRemote(&wg)
	wg.Wait()
	//check local clipboard
	//send to mq

	// start goroutine to receive the message from mq
	// check if update local clipboard
	//update local clipboard
}
func clipboardSyncRemote(wg *sync.WaitGroup) {
	defer wg.Done()
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Specify brokers address. This is default one
	brokers := []string{"localhost:9092"}

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	topic := "important"
	// How to decide partition, is it fixed value...?
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	for {
		msg := consumer.Messages()

		fmt.Println("consume msg:", <-msg)
		time.Sleep(1000 * time.Millisecond)
	}
}

func localClipboardChangeMonitor(wg *sync.WaitGroup) {
	defer wg.Done()

	brokerList := strings.Split("localhost:9092", ",")
	log.Printf("Kafka brokers: %s", strings.Join(brokerList, ", "))

	// For the data collector, we are looking for strong consistency semantics.
	// Because we don't change the flush settings, sarama will try to produce messages
	// as fast as possible to keep latency low.
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message


	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	var tmpMsg string
	for {
		msg, err := clipboard.ReadAll()
		//fmt.Println("msg:", msg)
		if err != nil {
			fmt.Println("failed to read clipboard")
		}
		if msg != ""&& !strings.EqualFold(msg, tmpMsg) {
			tmpMsg = msg
			fmt.Println("clipboard change, msg:", msg)
			partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
				Topic: "test",
				Value: sarama.StringEncoder(msg),
			})
			if err != nil {
				//

			} else {
				fmt.Println("partition:", partition)
				fmt.Println("offect:", offset)
			}
		}

		time.Sleep(1000 * time.Millisecond)
	}
}


