package api

import (
	"context"

	"github.com/Lazy-Parser/Collector/config"
	api_internal "github.com/Lazy-Parser/Collector/internal/adapter/api"
	"github.com/Lazy-Parser/Collector/market"
)

// TODO: maybe add protobuff price streaming here on in exchange???
type MexcAPI interface {
	// config/getall
	FetchCurrencyInformation(ctx context.Context) ([]market.MexcAsset, error)
	// contract/detail
	FetchContractInformation(ctx context.Context) ([]market.MexcContractDetail, error)
	// ticker/24hr
	Fetch24hTickerStats(ctx context.Context) ([]market.MexcTickerStats, error)
}

func NewMexcApi(cfg *config.Config) MexcAPI {
	return api_internal.NewMexcApi(cfg)
}
