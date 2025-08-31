package main

import (
	"fmt"
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

	for {
		rep, err := nc.Request("courier.status", nil, time.Second)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("courier status: ", string(rep.Data))
		time.Sleep(time.Millisecond * 500)
	}
}
