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
	Amount24  float64   // обьем за 24 часа
	Timestamp time.Time // время обновления
}

type Joiner struct {
	cache map[string]map[string]AggregatorStruct // Symbol -> Exchange -> Feed
	mu    sync.RWMutex
}
