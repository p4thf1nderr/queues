package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Event struct {
	ID        int64  `json:"id"`
	EventType string `json:"event_type"`
	ProductID int64  `json:"product_id"`
	Source    string `json:"source"`
}

func main() {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("BOOTSTRAP_SERVERS"),
		"group.id":          "product-group",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	consumer.Subscribe("product-events", nil)

	// рецепт хранит значение, сколько ингридиентов нужно списать, чтобы
	// выпустить продукт
	recipies := map[int64]map[string]int{
		1: {
			"salt":  100,
			"meal":  200,
			"water": 100,
		},
		2: {
			"salt":  50,
			"meal":  100,
			"water": 100,
		},
		3: {
			"salt":  50,
			"meal":  100,
			"water": 100,
		},
	}

	// сколько ингредиентов есть на балансе у склада
	balance := map[string]int{
		"salt":  1000000,
		"meal":  1000000,
		"water": 1000000,
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

		var event Event
		err = json.Unmarshal(msg.Value, &event)
		if err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			continue
		}

		// списываем с баланса склада ингредиенты для выпуска продукта
		// находим рецепт чтобы узнать что списывать
		reciepie := recipies[event.ProductID]

		for name, quanity := range reciepie {
			balance[name] = balance[name] - quanity
		}

		fmt.Printf("-------consumer-warehouse-------\n")
		// распечатаем оставшийся баланс, в логах он должен постоянно убывать
		fmt.Printf("| eventID: %d | source: %s | writeoff success. actual balance:\n",
			event.ID, event.Source)
		fmt.Printf("--------------------------------\n")
		fmt.Printf("| %-28s|%-10d%s\n", "salt", balance["salt"], "|")
		fmt.Printf("| %-28s|%-10d%s\n", "meal", balance["meal"], "|")
		fmt.Printf("| %-28s|%-10d%s\n", "water", balance["water"], "|")
		fmt.Printf("--------------------------------\n")
	}
}
