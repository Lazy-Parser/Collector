package worker

import (
	"context"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/chains"
	worker_internal "github.com/Lazy-Parser/Collector/internal/worker"
	"github.com/Lazy-Parser/Collector/market"
)

type DexscreenerWorker interface {
	FetchPairByToken(ctx context.Context, token market.Token) (market.Pair, error)
}

func NewDexscreenerWorker(api api.DexscreenerApi, chains *chains.Chains) DexscreenerWorker {
	return worker_internal.NewDexscreenerWorker(api, chains)
}
