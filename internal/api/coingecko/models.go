package api_coingecko

import market "github.com/Lazy-Parser/Collector/internal/domain/market"

// ------
type requestGroup struct {
	network string
	tokens  []market.Token
}

func (rg requestGroup) GetSize() int {
	return len(rg.tokens)
}

func (rg requestGroup) Push(token market.Token) {
	rg.tokens = append(rg.tokens, token)
}

// Root object
type Response struct {
	Data []Token `json:"data"`
}

// ------
// One token entry
type Token struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Attributes    TokenAttributes `json:"attributes"`
	Relationships TokenRelations  `json:"relationships"`
}

// Attributes of a token
type TokenAttributes struct {
	Address               string    `json:"address"`
	Name                  string    `json:"name"`
	Symbol                string    `json:"symbol"`
	Decimals              int       `json:"decimals"`
	ImageURL              *string   `json:"image_url"`               // nullable
	CoingeckoCoinID       *string   `json:"coingecko_coin_id"`       // nullable
	TotalSupply           string    `json:"total_supply"`            // numbers are strings in the payload
	NormalizedTotalSupply string    `json:"normalized_total_supply"` // numbers are strings
	PriceUSD              string    `json:"price_usd"`
	FDVUSD                string    `json:"fdv_usd"`
	TotalReserveInUSD     string    `json:"total_reserve_in_usd"`
	VolumeUSD             VolumeUSD `json:"volume_usd"`
	MarketCapUSD          *string   `json:"market_cap_usd"` // nullable
}

// Nested object for volumes
type VolumeUSD struct {
	H24 string `json:"h24"`
}

// Relationships section
type TokenRelations struct {
	TopPools TopPools `json:"top_pools"`
}

// List of pool references
type TopPools struct {
	Data []PoolRef `json:"data"`
}

// Minimal pool reference
type PoolRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
