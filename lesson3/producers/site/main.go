package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Event struct {
	ID        int     `json:"id"`
	EventType string  `json:"event_type"`
	ProductID int     `json:"product_id"`
	Source    string  `json:"source"`
	Timestamp float64 `json:"timestamp"`
}

func main() {

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

	id := 1
	for {
		districtID := rand.Intn(5) + 1 // случайный id района в который сделали заказ на доставку
		event := Event{
			ID:        id,
			EventType: "output",
			ProductID: rand.Intn(3) + 1, // случайный ID продукта
			Source:    "site",
			Timestamp: float64(time.Now().Unix()),
		}

		eventBytes, _ := json.Marshal(event)

		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(fmt.Sprint(districtID)),
			Value: eventBytes,
		}

		err := producer.Produce(message, deliveryChan)
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
			fmt.Printf("producer: delivered message to %v, Event: %s\n", m.TopicPartition, string(eventBytes))
		}

		id++
		time.Sleep(300*time.Millisecond + time.Duration(rand.Intn(500))*time.Millisecond)
	}
}
