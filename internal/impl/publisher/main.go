// Package Publisher provides methods to connect, publish data from aggregator to some consumer
package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/ui"
	"github.com/Lazy-Parser/Collector/internal/utils"
	"github.com/nats-io/nats.go"
)

var (
	once sync.Once
	pub  *Publisher
)

// Create connection to the NATS. Singleton
func Init() {
	once.Do(func() {
		natsUrl, err := utils.LoadEnv("NATS_URL")
		if err != nil {
			log.Panic("Failed to load NATS_URL dotenv var!")
		}

		conn, err := nats.Connect(natsUrl)
		if err != nil {
			ui.GetUI().LogsView(fmt.Sprintf("connect to NATS: %w", err), "error")
		}
		ui.GetUI().LogsView("Connected to NATS ✅", "log")

		pub = &Publisher{conn}
	})
}

// return nats connection
func GetPublisher() *Publisher {
	if pub == nil {
		ui.GetUI().LogsView("Publisher is nil. Call InitPublisher first!", "error")
	}

	return pub
}

// publish (send) message by provided subject. Call in goroutine!
func (p *Publisher) PublishStreamDex(dataFlow chan core.CollectorDexResponse) error {
	subject := "price.dex"

	for msg := range dataFlow {
		payload, _ := json.Marshal(msg)
		if err := p.nc.Publish(subject, payload); err != nil {
			return err
		}
	}

	return nil
}

func (p *Publisher) PublishStreamCex(dataFlow chan core.MexcResponse) error {
	subject := "price.cex.mexc"

	for msg := range dataFlow {
		payload, _ := json.Marshal(msg)
		if err := p.nc.Publish(subject, payload); err != nil {
			return err
		}
	}

	return nil
}

func (p *Publisher) Close() {
	p.nc.Close()
	fmt.Println("🛑 NATS connection closed")
}
