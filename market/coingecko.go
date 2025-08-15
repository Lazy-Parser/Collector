package market

// CG - coingecko

type CGRequestGroup struct {
	network string
	tokens  []Token
}

func (rg CGRequestGroup) GetSize() int {
	return len(rg.tokens)
}

func (rg CGRequestGroup) Push(token Token) {
	rg.tokens = append(rg.tokens, token)
}

// Root object
type CGResponse struct {
	Data []Token `json:"data"`
}

// ------
// One token entry
type CGToken struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Attributes    CGTokenAttributes `json:"attributes"`
	Relationships CGTokenRelations  `json:"relationships"`
}

// Attributes of a token
type CGTokenAttributes struct {
	Address               string      `json:"address"`
	Name                  string      `json:"name"`
	Symbol                string      `json:"symbol"`
	Decimals              int         `json:"decimals"`
	ImageURL              *string     `json:"image_url"`               // nullable
	CoingeckoCoinID       *string     `json:"coingecko_coin_id"`       // nullable
	TotalSupply           string      `json:"total_supply"`            // numbers are strings in the payload
	NormalizedTotalSupply string      `json:"normalized_total_supply"` // numbers are strings
	PriceUSD              string      `json:"price_usd"`
	FDVUSD                string      `json:"fdv_usd"`
	TotalReserveInUSD     string      `json:"total_reserve_in_usd"`
	VolumeUSD             CGVolumeUSD `json:"volume_usd"`
	MarketCapUSD          *string     `json:"market_cap_usd"` // nullable
}

// Nested object for volumes
type CGVolumeUSD struct {
	H24 string `json:"h24"`
}

// Relationships section
type CGTokenRelations struct {
	TopPools CGTopPools `json:"top_pools"`
}

// List of pool references
type CGTopPools struct {
	Data []CGPoolRef `json:"data"`
}

// Minimal pool reference
type CGPoolRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
