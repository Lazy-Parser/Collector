package aggregator

import (
	"sync"
	"time"

	"github.com/Lazy-Parser/Collector/internal/core"
)

type AggregatorStruct struct {
	Exchange  string // "MEXC", "MCX", "DexScreener"
	Type      string // "SPOT", "FUTURES", "DEX"
	Symbol    string // например: "BTC/USDT"
	Spot      core.SpotData
	Futures   core.FuturesData
	Timestamp time.Time // время обновления
}

type Joiner struct {
	cache map[string]map[string]core.AggregatorPayload // Symbol -> Exchange -> Feed
	mu    sync.RWMutex
}
