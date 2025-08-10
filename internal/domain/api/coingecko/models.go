package worker_coingecko

import market "github.com/Lazy-Parser/Collector/internal/domain/market"

type Chunk struct {
	network string
	tokens  []market.Token
}
