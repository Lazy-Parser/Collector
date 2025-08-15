package api

import (
	"context"

	"github.com/Lazy-Parser/Collector/config"
	api_internal "github.com/Lazy-Parser/Collector/internal/adapter/api"
	"github.com/Lazy-Parser/Collector/market"
)

type MexcAPI interface {
	FetchCurrencyInformation(ctx context.Context) ([]market.MexcAsset, error)
	FetchContractInformation(ctx context.Context, url string) ([]market.MexcContractDetail, error)
}

func NewMexcApi(cfg *config.Config) MexcAPI {
	return api_internal.NewMexcApi(cfg)
}
