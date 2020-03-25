package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"time"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis.wiloon.com:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	go subscribe(client)
	go publish(client)

	time.Sleep(10 * time.Second)
}

func publish(client *redis.Client) {
	// Publish a message.
	for {
		err := client.Publish("channel0", "hello").Err()
		if err != nil {
			panic(err)
		}
		time.Sleep(3 * time.Second)
	}

}

func subscribe(client *redis.Client) {
	pubsub := client.Subscribe("channel0")

	// Wait for confirmation that subscription is created before publishing anything.
	_, err := pubsub.Receive()
	if err != nil {
		panic(err)
	}

	// Go channel which receives messages.
	ch := pubsub.Channel()
	// Consume messages.
	for msg := range ch {
		fmt.Println(msg.Channel, msg.Payload)
	}
}
