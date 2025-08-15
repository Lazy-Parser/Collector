package worker_internal

import (
	"context"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/chains"
	"github.com/Lazy-Parser/Collector/market"
)

type CoingeckoWorker struct {
	api    api.CoingeckoApi
	chains *chains.Chains
}

func NewCoingeckoWorker(api api.CoingeckoApi, chains *chains.Chains) *CoingeckoWorker {
	return &CoingeckoWorker{
		api:    api,
		chains: chains,
	}
}

func (worker *CoingeckoWorker) FetchDecimals(ctx context.Context, chunk market.Chunk) (market.CGResponse, error) {
	return worker.api.GetTokenData(ctx, chunk.Network, chunk.GetAddresses())
}

func (worker *CoingeckoWorker) CreateChunks(tokens []market.Token) []market.Chunk {
	chunks := make([]market.Chunk, 0, 32)
loop:
	for _, token := range tokens {
		for i := range chunks {
			// try to find existing chunk
			if chunks[i].Network == token.Network && len(chunks[i].Tokens) < market.ChunkMaxSize {
				chunks[i].Push(token)
				continue loop
			}
		}

		// none found, create one
		arr := make([]market.Token, 0, market.ChunkMaxSize)
		arr = append(arr, token)
		chunks = append(chunks, market.Chunk{
			Network: token.Network,
			Tokens:  arr,
		})
	}

	return chunks
}
