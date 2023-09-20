package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/segmentio/kafka-go"
)


func getKafkaReader(cfg internal.Config) *kafka.Reader {
	brokers := strings.Split(cfg.KafkaURL, ",")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  "6",
		Topic:    cfg.KafkaConsumerTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		//StartOffset: 0,
	})
}


func Consumer(ct *context.Context,cfg internal.Config,logger *slog.Logger,ch chan kafka.Message){
	op:="kafka.consumerFIO"

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)

	ctx, cancel := context.WithCancel(*ct)

	// go routine for getting signals asynchronously
	go func() {
		sig := <-signals
		logger.Info(fmt.Sprintf("%s Got signal: %v", op,sig))
		cancel()
	}()





	r := getKafkaReader(cfg)

	logger.Info(fmt.Sprintf("%s Consumer configuration: %v",op,r.Config()))

	defer func() {
		err := r.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("%s: Error closing consumer: %s",op,err))
			return
		}
		logger.Info(fmt.Sprintf("%s Consumer closed: %v",op,r.Config()))
	}()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			logger.Error(fmt.Sprintf(" %s: Error reading message: %s",op,err))
			break
		}
		ch<-m
		logger.Info(fmt.Sprintf("%s:Received message from %s-%d [%d]: %s = %s\n", op, m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value)))
	}
}


