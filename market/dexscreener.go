package market

type DexscreenerResponse = []DSPair

// responce from DexScreener
type DSPair struct {
	ChainID     string        `json:"chainId"`
	DexID       string        `json:"dexId"`
	URL         string        `json:"url"`
	PairAddress string        `json:"pairAddress"`
	Labels      []string      `json:"labels"`
	BaseToken   DSToken       `json:"baseToken"`
	QuoteToken  DSToken       `json:"quoteToken"`
	PriceNative string        `json:"priceNative"`
	PriceUSD    string        `json:"priceUsd"`
	Txns        DSTxns        `json:"txns"`
	Volume      DSVolume      `json:"volume"`
	PriceChange DSPriceChange `json:"priceChange"`
	Liquidity   DSLiquidity   `json:"liquidity"`
	FDV         float64       `json:"fdv"`
	MarketCap   float64       `json:"marketCap"`
	PairCreated int64         `json:"pairCreatedAt"`
	Info        DSInfo        `json:"info"`
}

type DSToken struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

type DSQuoteToken struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
	Network string `json:"network"`
}

type DSTxns struct {
	M5  DSTxnCount `json:"m5"`
	H1  DSTxnCount `json:"h1"`
	H6  DSTxnCount `json:"h6"`
	H24 DSTxnCount `json:"h24"`
}

type DSTxnCount struct {
	Buys  int `json:"buys"`
	Sells int `json:"sells"`
}

type DSVolume struct {
	H24 float64 `json:"h24"`
	H6  float64 `json:"h6"`
	H1  float64 `json:"h1"`
	M5  float64 `json:"m5"`
}

type DSPriceChange struct {
	H1  float64 `json:"h1,omitempty"`
	H6  float64 `json:"h6,omitempty"`
	H24 float64 `json:"h24"`
}

type DSLiquidity struct {
	USD   float64 `json:"usd"`
	Base  float64 `json:"base"`
	Quote float64 `json:"quote"`
}

type DSInfo struct {
	ImageURL  string          `json:"imageUrl"`
	Header    string          `json:"header"`
	OpenGraph string          `json:"openGraph"`
	Websites  []DSLabeledLink `json:"websites"`
	Socials   []DSSocialLink  `json:"socials"`
}

type DSLabeledLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type DSSocialLink struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}
