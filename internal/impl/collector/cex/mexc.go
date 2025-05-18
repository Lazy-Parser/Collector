package cex

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/utils"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	pingTimeout = flag.Duration("pingTimeout", time.Second*10, "How long should this service wait to ping MEXC")
	state       chan bool // true - working / false - not working / stop / error
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

	state = make(chan bool, 1)
	err := m.conn.WriteJSON(payload)
	if err != nil {
		state <- false
		return err
	}

	state <- true
	return nil
}

func (m *MexcSource) Run(ctx context.Context, push func(core.AggregatorPayload)) {
	go m.ping(ctx)

	for {
		select {
		case <-ctx.Done():
			m.stop()
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
func (m *MexcSource) handleMsg(msg []byte) []core.AggregatorPayload {
	var res []core.AggregatorPayload
	var data core.Tickers
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		fmt.Printf("⚠️ Ошибка парсинга внутреннего JSON из data: %v", err)
		fmt.Println(string(msg))
		return []core.AggregatorPayload{}
	}

	for _, ticker := range data.Data {
		payload := core.AggregatorPayload{
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
		"method": "unsub.tickers",
		"param":  map[string]interface{}{},
	}

	state <- false
	fmt.Println("Stopping Mexc...")
	m.conn.WriteJSON(payload)
	m.conn.Close()
}

func (m *MexcSource) ListenState() <-chan bool {
	return state
}

func (m *MexcSource) SetState(value bool) {
	state <- value
}
