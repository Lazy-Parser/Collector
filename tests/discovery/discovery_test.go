package discovery_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/chains"
	"github.com/Lazy-Parser/Collector/config"
	"github.com/Lazy-Parser/Collector/discovery"
	"github.com/Lazy-Parser/Collector/exchange"
	"github.com/Lazy-Parser/Collector/market"
)

func initApp() (*chains.Chains, *config.Config, error) {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "..", "config", "chains.json")
	chainsService, err := chains.NewChains(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup chains service. \nError: %v", err)
	}
	path = filepath.Join(wd, "..", "config", ".env")
	cfg, err := config.NewConfig(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup config service. \nError: %v", err)
	}

	return chainsService, cfg, nil
}

func getFuturesMexc(chainsService *chains.Chains, cfg *config.Config) ([]market.Token, error) {
	ctx := context.Background()

	mexcApi := api.NewMexcApi(cfg)
	mexc := exchange.NewMexc(mexcApi)

	s, err := mexc.Spots(ctx, true, -1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spots from mexc exchange. Reason: %v", err)
	}
	// filter by networks
	filtered := exchange.FilterByNetworks(&s, chainsService)

	f, err := mexc.Futures(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch futures from mexc exchange. Reason: %v", err)
	}

	return exchange.CompareFutures(&filtered, &f), nil
}

func TestPoolsFething(t *testing.T) {
	ctx := context.Background()

	chains, cfg, err := initApp()
	if err != nil {
		t.Error(err)
	}

	tokens, err := getFuturesMexc(chains, cfg)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Loaded %d futures from Mexc!", len(tokens))
	t.Log("Try to load more info...")

	dsApi := api.NewDexscreenerApi(cfg)
	cgApi := api.NewCoingeckoApi(cfg)
	discoveryService := discovery.NewDiscovery(dsApi, cgApi, chains)
	discoveryService.ApplyOptions(discoveryService.WithDecimals(true))

	// for first 20
	limit := 20
	empty := market.Pool{}
	var pools []market.Pool
	for i := range limit {
		pool, err := discoveryService.Meta(ctx, tokens[i])
		if err != nil {
			t.Errorf("Failed to fetch pool for payload: \nToken: %+v\nError: %v", tokens[i], err)
			continue
		}
		if pool == empty {
			// t.Logf("Found NOTHING for Token: \nName: %s | Address: %s", tokens[i].Name, tokens[i].Address)
			continue
		}

		t.Logf(
			"%s\n%s(%d | %s)\n%s(%d | %s)\n-------------",
			pool.Network,
			pool.Pair.BaseToken.Name,
			pool.Pair.BaseToken.Decimal,
			pool.Pair.BaseToken.Address,
			pool.Pair.QuoteToken.Name,
			pool.Pair.BaseToken.Decimal,
			pool.Pair.QuoteToken.Address,
		)
		pools = append(pools, pool)
	}

	// find quote pools: add quote pools manually

	t.Logf("Found pools %d from %d tokens", len(pools), limit) // Or len(tokens) in prod

	// try to add decimals

}
