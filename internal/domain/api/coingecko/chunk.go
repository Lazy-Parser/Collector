package worker_coingecko

import market "github.com/Lazy-Parser/Collector/internal/domain/market"

const ChunkMaxSize = 30

func (chunk *Chunk) Push(token market.Token) {
	chunk.tokens = append(chunk.tokens, token)
}

func (chunk *Chunk) GetAddresses() []string {
	var res []string

	for _, token := range chunk.tokens {
		res = append(res, token.Address)
	}

	return res
}
