package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("BOOTSTRAP_SERVERS"),
		"group.id":          "gift-group",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	consumer.Subscribe("gift-events", nil)

	balance := map[int64]int64{
		1: 1000000, // на складе есть 1000000 подарков с ID=1
		2: 5000000, // на складе есть 5000000 подарков с ID=2

	}

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

		strVal := string(msg.Value) // например, "12345"
		itemID, err := strconv.Atoi(strVal)
		if err != nil {
			log.Printf("Ошибка преобразования '%s' в число: %v", strVal, err)
			continue
		}

		balance[int64(itemID)] = balance[int64(itemID)] - 1

		fmt.Printf("-------consumer-gifts-------\n")
		// распечатаем оставшийся баланс, в логах он должен постоянно убывать
		fmt.Printf("| itemID: %d | writeoff success. actual balance:\n", itemID)
		fmt.Printf("-------------------------------------------\n")
		fmt.Printf("| %-28s|%-10d%s\n", "gifts", balance[int64(itemID)], "|")
		fmt.Printf("-------------------------------------------\n")
	}
}
