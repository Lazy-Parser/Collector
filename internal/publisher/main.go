// Package Publisher provides methods to connect, publish data from aggregator to some consumer
package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

type Message struct {
	Symbol       string  `json:"symbol"`
	SpotPrice    string  `json:"spot_price"`
	FuturesPrice string  `json:"futures_price"`
	Timestamp    int64   `json:"timestamp"`
	Amount24     float64 `json:"amount24"`
}

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
		log.Println("Connected to NATS ✅")

		pub = &Publisher{conn}
	})
}

// return nats connection
func GetPublisher() *Publisher {
	if pub == nil {
		log.Fatalln("Publisher is nil. Call InitPublisher first!")
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
	log.Println("🛑 NATS connection closed")
}

func getDotenv() (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", fmt.Errorf("error trying to get .env var: %w", err)
	}

	natsURL := os.Getenv("NATS_URL")
	return natsURL, nil
}
