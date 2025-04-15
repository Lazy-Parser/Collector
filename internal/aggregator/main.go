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
)

var (
	once   sync.Once
	joiner *Joiner
	bus    chan AggregatorStruct
)

// InitJoiner initializes the singleton Joiner and its data structures.
func InitJoiner() {
	once.Do(func() {
		joiner = &Joiner{
			cache: make(map[string]map[string]AggregatorStruct),
		}
		bus = make(chan AggregatorStruct, 5000) // буферизированный канал
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

	// Пример: логика сравнения цен с разных бирж
	futures := j.cache[data.Symbol]["FUTURES"]
	spot := j.cache[data.Symbol]["SPOT"]

	// Если все три есть, покажем расхождение
	if futures.Price != "" && spot.Price != "" {
		log.Printf("🔁 %s: FUTURES %s | SPOT %s",
			data.Symbol, futures.Price, spot.Price)
	}
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

