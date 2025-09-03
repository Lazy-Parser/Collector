package discovery_internal

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Lazy-Parser/Collector/market"
)

var (
	emptyPool = market.Pool{}
)

func (d *Discovery) Meta(ctx context.Context, token market.Token) (market.Pool, error) {
	// STEP #1. Fetch Pool
	dsRes, err := d.fetchPool(ctx, token)
	if err != nil {
		return emptyPool, fmt.Errorf("failed to fetch pool: %v", err)
	}
	// check if non found
	if len(*dsRes) == 0 {
		return emptyPool, errors.New("Cannot find good pool from list!")
	}
	log.Println("Pool found!")
	// select the best pair from arr
	bestPair, ok := findBest(dsRes, 70_000.0) // min volume is 70k$. TODO: set voume via config (yaml for example or json)
	if !ok {
		return emptyPool, nil
	}
	// normalize to market.Pool
	pool := normalizePool(bestPair)

	// STEP #1.5 (is unique)
	if d.onlyNewMode {
		// TODO: check in database
	}

	// STEP #2. Fetch decimal. TODO: also can fetch create time
	if d.decimalsMode {
		cgNetwork, _ := d.chains.Select(pool.Network).ToCoingecko()
		base := pool.Pair.BaseToken.Address
		quote := pool.Pair.QuoteToken.Address
		res, err := d.coingecko.GetTokenData(ctx, cgNetwork, []string{base, quote}) // fetch both base and quote token
		if err != nil {
			return emptyPool, fmt.Errorf("failed to fetch decimal for payload: \nNetwork: %s | Address: %s | Name: %s \n error: %v", cgNetwork, token.Address, token.Name, err)
		}

		// some token not found
		if len(res.Data) < 2 {
			return emptyPool, nil
		}

		pool.Pair.BaseToken.Decimal = uint8(res.Data[0].Attributes.Decimals)
		pool.Pair.QuoteToken.Decimal = uint8(res.Data[1].Attributes.Decimals)
	}

	// cast network to the base style
	network, _ := d.chains.Select(pool.Network).ToBase()
	pool.Pair.BaseToken.Network = network
	pool.Pair.QuoteToken.Network = network
	pool.Network = network

	return pool, nil
}

// TODO: check this algorithm in future
// This function fetch info about every token in the [chunk].
//
// [Chunk] - is a group of tokens with the same network with maximum 30 elems
// Returns: a list of tokens (that was provided) with injected decimals.
// But pay attention! If coingecko does not know about your provided token, then the returned token will be without decimal
func (d *Discovery) Decimals(ctx context.Context, chunk market.Chunk) ([]market.Token, error) {
	cgNetwork, _ := d.chains.Select(chunk.Network).ToCoingecko()
	res, err := d.coingecko.GetTokenData(ctx, cgNetwork, chunk.GetAddresses())
	if err != nil {
		return nil, err
	}

	// res from coingecko has the same order as provided tokens. But some of them coingecko can skeep if he does not know token
	idx := 0 // index for res from coingecko
	for i := range chunk.Tokens {
		if chunk.Tokens[i].Address != res.Data[idx].Attributes.Address {
			// coingecko does not know info about this token, skip
			continue
		}

		chunk.Tokens[i].Decimal = uint8(res.Data[idx].Attributes.Decimals)
		idx++
	}

	return chunk.Tokens, nil
}

func (d *Discovery) fetchPool(ctx context.Context, token market.Token) (*market.DexscreenerResponse, error) {
	dsNetwork, ok := d.chains.Select(token.Network).ToDexscreener()
	if !ok {
		return nil, fmt.Errorf("failed to cast '%s' to the dexscreener type. Maybe provided netowrk name does not exist in our list or misspelle", token.Network)
	}
	return d.dexscreener.FetchPool(ctx, dsNetwork, token.Address)
}
func findBest(data *[]market.DSPair, minVolume float64) (market.DSPair, bool) {
	bestToken := market.DSPair{Volume: market.DSVolume{H24: minVolume}} // create empty result
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
func normalizePool(pair market.DSPair) market.Pool {
	var label string
	if len(pair.Labels) == 0 {
		label = ""
	} else {
		label = pair.Labels[0]
	}

	normalized := market.Pool{
		Pair: market.Pair{
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

func (d *Discovery) MetaByAddress(ctx context.Context, network string, poolAddress string) (market.Pool, error) {
	cgNetwork, ok := d.chains.Select(network).ToCoingecko()
	if !ok {
		return market.Pool{}, errors.New(fmt.Sprintf("failed to cast '%s' to the coingecko type", network))
	}
	res, err := d.coingecko.GetPoolInfo(ctx, cgNetwork, poolAddress)
	if err != nil {
		return market.Pool{}, fmt.Errorf("failed to find pool meta for:\nRequest: Network: %s | Pool Address: %s\nError: %v", network, poolAddress, err)
	}

	base := res.Included[0].Attributes
	quote := res.Included[1].Attributes
	globalNetwork, _ := d.chains.Select(network).ToBase()
	return market.Pool{
		Pair: market.Pair{
			BaseToken: market.Token{
				Name: base.Symbol,
				Address: base.Address,
				Network: globalNetwork,
				Decimal: uint8(base.Decimals),
			},
			QuoteToken: market.Token{
				Name: quote.Symbol,
				Address: quote.Address,
				Network: globalNetwork,
				Decimal: uint8(quote.Decimals),
			},
		},
		Address: res.Data.Attributes.Address,
		Network: globalNetwork,
		Pool: res.Data.Relationships.Dex.Data.ID,
	}, nil
}
