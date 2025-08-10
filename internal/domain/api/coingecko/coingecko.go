package worker_coingecko

import (
	config "github.com/Lazy-Parser/Collector/config/service"
	api_coingecko "github.com/Lazy-Parser/Collector/internal/api/coingecko"
	market "github.com/Lazy-Parser/Collector/internal/domain/market"
	"context"
)

type CoingeckoWorker struct {
	cfg *config.Config
	api *api_coingecko.CoingeckoAPI
}

func NewWorker(cfg *config.Config, api *api_coingecko.CoingeckoAPI) *CoingeckoWorker {
	return &CoingeckoWorker{
		cfg: cfg,
		api: api,
	}
}

func (worker *CoingeckoWorker) FetchDecimals(ctx context.Context, chunk Chunk) (api_coingecko.Response, error) {
	return worker.api.GetTokenData(ctx, chunk.network, chunk.GetAddresses())
}

func (worker *CoingeckoWorker) CreateChunks(tokens []market.Token) []Chunk {
	chunks := make([]Chunk, 0, 32)
loop:
	for _, token := range tokens {
		for i := range chunks {
			// try to find existing chunk
			if chunks[i].network == token.Network && len(chunks[i].tokens) < ChunkMaxSize {
				chunks[i].Push(token)
				continue loop
			}
		}

		// none found, create one
		arr := make([]market.Token, 0, ChunkMaxSize)
		arr = append(arr, token)
		chunks = append(chunks, Chunk{
			network: token.Network,
			tokens:  arr,
		})
	}

	return chunks
}
