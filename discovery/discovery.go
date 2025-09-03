// The task of the package 'discovery' is to listen / fetch tokens from exchanges and find additional info (meta) and find pairs about them
package discovery

import (
	"context"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/chains"
	"github.com/Lazy-Parser/Collector/exchange"
	discovery_internal "github.com/Lazy-Parser/Collector/internal/discovery"
	"github.com/Lazy-Parser/Collector/market"
)

type Option = discovery_internal.Option

type Discovery interface {
	// Find additional info for each token. In Token only 'Address' and 'Network' are required. Use it of you need to find pool for base token!
	Meta(ctx context.Context, token market.Token) (market.Pool, error)
	MetaByAddress(ctx context.Context, network string, poolAddress string) (market.Pool, error)
	// DO NOT USE IT!!!
	Fetch(exchange exchange.Exchange) []market.Token
	// Options
	WithOnlyNew(enable bool, tokenRepo market.TokenRepo, pairRepo market.PairRepo) Option
	WithDecimals(enable bool) Option
	ApplyOptions(opts ...Option)
}

func NewDiscovery(dsApi api.DexscreenerApi, cgApi api.CoingeckoApi, chains *chains.Chains) Discovery {
	return discovery_internal.NewDiscovery(dsApi, cgApi, chains)
}
