// Package Publisher provides methods to connect, publish data from aggregator to some consumer
package publisher

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/Lazy-Parser/Collector/internal/domain"

	"github.com/nats-io/nats.go"
)

type Message struct {
	Symbol    string             `json:"symbol"`
	Futures   domain.FuturesData `json:"futures"`
	Spot      domain.SpotData    `json:"spot"`
	Timestamp int64              `json:"timestamp"`
}

var (
	once sync.Once
	pub  *Publisher
)

// Create connection to the NATS. Singleton
func InitPublisher() {
	once.Do(func() {
		natsUrl := os.Getenv("NATS_URL")

		conn, err := nats.Connect(natsUrl)
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
