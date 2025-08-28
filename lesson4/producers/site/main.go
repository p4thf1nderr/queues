package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/queues/lesson4/producers/site/job"
	"github.com/queues/lesson4/producers/site/models"
	"github.com/robfig/cron/v3"
)

func main() {
	server := gin.Default()

	db, err := sqlx.Connect("postgres", "user=admin dbname=site password=qaz123 host=db-site port=5432 sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	startCron(db)

	server.POST("order", func(ctx *gin.Context) {
		var order models.Order

		if err := ctx.BindJSON(&order); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "can't marshal object"})
			return
		}

		tx := db.MustBegin()

		var orderID int64

		err := tx.QueryRow(`INSERT INTO orders (product_id) VALUES ($1) RETURNING id`, order.ProductID).Scan(&orderID)
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		event := models.Event{
			ID:        orderID,
			EventType: "output",
			ProductID: order.ProductID,
			Source:    "site",
			Timestamp: float64(time.Now().Unix()),
		}

		payload, err := json.Marshal(event)
		if err != nil {
			panic(err)
		}

		_, err = tx.NamedExec(`INSERT INTO outbox (payload, event_id) VALUES (:payload, :event_id)`,
			map[string]interface{}{
				"payload":  payload,
				"event_id": orderID,
			})
		if err != nil { // в случае ошибки записи в outbox откатываем транзакцию создания заказа
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		tx.Commit()

		ctx.JSON(http.StatusCreated, gin.H{"message": "success"})
	})

	server.Run("0.0.0.0:8080")
}

func startCron(db *sqlx.DB) {
	c := cron.New()
	_, err := c.AddJob("@every 5s", &job.OutboxJob{
		DB: db,
	})
	if err != nil {
		log.Fatalf("ошибка при добавлении задачи: %v", err)
	}
	c.Start()
}
