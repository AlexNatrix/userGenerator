package kafka_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"usergenerator/internal"
	models "usergenerator/internal/lib/api/model/user"

	"github.com/brianvoe/gofakeit"
	"github.com/segmentio/kafka-go"
)


func FakeUserGenerator(n int) []models.User {
	ret := make([]models.User, n)
	for i := 0; i < n; i++ {
		u := gofakeit.Person()
		user := models.NewUser()
		user.Name = u.FirstName
		user.Surname = u.LastName
		user.Patronymic = "Sanich"
		user.Sex = u.Gender
		user.Nationality = u.Address.Country
		user.Age = rand.Intn(100)
		ret[i] = user
	}
	return ret
}

func FakeBaseUserGenerator(n int) []models.BaseUser {
	ret := make([]models.BaseUser, n)
	for i := 0; i < n; i++ {
		des := rand.Intn(100)
		u := gofakeit.Person()
		user := models.BaseUser{Name:u.FirstName,Surname:u.LastName}
		if des < 75 {
			user.Patronymic = "Sanich"
		}
		if des<3{
			user.Name=""
		}
		ret[i] = user
	}
	return ret
}

func newKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
}

func Populate(n int,logger *slog.Logger, cfg internal.Config) {
	op:="kafka.test.Populate"
	kafkaURL := cfg.KafkaCFG.KafkaURL
	topic := cfg.KafkaCFG.KafkaConsumerTopic
	writer := newKafkaWriter(kafkaURL, topic)
	logger.Info(fmt.Sprintf("%s start producing ... !!",op))
	data:=FakeBaseUserGenerator(n)
	for i := 0; i<n; i++ {
		key := fmt.Sprintf("Key-%d", i)

		val,err:=json.Marshal(data[i])

		if err!=nil{
			logger.Error(fmt.Errorf("%s: %w",op,err).Error())
		}
		msg := kafka.Message{
			Key:   []byte(key),
			Value: val,
		}
		err = writer.WriteMessages(context.Background(), msg)
		if err != nil {
			logger.Error("%s :%w",op,err)
		} else {
			logger.Info(fmt.Sprintf("%s produced %s", op, string(val)))
		}
		//time.Sleep(1 * time.Millisecond)
	}
	writer.Close()
}