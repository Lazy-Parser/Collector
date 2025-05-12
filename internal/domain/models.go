package domain

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Payload for Joiner. When some message parsed, we pass AggregatorPayload to Push method, and Joiner will process this payload and publish needed data.
// [Exchange, Symbol, Timestamp] - fields, that Joiner need for it work. Data - struct, that Joiner will send to Nats
type AggregatorPayload struct {
	Exchange  string      // "MEXC", "OKX", "DEX" ...
	Symbol    string      // "BTC/USDT" ...
	Timestamp time.Time   // time when data getted
	Data      interface{} // main data, that will be sent via NATS.
}

// maybe add in future
// Type  string  - "Futures" / "SPOT"

type FuturesData struct {
	Symbol       string  `json:"symbol"`
	LastPrice    float64 `json:"lastPrice"`
	RiseFallRate float64 `json:"riseFallRate"`
	FairPrice    float64 `json:"fairPrice"`
	IndexPrice   float64 `json:"indexPrice"`
	Volume24     float64 `json:"volume24"`
	Amount24     float64 `json:"amount24"`
	MaxBidPrice  float64 `json:"maxBidPrice"`
	MinAskPrice  float64 `json:"minAskPrice"`
	Lower24Price float64 `json:"lower24Price"`
	High24Price  float64 `json:"high24Price"`
	Timestamp    int64   `json:"timestamp"`
}

type Tickers struct {
	Data []FuturesData `json:data`
}

type SpotData struct {
	Symbol     string  `json:"s"` // Kept as string
	Price      float64 `json:"p"`
	Change     float64 `json:"r"`
	TrueChange float64 `json:"tr"`
	High       float64 `json:"h"`
	Low        float64 `json:"l"`
	VolumeUSDT float64 `json:"v"`
	VolumeBase float64 `json:"q"`
	LastRT     float64 `json:"lastRT"`
	MT         float64 `json:"MT"`
	NV         string  `json:"NV"`
}

// custom type, that combine DS (DexScreenr) responce and others. For global usage
type Pair struct {
	Base        Token
	Quote       Token
	PairAddress string   `json:"pairAddress"`
	Network     string   `json:"network"`
	Pool        string   `json:"pool"`
	Labels      []string `json:"labels"` // pool version. For exmaple: pancakeswap v2 / v3
	URL         string   `json:"url"`
}
type Token struct {
	Name     string `json:"token"`
	Address  string `json:"tokenAddress"`
	Decimals int    `json:"tokenDecimals"`
}

type PancakeswapV2Responce struct {
	Pool      string
	Timestamp string
	Price     *big.Float
	Hex       string
	Type      string // "BUY" / "SELL"
}

type TokenMeta struct {
	Symbol   string
	Address  common.Address
	Decimals uint8
}

type Whitelist struct {
	Network      string   `json:"network"`      // solana
	NetworkShort string   `json:"networkShort"` // SOL
	Pools        []string `json:"pools"`        // radium
}