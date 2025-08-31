package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

type Order struct {
	Status string
}

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

	// топик: order, косньюмер группа: logist
	sub, err := nc.QueueSubscribeSync("order", "user")
	if err != nil {
		panic(err)
	}

	for {
		var order Order

		msg, err := sub.NextMsg(10 * time.Millisecond)
		if err != nil {
			fmt.Printf("can't read next message from nats: %v", err)
			continue
		}

		err = json.Unmarshal(msg.Data, &order)
		if err != nil {
			fmt.Printf("can't unmarshal message from nats: %v", err)
			continue
		}

		fmt.Printf("user recieved order status from courier: %s", order.Status)
	}
}
