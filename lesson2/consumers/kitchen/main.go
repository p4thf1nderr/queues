package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Event struct {
	ID        int    `json:"id"`
	EventType string `json:"event_type"`
	ProductID int64  `json:"product_id"`
	Source    string `json:"source"`
}

func main() {

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("BOOTSTRAP_SERVERS"),
		"group.id":          "kitchen-group",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	consumer.Subscribe("product-events", nil)

	// кухня только готовит продукт

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			if kafkaError, ok := err.(kafka.Error); ok && kafkaError.Code() == kafka.ErrPartitionEOF {
				fmt.Printf("End of partition reached %v\n", msg)
			} else {
				fmt.Printf("Error consuming message: %v\n", err)
			}
			continue
		}

		var event Event
		err = json.Unmarshal(msg.Value, &event)
		if err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			continue
		}

		fmt.Printf("-------consumer-kitchen-------\n")
		fmt.Printf("| eventID: %d | source: %s | productID: %d | is processing\n",
			event.ID, event.Source, event.ProductID)
	}
}
