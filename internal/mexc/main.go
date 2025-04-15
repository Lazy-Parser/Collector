package mexc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	a "github.com/Lazy-Parser/Collector/internal/aggregator"
	m "github.com/Lazy-Parser/Collector/internal/models"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
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

func runWS(ctx context.Context, conf MexcConf) error {
	conn, err := connect(conf.URL, conf.Type)
	if err != nil {
		return err
	}

	subscribe(ctx, conn, conf)

	return nil
}

func connect(mexcWS string, tp ConfType) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(mexcWS, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to Mexc: %w", err)
	}
	log.Printf("Connected %s to Mexc ✅ \n", tp)

	return conn, nil
}

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

func parseFutures(msg []byte) {
	var data m.Tickers
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		log.Printf("⚠️ Ошибка парсинга внутреннего JSON из data: %v", err)
		return
	}

	for _, ticker := range data.Data {
		payload := a.AggregatorStruct{
			Exchange:  "MEXC",
			Type:      "FUTURES",
			Symbol:    a.NormalizeSymbol(ticker.Symbol),
			Price:     strconv.FormatFloat(float64(ticker.FairPrice), 'f', -1, 32),
			Timestamp: time.Now(),
		}
		a.GetJoiner().Push(payload)
	}
}

func parseSpot(msg []byte) {
	var data SpotMiniTickersResponse
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		log.Printf("⚠️ Ошибка парсинга внутреннего JSON из data: %v", err)
		return
	}

	for _, ticker := range data.Data {
		payload := a.AggregatorStruct{
			Exchange:  "MEXC",
			Type:      "SPOT",
			Symbol:    a.NormalizeSymbol(ticker.Symbol),
			Price:     ticker.Price,
			Timestamp: time.Now(),
		}
		a.GetJoiner().Push(payload)
	}
}

func unsubscribe(conn *websocket.Conn, conf MexcConf) {
	log.Println("Exiting...")
	conn.WriteJSON(conf.Unsubscribe)
	conn.Close()
}

func getDotenv() (string, string, error) {
	if err := godotenv.Load(); err != nil {
		return "", "", err
	}

	mexcFutures := os.Getenv("MEXC_FUTURES_WS")
	mexcSpot := os.Getenv("MEXC_SPOT_WS")
	return mexcFutures, mexcSpot, nil
}
