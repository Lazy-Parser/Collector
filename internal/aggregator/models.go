package aggregator

import (
	"sync"
	"time"
)

type AggregatorStruct struct {
	Exchange  string    // "MEXC", "MCX", "DexScreener"
	Type      string    // "SPOT", "FUTURES", "DEX"
	Symbol    string    // например: "BTC/USDT"
	Price     string    // цена как строка
	Timestamp time.Time // время обновления
}

type Joiner struct {
	cache map[string]map[string]AggregatorStruct // Symbol -> Exchange -> Feed
	mu    sync.RWMutex
}
