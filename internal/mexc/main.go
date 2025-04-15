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
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	a "github.com/Lazy-Parser/Collector/internal/aggregator"
	m "github.com/Lazy-Parser/Collector/internal/models"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	pingTimeout = flag.Duration("pingTimeout", 20*time.Second, "Repeat PING signal every 20 seconds")
)

func Run(ctx context.Context) error {
	mexcFutures, mexcSpot, err := getDotenv()
	if err != nil {
		return nil
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
	conn, _, err := websocket.DefaultDialer.Dial(mexcWS, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to Mexc: %w", err)
	}
	log.Printf("Connected %s to Mexc ✅ \n", tp)

	return conn, nil
}

// Subscribe to Mexc WebSocket and parse incoming messages.
//
// The function writes a subscription message to the websocket connection and then
// listens for incoming messages. When a message is received, it is parsed using
// the ParseFunc provided in the MexcConf struct.
func subscribe(ctx context.Context, conn *websocket.Conn, conf MexcConf) {
	if err := conn.WriteJSON(conf.Subscribe); err != nil {
		log.Fatal("❌ Subscription failed:", err)
	}

	log.Printf("📡 Subscribed to %s ticker \n", conf.Type)

	for {
		select {
		case <-ctx.Done():
			unsubscribe(conn, conf)
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Fatalf("Error Reading message: %w \n", err)
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
		log.Printf("⚠️ Ошибка парсинга внутреннего JSON из data: %v", err)
		fmt.Println(string(msg))
		return
	}

	for _, ticker := range data.Data {
		payload := a.AggregatorStruct{
			Exchange:  "MEXC",
			Type:      "FUTURES",
			Symbol:    a.NormalizeSymbol(ticker.Symbol),
			Price:     strconv.FormatFloat(float64(ticker.FairPrice), 'f', -1, 32),
			Amount24:  ticker.Amount24,
			Timestamp: time.Now(),
		}
		a.GetJoiner().Push(payload)
	}
}

func print(s interface{}) {
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	// Если передан указатель — разворачиваем
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		fmt.Println("Not a struct")
		return
	}

	fmt.Println("🔍 Struct fields:")
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		fmt.Printf(" - %s: %v\n", field.Name, value.Interface())
	}
}

// parseSpot takes a JSON message (in bytes) and parses it into a slice of SpotMiniTickers.
// It then creates an AggregatorStruct for each ticker and pushes it to the Joiner.
func parseSpot(msg []byte) {
	var data SpotMiniTickersResponse
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		// log.Printf("⚠️ Ошибка парсинга внутреннего JSON из data: %v", err)
		fmt.Println(string(msg))
		return
	}

	// for _, ticker := range data.Data {
	// 	payload := a.AggregatorStruct{
	// 		Exchange: "MEXC",
	// 		Type:     "SPOT",
	// 		Symbol:   a.NormalizeSymbol(ticker.Symbol),
	// 		// Volume:	   ticker.VolumeUSDT,
	// 		Price:     ticker.Price,
	// 		Timestamp: time.Now(),
	// 	}
	// 	a.GetJoiner().Push(payload)
	// }
}

func unsubscribe(conn *websocket.Conn, conf MexcConf) {
	log.Println("Exiting...")
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
				log.Println("Ping %s", tp)
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
