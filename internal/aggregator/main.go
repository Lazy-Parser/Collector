// Package aggregator implements the core logic for aggregating real-time
// market data from various sources such as MEXC Spot, Futures, and others.
//
// It handles normalization, caching, deduplication, and comparison of prices.
// Once data from different sources for the same symbol is available, it
// publishes a unified message (e.g., spread) for external use.
//
// This package is designed to be thread-safe and run concurrently in the background.
// It uses channels and sync primitives to ensure safe access to shared data.

package aggregator

import (
	"log"
	"strings"
	"sync"
	"time"

	p "github.com/Lazy-Parser/Collector/internal/publisher"
)

var (
	once   sync.Once
	joiner *Joiner
	bus    chan AggregatorStruct
)

// InitJoiner initializes the singleton Joiner and NATS connection structures.
func InitJoiner() {
	once.Do(func() {
		joiner = &Joiner{
			cache: make(map[string]map[string]AggregatorStruct),
		}
		bus = make(chan AggregatorStruct, 5000)

		// init nats connect to publish joiners
		p.InitPublisher()
		// defer p.GetPublisher().Close()

		go joiner.run()
	})
}

func GetJoiner() *Joiner {
	return joiner
}

// run listens to the bus channel for incoming AggregatorStruct data.
func (j *Joiner) run() {
	log.Println("🧠 Aggregator started...")

	for data := range bus {
		j.Update(data)
	}
}

// add data from param to the bus. Data will be added to cache
func (j *Joiner) Push(data AggregatorStruct) {
	bus <- data
}

// Update cache and aggregate all data
func (j *Joiner) Update(data AggregatorStruct) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, ok := j.cache[data.Symbol]; !ok {
		j.cache[data.Symbol] = make(map[string]AggregatorStruct)
	}
	j.cache[data.Symbol][data.Type] = data

	futures := j.cache[data.Symbol]["FUTURES"]
	spot := j.cache[data.Symbol]["SPOT"]

	// send via nats
	if futures.Price != "" && spot.Price != "" {
		payload := &p.Message{
			Symbol:       data.Symbol,
			SpotPrice:    spot.Price,
			FuturesPrice: futures.Price,
			Timestamp:    max(futures.Timestamp, spot.Timestamp).UnixMilli(),
		}

		p.GetPublisher().Publish("mexc.spread", *payload)

		log.Printf("🔁 %s: FUTURES %s | SPOT %s",
			data.Symbol, futures.Price, spot.Price)
	}
}

// return the latest time
func max(t1 time.Time, t2 time.Time) time.Time {
	if t1.Compare(t2) == -1 { // t1 < t2
		return t2
	}

	return t1
}

// converts symbol string with underscores or without
// a separator to the "BTC/USDT" format by replacing "_" with "/" or
// appending "/" before "USDT" if no separators are found.
func NormalizeSymbol(symbol string) string {
	newString := strings.ReplaceAll(symbol, "_", "/")

	if newString == symbol {
		newString = strings.ReplaceAll(symbol, "USDT", "/USDT")
	}

	return newString
}
