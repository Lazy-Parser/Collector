package api_dexscreener

type DexscreenerResponse = []PairDS

// responce from DexScreener
type PairDS struct {
	ChainID     string      `json:"chainId"`
	DexID       string      `json:"dexId"`
	URL         string      `json:"url"`
	PairAddress string      `json:"pairAddress"`
	Labels      []string    `json:"labels"`
	BaseToken   TokenDS     `json:"baseToken"`
	QuoteToken  TokenDS     `json:"quoteToken"`
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

type TokenDS struct {
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

// func normalizePair(pair PairDS, chainsService *chains.Chains) (market.PairCandidat, error) {
// 	globalNetwork, ok := chainsService.Select(pair.ChainID).ToBase()
// 	if !ok {
// 		return market.PairCandidat{}, fmt.Errorf("failed to change dexscreener network name to the global one. Dexscreener: %s", pair.ChainID)
// 	}
// 	var label string
// 	if len(pair.Labels) == 0 {
// 		label = ""
// 	} else {
// 		label = pair.Labels[0]
// 	}

// 	normalized := market.PairCandidat{
// 		BaseToken: market.Token{
// 			Name:    pair.BaseToken.Symbol,
// 			Address: pair.BaseToken.Address,
// 			Decimal: 0, // not specified
// 		},
// 		QuoteToken: market.Token{
// 			Name:    pair.QuoteToken.Symbol,
// 			Address: pair.QuoteToken.Address,
// 			Decimal: 0, // not specified
// 		},
// 		Address: pair.PairAddress,
// 		Network: globalNetwork,
// 		Pool:    pair.DexID,

// 		Volume:    pair.Volume.H24,
// 		Liquidity: market.Liquidity(pair.Liquidity),

// 		Label: label, // TODO: think about make market.Pair Label from string to []string
// 		URL:   pair.URL,
// 		// Type:        pairType,
// 		// PriceNative: pair.PriceNative,
// 		// PriceUsd:    pair.PriceUSD,
// 	}

// 	return normalized, nil
// }
