// Package Publisher provides methods to connect, publish data from aggregator to some consumer
package publisher

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	m "github.com/Lazy-Parser/Collector/internal/models"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

type Message struct {
	Symbol    string        `json:"symbol"`
	Futures   m.FuturesData `json:"futures"`
	Spot      m.SpotData    `json:"spot"`
	Timestamp int64         `json:"timestamp"`
}

// type FuturesData struct {
// 	LastPrice    float64 `json:"price"`
// 	FairPrice    float64 `json:"fair_price"`
// 	IndexPrice   float64 `json:"index_price"`
// 	Amount24     float64 `json:"amount24"`
// 	FundingRate  float64 `json:"funding_rate"`
// 	RiseFallRate float64 `json:"rise_fall_rate"`
// 	Bid1         float64 `json:"bid1"`
// 	Ask1         float64 `json:"ask1"`
// }

var (
	once sync.Once
	pub  *Publisher
)

// Create connection to the NATS. Singleton
func InitPublisher() {
	once.Do(func() {
		natsURL, err := getDotenv()
		if err != nil {
			fmt.Errorf("load NATS URL from env: %w", err)
		}

		conn, err := nats.Connect(natsURL)
		if err != nil {
			fmt.Errorf("connect to NATS: %w", err)
		}
		fmt.Println("Connected to NATS ✅")

		pub = &Publisher{conn}
	})
}

// return nats connection
func GetPublisher() *Publisher {
	if pub == nil {
		fmt.Errorf("Publisher is nil. Call InitPublisher first!")
	}

	return pub
}

// publish (send) message by provided subject
func (p *Publisher) Publish(subject string, data Message) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal nats payload: %w", err)
	}

	return p.nc.Publish(subject, payload)
}

func (p *Publisher) Close() {
	p.nc.Close()
	fmt.Println("🛑 NATS connection closed")
}

func getDotenv() (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", fmt.Errorf("error trying to get .env var: %w", err)
	}

	natsURL := os.Getenv("NATS_URL")
	return natsURL, nil
}
