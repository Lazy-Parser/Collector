// Package api provides and ability to interact with Dexscreener / Mexc / CoinGecko API with already set limiter.
package api

import (
	"context"

	"github.com/Lazy-Parser/Collector/config"
	api_internal "github.com/Lazy-Parser/Collector/internal/adapter/api"
	"github.com/Lazy-Parser/Collector/market"
)

type DexscreenerApi interface {
	GetTokenPairs(ctx context.Context, network string, address string) (*market.DexscreenerResponse, error)
}

func NewDexscreenerApi(cfg *config.Config) DexscreenerApi {
	return api_internal.NewDexscreenerApi(cfg)
}
