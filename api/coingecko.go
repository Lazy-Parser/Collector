package api

import (
	"context"

	"github.com/Lazy-Parser/Collector/config"
	api_internal "github.com/Lazy-Parser/Collector/internal/adapter/api"
	"github.com/Lazy-Parser/Collector/market"
)

type CoingeckoApi interface {
	GetTokenData(ctx context.Context, network string, addresses []string) (market.CGResponse, error)
}

func NewCoingeckoApi(cfg *config.Config) CoingeckoApi {
	return api_internal.NewCoingeckoApi(cfg)
}
