package discovery_internal

import (
	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/chains"
	"github.com/Lazy-Parser/Collector/market"
)

type Discovery struct {
	dexscreener api.DexscreenerApi
	coingecko   api.CoingeckoApi
	chains      *chains.Chains

	// only new mode
	onlyNewMode bool
	tokenRepo   market.TokenRepo
	pairRepo    market.PairRepo // TODO: rename to the PoolRepo

	decimalsMode bool
}

func NewDiscovery(dsApi api.DexscreenerApi, coingecko api.CoingeckoApi, chains *chains.Chains) *Discovery {
	return &Discovery{dexscreener: dsApi, coingecko: coingecko, chains: chains, onlyNewMode: false, decimalsMode: false}
}
