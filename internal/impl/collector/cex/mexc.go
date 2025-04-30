package cex

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Lazy-Parser/Collector/internal/domain"
	"github.com/Lazy-Parser/Collector/internal/utils"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	pingTimeout = flag.Duration("pingTimeout", time.Second*10, "How long should this service wait to ping MEXC")
)

type MexcSource struct {
	conn *websocket.Conn
}

func (m *MexcSource) Name() string {
	return "MEXC"
}

func (m *MexcSource) Connect() error {
	godotenv.Load(".env")
	mexcFutures := os.Getenv("MEXC_FUTURES_WS")
	if len(mexcFutures) == 0 {
		log.Panic("Faild to load MEXC_FUTURES_WS from dotenv")
	}

	conn, _, err := websocket.DefaultDialer.Dial(mexcFutures, nil)
	if err != nil {
		return fmt.Errorf("connect to Mexc: %w", err)
	}
	fmt.Printf("Connected to Mexc ✅ \n")

	m.conn = conn

	return nil
}

func (m *MexcSource) Subscribe() error {
	payload := map[string]interface{}{
		"method": "sub.tickers",
		"param":  map[string]interface{}{},
	}

	return m.conn.WriteJSON(payload)
}

func (m *MexcSource) Run(ctx context.Context, push func(domain.AggregatorPayload), setState func(bool)) {
	go m.ping(ctx)

	for {
		select {
		case <-ctx.Done():
			m.stop()
			setState(false)
			return
		default:
			_, msg, err := m.conn.ReadMessage()
			if err != nil {
				fmt.Errorf("Error Reading message: %w \n", err)
				fmt.Println(string(msg))
				continue
			}

			payloads := m.handleMsg(msg)
			if len(payloads) == 0 {
				continue
			}

			for _, payload := range payloads {
				push(payload)
			}
		}
	}
}

// private methods
func (m *MexcSource) handleMsg(msg []byte) []domain.AggregatorPayload {
	var res []domain.AggregatorPayload
	var data domain.Tickers
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		fmt.Printf("⚠️ Ошибка парсинга внутреннего JSON из data: %v", err)
		fmt.Println(string(msg))
		return []domain.AggregatorPayload{}
	}

	for _, ticker := range data.Data {
		payload := domain.AggregatorPayload{
			Exchange:  m.Name(),
			Symbol:    utils.NormalizeSymbol(ticker.Symbol),
			Timestamp: time.Now(),
			Data:      ticker,
		}

		res = append(res, payload)
	}

	return res
}

func (m *MexcSource) ping(ctx context.Context) {
	ticker := time.NewTicker(*pingTimeout)
	defer ticker.Stop()

	payload := map[string]interface{}{"method": "ping"}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := m.conn.WriteJSON(payload)
			if err != nil {
				fmt.Errorf("ping error: %w", err)
			} else {
				fmt.Printf("Ping MEXC \n")
			}
		}
	}
}

func (m *MexcSource) stop() {
	payload := map[string]interface{}{
		"method": "sub.tickers",
		"param":  map[string]interface{}{},
	}

	fmt.Println("Stopping Mexc...")
	m.conn.WriteJSON(payload)
	m.conn.Close()
}
