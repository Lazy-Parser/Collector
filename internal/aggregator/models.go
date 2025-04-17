package aggregator

import (
	"sync"
	"time"

	m "Collector/internal/models"
)

type AggregatorStruct struct {
	Exchange  string // "MEXC", "MCX", "DexScreener"
	Type      string // "SPOT", "FUTURES", "DEX"
	Symbol    string // например: "BTC/USDT"
	Spot      m.SpotData
	Futures   m.FuturesData
	Timestamp time.Time // время обновления
}

type Joiner struct {
	cache map[string]map[string]AggregatorStruct // Symbol -> Exchange -> Feed
	mu    sync.RWMutex
}
