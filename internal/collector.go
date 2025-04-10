package collector

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/Lazy-Parser/Collector/gen/ticker"
	// "github.com/Lazy-Parser/Collector/gen/wrapper"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	// "google.golang.org/protobuf/proto"
)

func Run(ctx context.Context) error {
	mexcWS, dexWS, err := loadDotenv()
	if err != nil {
		return fmt.Errorf("error retreiving dotenv vars: %w", err)
	}

	log.Printf("Dotenv vars success ✅")
	fmt.Println(mexcWS)

	conn, err := connect(ctx, mexcWS, dexWS)
	if err != nil {
		return fmt.Errorf("error connection: %w", err)
	}
	defer conn.Close()

	subsribeToMexcStream(ctx, conn)

	return nil
}

func loadDotenv() (string, string, error) {
	if err := godotenv.Load(); err != nil {
		return "", "", err
	}

	mexcWS := os.Getenv("MEXC_WS")
	dexWS := os.Getenv("DEX_WS")
	return mexcWS, dexWS, nil
}

func connect(ctx context.Context, mexcWS string, dexWS string) (*websocket.Conn, error) {
	// connect to mexc
	conn, _, err := websocket.DefaultDialer.Dial(mexcWS, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to Mexc: %w", err)
	}
	log.Println("Connected to Mexc ✅")

	return conn, nil
}

func subsribeToMexcStream(ctx context.Context, conn *websocket.Conn) {
	subscribe := map[string]interface{}{
		"method": "SUBSCRIPTION",
		"params": []string{"spot@public.miniTickers.v3.api@UTC+8"},
	}
	//   spot@public.miniTickers.v3.api@BTCUSDT@UTC+8

	if err := conn.WriteJSON(subscribe); err != nil {
		log.Fatal("❌ Subscription failed:", err)
	}

	log.Println("📡 Subscribed to ETHUSDT miniTicker")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("❌ Read error:", err)
			return
		}
		log.Println("New Mesasge!")

		// Пропускаем JSON
		fmt.Println(string(msg))
		fmt.Println("\n")
		// if len(msg) > 0 && msg[0] == '{' {
		// 	log.Println("🔁 JSON control message:", string(msg))
		// 	continue
		// }

		// var wrapperMsg wrapper.PushDataV3ApiWrapper
		// if err := proto.Unmarshal(msg, &wrapperMsg); err != nil {
		// 	log.Println("❌ Unmarshal wrapper error:", err)
		// 	continue
		// }

		// tickerMsg, ok := wrapperMsg.Body.(*wrapper.PushDataV3ApiWrapper_PublicMiniTicker)
		// if !ok {
		// 	log.Fatalln("Error while trying to cast message to PublicMiniTicker")
		// 	continue
		// }

		// eth := tickerMsg.PublicMiniTicker

		// fmt.Printf("PRICE: %s | VOLUME: %s | RATE: %s \n", eth.Price, eth.Volume, eth.Rate)
		// print(eth)
	}
}

func print(t *ticker.PublicMiniTickerV3Api) {
	value := reflect.ValueOf(t)
	key := reflect.TypeOf(t)

	fmt.Println("-------------------------")
	for index := range key.NumField() {
		if !value.Field(index).CanInterface() {
			continue
		}

		fmt.Printf("%s: %v\n", key.Field(index).Name, value.Field(index).Interface())
	}
	fmt.Println("-------------------------")
}
