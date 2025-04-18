// Package Publisher provides methods to connect, publish data from aggregator to some consumer
package publisher

import (
	"encoding/json"
	"fmt"
	"sync"

	m "Collector/internal/models"
	"Collector/internal/utils"

	"github.com/nats-io/nats.go"
)

type Message struct {
	Symbol    string        `json:"symbol"`
	Futures   m.FuturesData `json:"futures"`
	Spot      m.SpotData    `json:"spot"`
	Timestamp int64         `json:"timestamp"`
}
var (
	once sync.Once
	pub  *Publisher
)

// Create connection to the NATS. Singleton
func InitPublisher() {
	once.Do(func() {
		dotenv, err := utils.GetDotenv("NATS_URL")
		if err != nil {
			fmt.Errorf("load NATS URL from env: %w", err)
		}

		conn, err := nats.Connect(dotenv[0])
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
