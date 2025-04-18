// Package mexc provides a WebSocket client for consuming real-time
// market data from the MEXC exchange (Spot and Futures).
//
// It handles connection management, subscription, message reading,
// and parsing of MEXC-specific payloads.
//
// Parsed data is converted into unified PriceFeed structs and pushed
// into the aggregator for further processing.

package mexc

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	a "Collector/internal/aggregator"
	m "Collector/internal/models"

	// p "Collector/internal/publisher"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	pingTimeout = flag.Duration("pingTimeout", 10*time.Second, "Repeat PING signal every 20 seconds")
)

func Run(ctx context.Context) error {
	fmt.Println("Loaded dotenvs successful")
	mexcFutures, mexcSpot, err := getDotenv()
	if err != nil {
		return fmt.Errorf("Error getting data from env: %w", err)
	}

	// TODO: add PING
	futuresConf := MexcConf{
		Type: Futures,
		URL:  mexcFutures,
		Subscribe: map[string]interface{}{
			"method": "sub.tickers",
			"param":  map[string]interface{}{},
		},
		Unsubscribe: map[string]interface{}{
			"method": "unsub.tickers",
			"param":  map[string]interface{}{},
		},
		ParseFunc: parseFutures,
	}
	// TODO: add PING
	spotConf := MexcConf{
		Type: Spot,
		URL:  mexcSpot,
		Subscribe: map[string]interface{}{
			"method": "SUBSCRIPTION",
			"params": []string{"spot@public.miniTickers.v3.api@UTC+8"},
		},
		Unsubscribe: map[string]interface{}{
			"method": "UNSUBSCRIPTION",
			"params": []string{"spot@public.miniTickers.v3.api@UTC+8"},
		},
		ParseFunc: parseSpot,
	}

	a.InitJoiner()

	go runWS(ctx, futuresConf)
	go runWS(ctx, spotConf)

	<-ctx.Done()

	return nil
}

// runWS establishes a WebSocket connection and subscribes to MEXC data.
func runWS(ctx context.Context, conf MexcConf) error {
	conn, err := connect(conf.URL, conf.Type)
	if err != nil {
		return err
	}

	// ping every 20 seconds
	go ping(conf.Type, conn)

	subscribe(ctx, conn, conf)

	return nil
}

// connect to Mexc WebSocket and returns an established connection.
func connect(mexcWS string, tp ConfType) (*websocket.Conn, error) {
	fmt.Println("Try to connect...")
	conn, _, err := websocket.DefaultDialer.Dial(mexcWS, nil)
	fmt.Println("Connected")
	if err != nil {
		return nil, fmt.Errorf("connect to Mexc: %w", err)
	}
	fmt.Printf("Connected %s to Mexc ✅ \n", tp)

	return conn, nil
}

// Subscribe to Mexc WebSocket and parse incoming messages.
//
// The function writes a subscription message to the websocket connection and then
// listens for incoming messages. When a message is received, it is parsed using
// the ParseFunc provided in the MexcConf struct.
func subscribe(ctx context.Context, conn *websocket.Conn, conf MexcConf) {
	if err := conn.WriteJSON(conf.Subscribe); err != nil {
		fmt.Errorf("❌ Subscription failed:", err)
	}

	fmt.Printf("📡 Subscribed to %s ticker \n", conf.Type)

	for {
		select {
		case <-ctx.Done():
			unsubscribe(conn, conf)
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Errorf("Error Reading message: %w \n", err)
				fmt.Println(string(msg))
				continue
			}

			conf.ParseFunc(msg)
		}
	}
}

// parseFutures takes a JSON message (in bytes) and parses it into a slice of m.TickerData.
// It then creates an AggregatorStruct for each ticker and pushes it to the Joiner.
func parseFutures(msg []byte) {
	var data m.Tickers
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		fmt.Printf("⚠️ Ошибка парсинга внутреннего JSON из data: %v", err)
		fmt.Println(string(msg))
		return
	}

	for _, ticker := range data.Data {
		payload := a.AggregatorStruct{
			Exchange:  "MEXC",
			Type:      "FUTURES",
			Symbol:    a.NormalizeSymbol(ticker.Symbol),
			Futures:   ticker,
			Timestamp: time.Now(),
		}

		a.GetJoiner().Push(payload)
	}
}

// parseSpot takes a JSON message (in bytes) and parses it into a slice of SpotMiniTickers.
// It then creates an AggregatorStruct for each ticker and pushes it to the Joiner.
func parseSpot(msg []byte) {
	var data m.SpotMiniTickersResponse
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		// fmt.Printf("⚠️ Ошибка парсинга внутреннего JSON из data: %v", err)
		fmt.Println(string(msg))
		return
	}

	for _, ticker := range data.Data {
		payload := a.AggregatorStruct{
			Exchange:  "MEXC",
			Type:      "SPOT",
			Symbol:    a.NormalizeSymbol(ticker.Symbol),
			Spot:      m.TickerToSpotData(ticker), // convert string fields to float64
			Timestamp: time.Now(),
		}
		a.GetJoiner().Push(payload)
	}
}

func unsubscribe(conn *websocket.Conn, conf MexcConf) {
	fmt.Println("Exiting...")
	conn.WriteJSON(conf.Unsubscribe)
	conn.Close()
}

func ping(tp ConfType, conn *websocket.Conn) {
	ticker := time.NewTicker(*pingTimeout)
	defer ticker.Stop()

	var payload map[string]interface{}
	if tp == Futures {
		payload = map[string]interface{}{"method": "ping"}
	} else {
		payload = map[string]interface{}{"method": "PING"}
	}

	for {
		select {
		case <-ticker.C:
			err := conn.WriteJSON(payload)
			if err != nil {
				fmt.Errorf("send ping: %w", err)
			} else {
				fmt.Printf("Ping %s \n", tp)
			}
		}
	}
}

func getDotenv() (string, string, error) {
	if err := godotenv.Load(); err != nil {
		return "", "", err
	}

	mexcFutures := os.Getenv("MEXC_FUTURES_WS")
	mexcSpot := os.Getenv("MEXC_SPOT_WS")
	return mexcFutures, mexcSpot, nil
}
