package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type Coordinate struct {
	X int64
	Y int64
}

type Order struct {
	Status string
}

func main() {
	// http сервер

	// 1. Subscribe
	// ручка, принимающая координату курьера
	// в хэндлере принимаем координату и отрпавляем ее логисту и юзеру, чтобы отобразить где курьер

	// 2. Request/Response
	// логист отправляет запрос к курьеру, с запросом, готов ли он взять новый заказ

	// 3. Queue группы.
	// События, взял заказ/доставил заказ - нужно отправить и логисту и юзеру

	server := gin.Default()

	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		panic(err)
	}

	defer nc.Drain()

	server.POST("coordinates", func(ctx *gin.Context) {
		var coordinate Coordinate

		if err := ctx.BindJSON(&coordinate); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "can't marshal object"})
			return
		}

		payload, err := json.Marshal(coordinate)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "can't marshal object"})
			return

		}

		err = nc.Publish("courier.coordinate", payload)
		if err != nil {
			panic(err)
		}

		ctx.JSON(http.StatusAccepted, gin.H{"message": "accepted"})
	})

	server.POST("order", func(ctx *gin.Context) {
		var order Order

		if err := ctx.BindJSON(&order); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "can't marshal object"})
			return
		}

		payload, err := json.Marshal(order)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "can't marshal object"})
			return

		}

		nc.Publish("order", payload)

		ctx.JSON(http.StatusAccepted, gin.H{"message": "accepted"})
	})

	server.Run("0.0.0.0:8080")
}
