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

// store symbol, chanId and pairAddress. We need it to get info about some coin
type PairMeta struct {
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
	Price float32
	// TODO: make all fields
}

type DexInfo struct {
	DexLink string // https://dex.com/sdkhfjsbdjfsdf/sdfsfd
	Price   float32
	// TODO: make all fields
}

// ----- storage -----

// ----- MEXC -------
// futures
type TickerMessage struct {
	Symbol  string     `json:"symbol"`
	Data    TickerData `json:"data"`
	Channel string     `json:"channel"`
	Ts      int64      `json:"ts"`
}

type TickerData struct {
	Symbol                  string    `json:"symbol"`
	LastPrice               float64   `json:"lastPrice"`
	RiseFallRate            float64   `json:"riseFallRate"`
	FairPrice               float32   `json:"fairPrice"`
	IndexPrice              float64   `json:"indexPrice"`
	Volume24                int64     `json:"volume24"`
	Amount24                float64   `json:"amount24"`
	MaxBidPrice             float64   `json:"maxBidPrice"`
	MinAskPrice             float64   `json:"minAskPrice"`
	Lower24Price            float64   `json:"lower24Price"`
	High24Price             float64   `json:"high24Price"`
	Timestamp               int64     `json:"timestamp"`
	Bid1                    float32   `json:"bid1"`
	Ask1                    float32   `json:"ask1"`
	HoldVol                 int64     `json:"holdVol"`
	RiseFallValue           float64   `json:"riseFallValue"`
	FundingRate             float64   `json:"fundingRate"`
	Zone                    string    `json:"zone"`
	RiseFallRates           []float64 `json:"riseFallRates"`
	RiseFallRatesOfTimezone []float64 `json:"riseFallRatesOfTimezone"`
}

type Tickers struct {
	Data []TickerData `json:data`
}

// futures

// spot

// spot

// "FUTURES" or "SPOT"


//{
//     "method": "SUBSCRIPTION",
//     "params": [
//         "spot@public.miniTicker.v3.api@BTCUSDT@UTC+8"
//     ]
// }
// ----- MEXC -------
