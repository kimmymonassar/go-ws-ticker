package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func getEnvVariable(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Could not load .env file")
	}

	return os.Getenv(key)
}

type ResponseData struct {
	eventType string `json:"e"`
	eventTime int `json:"E"`
	symbol string `json:"s"`
	aggregateTradeId int `json:"a"`
	price string `json:"p"`
	quantity string `json:"q"`
	firstTradeId int `json:"f"`
	lastTradeId int `json:"l"`
	tradeTime int `json:"T"`
	isBuyerMarketMaker bool `json:"m"`
	ignore bool `json:"M"`
}

func parseJsonData(data []byte) {
	var responseData ResponseData
	json.Unmarshal(data, &responseData)
	fmt.Printf("%+v\n", string(data))
}

func wsClient() {
	var addr = flag.String("addr", "stream.binance.com:9443", "wss URL:PORT")
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/ws/btcusdt@aggTrade"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
			// parseJsonData(message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			fmt.Printf("new tick: " + t.String());
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
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

func main() {
	API_KEY := getEnvVariable("API_KEY")

	fmt.Printf(API_KEY + "\n")
	wsClient();
}
