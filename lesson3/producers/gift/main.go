package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("BOOTSTRAP_SERVERS"),
		"acks":              "1",
	})

	if err != nil {
		panic(err)
	}

	defer producer.Close()

	// топик для событий какой подарок выдавать при покупке определенной продукции, из него заполняется KTable
	topic := "gift-table-events"
	deliveryChan := make(chan kafka.Event)

	for {
		productID := rand.Intn(3) + 1 // id продукта
		giftID := rand.Intn(2) + 1    // id подарка
		eventBytes, _ := json.Marshal(giftID)

		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(fmt.Sprint(productID)),
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

		time.Sleep(300*time.Millisecond + time.Duration(rand.Intn(500))*time.Millisecond)
	}
}
