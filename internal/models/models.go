package models

import (
	"reflect"
	"strconv"
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
// type TickerMessage struct {
// 	Symbol  string     `json:"symbol"`
// 	Data    TickerData `json:"data"`
// 	Channel string     `json:"channel"`
// 	Ts      int64      `json:"ts"`
// }

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

// contains only `string` fields. Do not use it for publishing. Instead, use SpotData struct for it
type SpotDataTicker struct {
	Symbol     string `json:"s"`      // Trading pair
	Price      string `json:"p"`      // Last price
	Change     string `json:"r"`      // 24h change %
	TrueChange string `json:"tr"`     // True 24h change %
	High       string `json:"h"`      // 24h high
	Low        string `json:"l"`      // 24h low
	VolumeUSDT string `json:"v"`      // Quote volume
	VolumeBase string `json:"q"`      // Base volume
	LastRT     string `json:"lastRT"` // Possibly latency (always -1?)
	MT         string `json:"MT"`     // Market type ("0")
	NV         string `json:"NV"`     // Possibly placeholder ("--")
}

type SpotMiniTickersResponse struct {
	Data    []SpotDataTicker `json:"d"`
	Channel string           `json:"c"` // "spot@public.miniTickers.v3.api@UTC+8"
	Ts      int64            `json:"t"` // Global timestamp
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

// move this func to somewhere
func TickerToSpotData(src SpotDataTicker) SpotData {
	var result SpotData

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(&result).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		dstField := dstVal.Field(i)

		if dstField.Kind() == reflect.String {
			dstField.SetString(srcField.String())
			continue
		}

		// try to convert string to float64
		valStr := srcField.String()
		num, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			num = 0
		}
		dstField.SetFloat(num)
	}

	return result
}
