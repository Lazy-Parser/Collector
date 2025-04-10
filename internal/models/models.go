package models

import (
	"time"
)

// ----- DEX -----
type GetInfoCoin struct {
	chainId     string
	dexId       string
	pairAddress string
	liquidity   int64
	price       string
}

type DexInfoResponse struct {
	SchemaVersion string `json:"schemaVersion"`
	Pairs         []pair `json:"pairs"`
}

type pair struct {
	ChainID       string             `json:"chainId"`
	DexID         string             `json:"dexId"`
	URL           string             `json:"url"`
	PairAddress   string             `json:"pairAddress"`
	Labels        []string           `json:"labels"`
	BaseToken     token              `json:"baseToken"`
	QuoteToken    token              `json:"quoteToken"`
	PriceNative   string             `json:"priceNative"`
	PriceUSD      string             `json:"priceUsd"`
	Txns          map[string]txn     `json:"txns"`        // ANY_ADDITIONAL_PROPERTY
	Volume        map[string]float64 `json:"volume"`      // ANY_ADDITIONAL_PROPERTY
	PriceChange   map[string]float64 `json:"priceChange"` // ANY_ADDITIONAL_PROPERTY
	Liquidity     liquidity          `json:"liquidity"`
	FDV           float64            `json:"fdv"`
	MarketCap     float64            `json:"marketCap"`
	PairCreatedAt int64              `json:"pairCreatedAt"`
	Info          pairInfo           `json:"info"`
	Boosts        boosts             `json:"boosts"`
}

type token struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

type txn struct {
	Buys  int `json:"buys"`
	Sells int `json:"sells"`
}

type liquidity struct {
	USD   float64 `json:"usd"`
	Base  float64 `json:"base"`
	Quote float64 `json:"quote"`
}

type pairInfo struct {
	ImageURL string        `json:"imageUrl"`
	Websites []website     `json:"websites"`
	Socials  []socialMedia `json:"socials"`
}

type website struct {
	URL string `json:"url"`
}

type socialMedia struct {
	Platform string `json:"platform"`
	Handle   string `json:"handle"`
}

type boosts struct {
	Active int `json:"active"`
}

// ----- DEX -----

// ----- storage -----
type Coin struct {
	Pair      string // BTC/USDT
	Mexc      *MexcInfo
	Dex       *DexInfo
	MexcReady bool // show if data from mexc is getted
	DexReady  bool // show if data from dex is getted
	Timestamp time.Time
}

type MexcInfo struct {
	MexcLink string // https://mexc.com/sdkhfjsbdjfsdf/sdfsfd
	
	// Bid      float32
	// Ask      float32
	Price    float32
	// TODO: make all fields
}

type DexInfo struct {
	DexLink string // https://dex.com/sdkhfjsbdjfsdf/sdfsfd
	Price   float32
	// TODO: make all fields
}

// ----- storage -----
