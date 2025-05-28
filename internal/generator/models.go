package generator

type DexScreenerResponse []PairDS

// responce from DexScreener
type PairDS struct {
	ChainID     string      `json:"chainId"`
	DexID       string      `json:"dexId"`
	URL         string      `json:"url"`
	PairAddress string      `json:"pairAddress"`
	Labels      []string    `json:"labels"`
	BaseToken   Token       `json:"baseToken"`
	QuoteToken  Token       `json:"quoteToken"`
	PriceNative string      `json:"priceNative"`
	PriceUSD    string      `json:"priceUsd"`
	Txns        Txns        `json:"txns"`
	Volume      Volume      `json:"volume"`
	PriceChange PriceChange `json:"priceChange"`
	Liquidity   Liquidity   `json:"liquidity"`
	FDV         float64     `json:"fdv"`
	MarketCap   float64     `json:"marketCap"`
	PairCreated int64       `json:"pairCreatedAt"`
	Info        Info        `json:"info"`
}

type Token struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

type QuoteToken struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
	Network string `json:"network"`
}

type Txns struct {
	M5  TxnCount `json:"m5"`
	H1  TxnCount `json:"h1"`
	H6  TxnCount `json:"h6"`
	H24 TxnCount `json:"h24"`
}

type TxnCount struct {
	Buys  int `json:"buys"`
	Sells int `json:"sells"`
}

type Volume struct {
	H24 float64 `json:"h24"`
	H6  float64 `json:"h6"`
	H1  float64 `json:"h1"`
	M5  float64 `json:"m5"`
}

type PriceChange struct {
	H1  float64 `json:"h1,omitempty"`
	H6  float64 `json:"h6,omitempty"`
	H24 float64 `json:"h24"`
}

type Liquidity struct {
	USD   float64 `json:"usd"`
	Base  float64 `json:"base"`
	Quote float64 `json:"quote"`
}

type Info struct {
	ImageURL  string        `json:"imageUrl"`
	Header    string        `json:"header"`
	OpenGraph string        `json:"openGraph"`
	Websites  []LabeledLink `json:"websites"`
	Socials   []SocialLink  `json:"socials"`
}

type LabeledLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type SocialLink struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// ---------- mexc ----------
// get data from mexc
type Network struct {
	NetworkShort  string `json:"network"` // "SOL"
	Network       string // "solana"
	Contract      string `json:"contract"` // 0x…
	DepositEnable bool   `json:"depositEnable"`
}
type Asset struct {
	Coin        string    `json:"coin"`        // "ETH"
	NetworkList []Network `json:"networkList"` // all chains
}

// https://contract.mexc.com/api/v1/contract/detail - all futures pairs
type Contracts struct {
	Data []ContractDetail `json:"data"`
}
type ContractDetail struct {
	BaseCoin string `json:"baseCoin"`
	Symbol   string `json:"symbol"` // "BTC_USDT"
}
