package worker_internal

import (
	"context"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/chains"
	"github.com/Lazy-Parser/Collector/market"
)

type DexscreenerWorker struct {
	api    api.DexscreenerApi
	chains *chains.Chains
}

func NewDexscreenerWorker(api api.DexscreenerApi, chains *chains.Chains) *DexscreenerWorker {
	return &DexscreenerWorker{
		api:    api,
		chains: chains,
	}
}

func (dw *DexscreenerWorker) FetchPairByToken(ctx context.Context, token market.Token) (market.Pair, error) {
	// normalizedNetwork, _ := dw.chains.Select(token.Network).ToDexscreener()
	// res, err := dw.api.GetTokenPairs(ctx, normalizedNetwork, token.Address)
	// if err != nil {
	// 	return market.Pair{}, fmt.Errorf("failed to fetch pair from dexscreener: \nToken: %+v \n Error: %v", token, err)
	// }

	// // return empty pair if dexscreener returns empty too.
	// if len(*res) == 0 {
	// 	return market.Pair{}, nil
	// }

	// // select the best pair from arr
	// bestPair, ok := selectBest(res, 70_000.0) // min volume in 70k$
	// if !ok {
	// 	return market.Pair{}, nil
	// }

	// // normalize to market.Pair
	// pair := normalizePair(bestPair)
	// // cast network to the base style
	// network, _ := dw.chains.Select(pair.Network).ToBase()
	// pair.BaseToken.Network = network
	// pair.QuoteToken.Network = network
	// pair.Network = network

	// // change network name to the global
	// globalNetwork, ok := dw.chains.Select(pair.Network).ToBase()
	// if !ok {
	// 	err := errors.New("Failed to cast dexscreener like network name to the global name. Failed on: " + pair.Network)
	// 	return market.Pair{}, err
	// }
	// pair.Network = globalNetwork

	return market.Pair{}, nil
}
