package main

import (
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

	wss := okex.NewWSSClient(nil)
	msgCh, err := wss.QuerySpot()
	if err != nil {
		log.Fatal("query error: ", err)
	}

	for {
		select {
		case <-interrupt:
			wss.Close()
			return
		case m := <-msgCh:
			fmt.Println(string(m))
		}
	}
}
