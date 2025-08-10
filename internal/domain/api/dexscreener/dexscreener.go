package worker_dexscreener

import (
	api_dexscreener "github.com/Lazy-Parser/Collector/internal/api/dexscreener"
	"github.com/Lazy-Parser/Collector/internal/common/chains"
	market "github.com/Lazy-Parser/Collector/internal/domain/market"
	"context"
	"errors"
	"fmt"
)

type DexscreenerWorker struct {
	api           *api_dexscreener.DexscreenerAPI
	chainsService chains.Chains
}

func NewWorker(
	dexscreenerApi *api_dexscreener.DexscreenerAPI,
	chainsService *chains.Chains,
) *DexscreenerWorker {
	return &DexscreenerWorker{
		api:           dexscreenerApi,
		chainsService: *chainsService,
	}
}

func (worker *DexscreenerWorker) FetchPairByToken(ctx context.Context, token market.Token) (market.Pair, error) {
	normalizedNetwork, _ := worker.chainsService.Select(token.Network).ToDexscreener()
	res, err := worker.api.GetTokenPairs(ctx, normalizedNetwork, token.Address)
	if err != nil {
		return market.Pair{}, fmt.Errorf("failed to fetch pair from dexscreener: \nToken: %+v \n Error: %v", token, err)
	}

	// return empty pair if dexscreener returns empty too.
	if len(*res) == 0 {
		return market.Pair{}, nil
	}

	// select the best pair from arr
	bestPair, ok := selectBest(res, 70_000.0) // min volume in 70k$
	if !ok {
		return market.Pair{}, nil
	}

	// normalize to market.Pair
	pair := normalizePair(bestPair)

	// change network name to the global
	globalNework, ok := worker.chainsService.Select(pair.Network).ToBase()
	if !ok {
		err := errors.New(fmt.Sprintf("Failed to cast dexscreener like network nemae to the global name. Failed on: %s", pair.Network))
		return market.Pair{}, err
	}
	pair.Network = globalNework

	return pair, nil
}

func selectBest(data *[]api_dexscreener.PairDS, minVolume float64) (api_dexscreener.PairDS, bool) {
	bestToken := api_dexscreener.PairDS{Volume: api_dexscreener.Volume{H24: 0}} // create empty result
	ok := false

	//var curQuoteSymbol string
	var curVolume24 float64

	for _, pair := range *data {
		//curQuoteSymbol = pair.QuoteToken.Symbol
		curVolume24 = pair.Volume.H24

		// filter by quote token. Only SOL, USDC, USDT allowed
		//if curQuoteSymbol != "SOL" &&
		//	curQuoteSymbol != "USDC" &&
		//	curQuoteSymbol != "USDT" &&
		//	curQuoteSymbol != "WBNB" {
		//	continue
		//}

		// filter by volume
		if curVolume24 < minVolume {
			continue
		}

		// TODO: maybe add filter by liquidity
		// filter by liquidity
		// if pair.Liquidity.USD < 10000 {
		// 	continue
		// }

		// TODO: add filter by allowed pools!!!!! VERY IMPORTANT

		// select pair with the biggest volume
		if curVolume24 > bestToken.Volume.H24 {
			bestToken = pair
			ok = true
		}
	}

	// mapping PairCandidat -> Pair
	return bestToken, ok
}

func normalizePair(pair api_dexscreener.PairDS) market.Pair {
	var label string
	if len(pair.Labels) == 0 {
		label = ""
	} else {
		label = pair.Labels[0]
	}

	normalized := market.Pair{
		BaseToken: market.Token{
			Name:    pair.BaseToken.Symbol,
			Address: pair.BaseToken.Address,
			Decimal: 0, // not specified
			Network: pair.ChainID,
		},
		QuoteToken: market.Token{
			Name:    pair.QuoteToken.Symbol,
			Address: pair.QuoteToken.Address,
			Decimal: 0, // not specified
			Network: pair.ChainID,
		},
		Address: pair.PairAddress,
		Network: pair.ChainID,
		Pool:    pair.DexID,

		Label: label, // TODO: think about make market.Pair Label from string to []string
		URL:   pair.URL,
		// Type:        pairType,
		// PriceNative: pair.PriceNative,
		// PriceUsd:    pair.PriceUSD,
	}

	return normalized
}
