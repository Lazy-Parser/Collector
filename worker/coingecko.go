package worker

import (
	"context"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/chains"
	worker_internal "github.com/Lazy-Parser/Collector/internal/worker"
	"github.com/Lazy-Parser/Collector/market"
)

type CoingeckoWorker interface {
	CreateChunks(tokens []market.Token) []market.Chunk
	FetchDecimals(ctx context.Context, chunk market.Chunk) (market.CGResponse, error)
}

func NewCoingeckoWorker(api api.CoingeckoApi, chains *chains.Chains) CoingeckoWorker {
	return worker_internal.NewCoingeckoWorker(api, chains)
}
