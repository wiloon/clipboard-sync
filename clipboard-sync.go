package main

import (
	"github.com/atotto/clipboard"
	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var logger *zap.SugaredLogger
var localClipboardContent string
var signals chan os.Signal

func main() {
	zapLogger, _ := zap.NewProduction()

	defer zapLogger.Sync() // flushes buffer, if any
	logger = zapLogger.Sugar()

	client := redis.NewClient(&redis.Options{
		Addr:     "redis.wiloon.com:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	go subscribe(client)
	go publish(client)

	for s := range signals {
		if s == os.Interrupt || s == os.Kill || s == syscall.SIGTERM {
			break
		}
	}
	signal.Stop(signals)
}

func publish(client *redis.Client) {
	logger.Infow("publish start",
		"client", client,
	)

	for {
		msg, err := clipboard.ReadAll()
		//	logger.Infof("read from clipboard, msg: %v", msg)

		//fmt.Println("msg:", msg)
		if err != nil {
			continue
		}
		if msg != "" && !strings.EqualFold(msg, localClipboardContent) {

			err := client.Publish("channel0", msg).Err()
			if err != nil {
				panic(err)
			}
			localClipboardContent = msg
			logger.Infof("clipboard change, publish, local clipboard content: %v, new msg: %v", localClipboardContent, msg)
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func subscribe(client *redis.Client) {
	logger.Infow("subscribe start",
		"client", client,
	)

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
		remoteMsg := msg.Payload
		if !strings.EqualFold(remoteMsg, localClipboardContent) {
			clipboard.WriteAll(remoteMsg)
			logger.Infof("write to clipboard, local: %v, new: %v", localClipboardContent, remoteMsg)
		}
	}
}
