package producer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"usergenerator/internal"

	"github.com/segmentio/kafka-go"
)

func newKafkaWriter(cfg internal.Config) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaURL),
		Topic:                  cfg.KafkaProducerTopic,
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
	}
}

func Produce(ct *context.Context, cfg internal.Config, logger *slog.Logger, messages chan kafka.Message) {
	op := "kafka.producerFIOfailed"

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)

	ctx, cancel := context.WithCancel(*ct)

	// go routine for getting signals asynchronously
	go func() {
		sig := <-signals
		logger.Info(fmt.Sprintf("%s: Got signal: %s", op, sig))
		cancel()
	}()

	delayMs, _ := strconv.Atoi(cfg.KafkaCFG.KafkaDelayMS)

	w := newKafkaWriter(cfg)

	logger.Info(fmt.Sprintf("%s Producer configuration: %v , %v", op, w.Addr,w.Topic))

	//i := 1

	defer func() {
		err := w.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("%s:Error closing producer: %s", op, err))
			return
		}
		logger.Info(fmt.Sprintf("%s:Producer closed", op))
	}()

	for {
		for i:=0;i<10;i++{
			<-messages
		}
		temp := <-messages
		m := kafka.Message{
			Key:   temp.Key,
			Value: temp.Value,
		}

		err := w.WriteMessages(ctx, m)
		if err == nil {
			logger.Info(fmt.Sprintf("%s:Sent message: %s-%d [%d]: %s = %s\n", op, w.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value)))
		} else if err == context.Canceled {
			logger.Error(fmt.Sprintf("%s: Context canceled: %s", op, err))
			break
		} else {
			logger.Error(fmt.Sprintf("%s: Error sending message: %s", op, err))
		}
		//i++

		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
}
