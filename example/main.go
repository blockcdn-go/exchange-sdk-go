package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

type event struct {
	Event      string    `json:"event"`
	Parameters parameter `json:"parameters"`
}

type parameter struct {
	Base    string `json:"base"`
	Binary  string `json:"binary"`
	Product string `json:"product"`
	Quote   string `json:"quote"`
	Type    string `json:"type"`
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	u := url.URL{Scheme: "wss", Host: "okexcomreal.bafang.com:10441", Path: "/websocket"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			typ, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read: ", err)
				return
			}

			log.Printf("revc: %s; type: %d\n", message, typ)
		}
	}()

	e := event{
		Event: "addChannel",
		Parameters: parameter{
			Base:    "okb",
			Binary:  "0",
			Product: "spot",
			Quote:   "btc",
			Type:    "ticker",
		},
	}
	err = c.WriteJSON(e)
	if err != nil {
		log.Fatal("write json: ", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("{'event':'ping'}"))
			if err != nil {
				log.Println("write: ", err)
				return
			}
		case <-interrupt:
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close: ", err)
				return
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
