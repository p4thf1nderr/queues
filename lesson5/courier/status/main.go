package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		panic(err)
	}

	defer nc.Drain()

	sub, err := nc.SubscribeSync("courier.status")
	if err != nil {
		panic(err)
	}

	for {
		msg, err := sub.NextMsg(time.Minute)
		if err != nil {
			panic(err)
		}

		msg.Respond([]byte(randStatus()))
		time.Sleep(time.Millisecond * 500)
	}
}

func randStatus() string {
	if rand.Intn(2) == 1 {
		return "is ready"
	}
	return "is busy"
}
