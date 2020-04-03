package main

import (
	"github.com/atotto/clipboard"
	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	var cores []zapcore.Core

	level := zapcore.InfoLevel
	writer := zapcore.Lock(os.Stdout)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	core := zapcore.NewCore(encoder, writer, level)
	cores = append(cores, core)
	combinedCore := zapcore.NewTee(cores...)
	logger = zap.New(combinedCore,
		zap.AddCallerSkip(2),
		zap.AddCaller(),
	).Sugar()

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
		logger.Debug("monitoring clipboard")
		msg, err := clipboard.ReadAll()
		logger.Debugf("read from clipboard, msg: %v", msg)

		if err == nil {
			if msg != "" && !strings.EqualFold(msg, localClipboardContent) {
				err := client.Publish("channel0", msg).Err()
				if err != nil {
					panic(err)
				}
				localClipboardContent = msg
				logger.Infof("clipboard change, publish, local clipboard content: %v, new msg: %v", localClipboardContent, msg)
			}
		} else {
			logger.Errorf("failed to read clipboard: %v", err)
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
