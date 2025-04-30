package generator

type DexScreenerResponse struct {
	SchemaVersion string `json:"schemaVersion"`
	Pairs         []Pair `json:"pairs"`
}

type Pair struct {
	ChainID       string      `json:"chainId"`
	DexID         string      `json:"dexId"`
	URL           string      `json:"url"`
	PairAddress   string      `json:"pairAddress"`
	Labels        []string    `json:"labels"`
	BaseToken     Token       `json:"baseToken"`
	QuoteToken    Token       `json:"quoteToken"`
	PriceNative   string      `json:"priceNative"`
	PriceUSD      string      `json:"priceUsd"`
	Txns          Txns        `json:"txns"`
	Volume        Volume      `json:"volume"`
	PriceChange   PriceChange `json:"priceChange"`
	Liquidity     Liquidity   `json:"liquidity"`
	FDV           float64     `json:"fdv"`
	MarketCap     float64     `json:"marketCap"`
	PairCreatedAt int64       `json:"pairCreatedAt"`
	Info          Info        `json:"info"`
	Boosts        Boosts      `json:"boosts"`
}

type Token struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

type Txns map[string]struct {
	Buys  int `json:"buys"`
	Sells int `json:"sells"`
}

type Volume struct {
	H24 float64 `json:"h24"`
	H6  float64 `json:"h6"`
	H1  float64 `json:"h1"`
	M5  float64 `json:"m5"`
}

type PriceChange map[string]float64

type Liquidity struct {
	USD   float64 `json:"usd"`
	Base  float64 `json:"base"`
	Quote float64 `json:"quote"`
}

type Info struct {
	ImageURL string    `json:"imageUrl"`
	Websites []Website `json:"websites"`
	Socials  []Social  `json:"socials"`
}

type Website struct {
	URL string `json:"url"`
}

type Social struct {
	Platform string `json:"platform"`
	Handle   string `json:"handle"`
}

type Boosts struct {
	Active int `json:"active"`
}


// -------
type PairNormalized struct {
	Pair              string `json:"pair"`
	PairAddress       string `json:"pairAddress"`
	BaseTokenAddress  string `json:"baseToken"`
	QuoteTokenAddress string `json:"quoteToken"`
	Network           string `json:"network"`
	Pull              string `json:"pull"`
	URL               string `json:"url"`
}

type Whitelist struct {
	Network string `json:"network"`
	Pools []string `json:"pools"`
}