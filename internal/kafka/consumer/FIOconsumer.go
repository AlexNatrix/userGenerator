package consumer

import (
	"context"
	"encoding/json"
	"log"
	models "main/internal/lib/api/model/user"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	topic          = "message-log"
	broker1Address = "localhost:9092"
	broker2Address = "localhost:9093"
	broker3Address = "localhost:9095"
)




func Produce(ctx context.Context) {
	topic := "FIO"
	partition := 0
	
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9093", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	
	conn.SetWriteDeadline(time.Now().Add(10*time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Value: []byte(`{"name": "Dmitriy","surname": "Ushakov","patronymic": "Vasilevich"}`)},
		kafka.Message{Value: []byte(`{"name": "Dmitriy","surname": "Glushakov","patronymic": "Vasilevich"}`)},
		kafka.Message{Value: []byte(`{"name": "Valisiy","surname": "Ushakov"}`)},
		kafka.Message{Value: []byte(`{"name": "Vasiliy","surname": "Ushakov"}`)},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
	
	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
//ch chan internal.GenUser
func Consume(){
	startTime := time.Now().Add(-time.Minute)
	endTime := time.Now()
	batchSize := int(10e6) // 10MB
	
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9093"},
		Topic:     "FIO",
		Partition: 0,
		MaxBytes:  batchSize,
	})
	
	r.SetOffsetAt(context.Background(), startTime)
	
	for {
		m, err := r.ReadMessage(context.Background())
	
		if err != nil {
			break
		}
		if m.Time.After(endTime) {
			break
		}
		// TODO: process message
		var user models.BaseUser
		if err=json.Unmarshal(m.Value,&user);err!=nil{
			//logger.Info()
			//msg := kafka.Message{Value: m.Value}
		}else{
			//ch<-user
		}
	}
	
	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}