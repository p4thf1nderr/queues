package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

type Coordinate struct {
	X int64
	Y int64
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

	sub, err := nc.SubscribeSync("courier.*")
	if err != nil {
		panic(err)
	}

	fmt.Println("start listening coordinates...")

	for {
		var coordinate Coordinate

		msg, err := sub.NextMsg(time.Minute)
		if err != nil {
			fmt.Printf("can't read next message from nats: %v", err)
			continue
		}

		fmt.Println("revcieved new event")

		err = json.Unmarshal(msg.Data, &coordinate)
		if err != nil {
			fmt.Printf("can't unmarshal message from nats: %v", err)
			continue
		}

		fmt.Printf("user recieved coordinate from courier: x:%d y:%d\n", coordinate.X, coordinate.Y)
	}

}
