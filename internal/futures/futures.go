package futures

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type TickerMessage struct {
	Symbol  string          `json:"symbol"`
	Data    json.RawMessage `json:"data"` // <-- временно как "сырая" часть
	Channel string          `json:"channel"`
	Ts      int64           `json:"ts"`
}

type TickerData struct {
	Symbol                  string    `json:"symbol"`
	LastPrice               float64   `json:"lastPrice"`
	RiseFallRate            float64   `json:"riseFallRate"`
	FairPrice               float64   `json:"fairPrice"`
	IndexPrice              float64   `json:"indexPrice"`
	Volume24                int64     `json:"volume24"`
	Amount24                float64   `json:"amount24"`
	MaxBidPrice             float64   `json:"maxBidPrice"`
	MinAskPrice             float64   `json:"minAskPrice"`
	Lower24Price            float64   `json:"lower24Price"`
	High24Price             float64   `json:"high24Price"`
	Timestamp               int64     `json:"timestamp"`
	Bid1                    float64   `json:"bid1"`
	Ask1                    float64   `json:"ask1"`
	HoldVol                 int64     `json:"holdVol"`
	RiseFallValue           float64   `json:"riseFallValue"`
	FundingRate             float64   `json:"fundingRate"`
	Zone                    string    `json:"zone"`
	RiseFallRates           []float64 `json:"riseFallRates"`
	RiseFallRatesOfTimezone []float64 `json:"riseFallRatesOfTimezone"`
}

func Run(ctx context.Context) error {

	// get ws string from dotenv
	mexWS, err := getDotenv()
	if err != nil {
		return fmt.Errorf("dotenv: %w", err)
	}

	// connect to ws
	conn, err := connect(mexWS)
	if err != nil {
		return fmt.Errorf("connection: %w", err)
	}

	// subscribe to futures
	subscribeToFutures(ctx, conn)

	return nil
}

func getDotenv() (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}

	mexcWS := os.Getenv("MEXC_FUTURES_WS")
	return mexcWS, nil
}

func connect(mexcWS string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(mexcWS, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to Mexc: %w", err)
	}
	log.Println("Connected to Mexc ✅")

	return conn, nil
}

func subscribeToFutures(ctx context.Context, conn *websocket.Conn) {
	subscribe := map[string]interface{}{
		"method": "sub.depth",
		"param": map[string]interface{}{
			"symbol": "BTC_USDT",
			"sdfsdf": "sdfsdf",
		},
	}

	if err := conn.WriteJSON(subscribe); err != nil {
		log.Fatal("❌ Subscription failed:", err)
	}
	log.Println("📡 Subscribed to FUTURES ticker")

	for {
		select {
		case <-ctx.Done():
			log.Println("Exiting...")
			unsubscribe(conn)
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Fatalf("Error Reading message: %w \n", err)
				continue
			}

			// var data TickerData;
			// if err := json.Unmarshal([]byte(msg.Data), &data); err != nil {
			// 	log.Printf("⚠️ Ошибка парсинга внутреннего JSON из data: %v", err);
			// 	continue;
			// }

			// fmt.Printf("Symbol: %s | Fair Price: %f | ASK1: %f | BID1: %f \n", data.Symbol, data.FairPrice, data.Ask1, data.Bid1);

			fmt.Println(string(msg));
			fmt.Println("\n")
		}
	}
}

func unsubscribe(conn *websocket.Conn) {
	unsubscribe := map[string]interface{}{
		"method": "unsub.tickers",
		"param":  map[string]interface{}{"symbol": "BTC_USDT"},
	}

	conn.WriteJSON(unsubscribe)
	conn.Close()
}

// bit, ask,
