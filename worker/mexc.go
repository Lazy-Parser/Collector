package worker

import (
	"context"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/chains"
	worker_internal "github.com/Lazy-Parser/Collector/internal/worker"
	"github.com/Lazy-Parser/Collector/market"
)

type MexcWorker interface {
	GetAllTokens(ctx context.Context) ([]market.MexcAsset, error)
	GetAllFutures(ctx context.Context) ([]market.MexcContractDetail, error)
	FindContractBySymbol(arr *[]market.MexcContractDetail, symbol string) (market.MexcContractDetail, bool)
}

func NewMexcWorker(api api.MexcAPI, chains *chains.Chains) MexcWorker {
	return worker_internal.NewMexcWorker(api, chains)
}
