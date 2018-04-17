package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/blockcdn-go/exchange-sdk-go/okex"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	go wss(ctx, "1")
	go wss(ctx, "2")

	for {
		select {
		case <-interrupt:
			cancel()
			return
		}
	}
}

func wss(ctx context.Context, id string) {
	wss := okex.NewWSSClient(nil)
	msgCh, err := wss.QuerySpot()
	if err != nil {
		log.Fatal("query error: ", err)
	}

	for {
		select {
		case m := <-msgCh:
			fmt.Printf("ID: %s, message: %s\n", id, string(m))
		case <-ctx.Done():
			wss.Close()
			return
		}
	}
}
