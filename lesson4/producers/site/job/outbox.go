package job

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/jmoiron/sqlx"
	"github.com/queues/lesson4/producers/site/models"
)

type OutboxJob struct {
	DB *sqlx.DB
}

func (j *OutboxJob) Run() {
	var tasks []models.Message // todo: переименовать модель

	err := j.DB.Select(&tasks, "SELECT payload, event_id FROM outbox WHERE processed_at is NULL")
	if errors.Is(err, sql.ErrNoRows) {
		return
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("BOOTSTRAP_SERVERS"),
		"acks":              "1",
	})

	if err != nil {
		panic(err)
	}

	defer producer.Close()

	topic := "product-events" // топик для событий о выпуске продуктов
	deliveryChan := make(chan kafka.Event)

	for _, task := range tasks {
		var event models.Event
		err = json.Unmarshal(task.Payload, &event)
		if err != nil {
			panic(err)
		}

		districtID := rand.Intn(5) + 1 // случайный id района в который сделали заказ на доставку

		eventBytes, err := json.Marshal(event)
		if err != nil {
			panic(err)
		}

		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(fmt.Sprint(districtID)),
			Value: eventBytes,
		}

		err = producer.Produce(message, deliveryChan)
		if err != nil {
			fmt.Printf("failed to produce message: %v", err)
		}
		producer.Flush(0)

		e := <-deliveryChan
		m := e.(*kafka.Message)

		// Check for delivery errors
		if m.TopicPartition.Error != nil {
			fmt.Printf("producer: delivery failed: %s", m.TopicPartition.Error)
		} else {
			_, err = j.DB.NamedExec(`UPDATE outbox SET processed_at = NOW() WHERE event_id = :event_id`,
				map[string]interface{}{
					"event_id": task.EventID,
				})
			if err != nil { // в случае ошибки обновления записи в outbox запись отправится снова
				fmt.Printf("producer: update outbox failed: %v", err)
				return
			}

			fmt.Printf("producer: delivered message to %v, Event: %s\n", m.TopicPartition, string(eventBytes))
		}

		time.Sleep(300*time.Millisecond + time.Duration(rand.Intn(500))*time.Millisecond)

	}
}
