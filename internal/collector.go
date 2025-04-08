package collector

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Lazy-Parser/Collector/gen/wrapper"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"google.golang.org/protobuf/proto"
)

func Run(ctx context.Context) error {
	mexcWS, dexWS, err := loadDotenv()
	if err != nil {
		return fmt.Errorf("Error retreiving dotenv vars: %w", err)
	}
	log.Printf("Dotenv vars success ✅")
	fmt.Println(mexcWS)

	conn, err := connect(ctx, mexcWS, dexWS)
	if err != nil {
		fmt.Errorf("error connection: %w", err)
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
	// connect nats
	// nc, err := nats.Connect(mexcWS);
	// if err != nil {
	// 	return nil, fmt.Errorf("connect to NATS Mexc: %w", err);
	// }
	// log.Println("Connected to Mexc ✅");

	// // TODO: connect to dex. Remove this string
	// fmt.Println("%s | %s", mexcWS, dexWS);

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
		"params": []string{ "spot@public.deals.v3.api.pb@BTCUSDT" },
	}

	if err := conn.WriteJSON(subscribe); err != nil {
		log.Fatal("❌ Subscription failed:", err)
	}

	log.Println("📡 Subscribed to ETHUSDT miniTicker")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("❌ Read error:", err)
			return;
		}

		// str := string(msg);
		var wrapperMsg wrapper.PushDataV3ApiWrapper

		// if str[0] == '{' {
		// 	log.Println(str);
		// 	continue;
		// } 
		if err := proto.Unmarshal(msg, &wrapperMsg); err != nil {
			log.Println("❌ Unmarshal wrapper error:", err)
			continue
		}

		tickerMsg, ok := wrapperMsg.Body.(*wrapper.PushDataV3ApiWrapper_PublicDeals)
		if !ok {
			continue
		}

		eth := tickerMsg.PublicDeals;

		fmt.Printf("Type: %s\n", eth.EventType);
		for index, deal := range eth.Deals {
			fmt.Printf("%d) Price: %s | Type: %s \n", index + 1, deal.Price, deal.TradeType);
		}
	}
}
