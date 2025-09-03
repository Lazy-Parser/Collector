package market

// CG - coingecko
type CGPoolRes struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Address         string  `json:"address"`
			Name            string  `json:"name"`
			PoolName        string  `json:"pool_name"`
			PoolCreatedAt   string  `json:"pool_created_at"`
			ReserveInUsd    string  `json:"reserve_in_usd"`
			LockedLiquidity *string `json:"locked_liquidity_percentage"`
			VolumeUsd       struct {
				M5  string `json:"m5"`
				M15 string `json:"m15"`
				M30 string `json:"m30"`
				H1  string `json:"h1"`
				H6  string `json:"h6"`
				H24 string `json:"h24"`
			} `json:"volume_usd"`
		} `json:"attributes"`

		Relationships struct {
			Dex struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"dex"`
		} `json:"relationships"`
	} `json:"data"`

	Included []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			// For tokens
			Address         string `json:"address,omitempty"`
			Name            string `json:"name,omitempty"`
			Symbol          string `json:"symbol,omitempty"`
			Decimals        int    `json:"decimals,omitempty"`
			ImageURL        string `json:"image_url,omitempty"`
			CoinGeckoCoinID string `json:"coingecko_coin_id,omitempty"`

			// For dex
			// (only name is provided, reused here)
		} `json:"attributes"`
	} `json:"included"`
}

// Root object
type CGResponse struct {
	Data []CGToken `json:"data"`
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

// ------- CHUNK ------------
type Chunk struct {
	Network string
	Tokens  []Token
}

const ChunkMaxSize = 30

func (chunk *Chunk) Push(token Token) {
	chunk.Tokens = append(chunk.Tokens, token)
}

func (chunk *Chunk) GetAddresses() []string {
	var res []string

	for _, token := range chunk.Tokens {
		res = append(res, token.Address)
	}

	return res
}
