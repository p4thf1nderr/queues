package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Event struct {
	ID        int64  `json:"id"`
	EventType string `json:"event_type"`
	ProductID int64  `json:"product_id"`
	Source    string `json:"source"`
}

func main() {
	// нужна таблица с балансом в бд
	// таблица с транзакциями, где будет event_id, product_id

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  os.Getenv("BOOTSTRAP_SERVERS"),
		"group.id":           "product-group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
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

	db, err := sqlx.Connect("postgres", "user=admin dbname=warehouse password=qaz123 host=db-warehouse port=5432 sslmode=disable")
	if err != nil {
		log.Fatalln(err)
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

		err = db.QueryRow("SELECT 1 FROM transactions WHERE transaction_id = $1", event.ID).Scan()
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			fmt.Println("inbox error:", err)
			return
		}

		if err == nil {
			consumer.CommitMessage(msg)
			fmt.Printf("duplicated message: event %d is already processed", event.ID)
			return
		}

		// списываем с баланса склада ингредиенты для выпуска продукта
		// находим рецепт чтобы узнать что списывать
		reciepie := recipies[event.ProductID]

		tx := db.MustBegin()

		for name, quantity := range reciepie {
			_, err = tx.NamedExec(`UPDATE balance SET amount = amount - :quantity WHERE ingredient_name = :ingredient_name`,
				map[string]interface{}{
					"ingredient_name": name,
					"quantity":        quantity,
				})
			if err != nil {
				tx.Rollback()
				fmt.Printf("warehouse: update balance failed: %v", err)
				return
			}
		}

		// таблица transactions сохраняет обработанные события заказов (транзакций)
		// и служит для проверки идемпотентности
		_, err = tx.NamedExec(`INSERT INTO transactions (transaction_id, transaction_type) VALUES (:transaction_id, :transaction_type)`,
			map[string]interface{}{
				"transaction_id":   event.ID,
				"transaction_type": "writeoff",
			})
		if err != nil { // в случае ошибки откатываем изменения баланса
			tx.Rollback()
			fmt.Printf("warehouse: update balance failed: %v", err)
			return
		}

		// ошибок нет, коммитим оффсет обновляем баланс и фиксируем, что транзакция была обработана
		tx.Commit()
		consumer.CommitMessage(msg)

		fmt.Printf("-------consumer-warehouse-------\n")
		fmt.Println("balance updated")
		fmt.Printf("--------------------------------\n")
	}
}
