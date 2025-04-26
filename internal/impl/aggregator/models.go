package aggregator

import (
	"sync"
	"time"

	"github.com/Lazy-Parser/Collector/internal/domain"
)

type AggregatorStruct struct {
	Exchange  string // "MEXC", "MCX", "DexScreener"
	Type      string // "SPOT", "FUTURES", "DEX"
	Symbol    string // например: "BTC/USDT"
	Spot      domain.SpotData
	Futures   domain.FuturesData
	Timestamp time.Time // время обновления
}

type Joiner struct {
	cache map[string]map[string]domain.AggregatorPayload // Symbol -> Exchange -> Feed
	mu    sync.RWMutex
}
