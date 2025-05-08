package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/Lazy-Parser/Collector/internal/utils"
	d "github.com/Lazy-Parser/Collector/internal/domain"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/time/rate"
)

var (
	dexEnpoint = "https://api.dexscreener.com/token-pairs/v1/" // {chainId}/{tokenAddress}
	minVolume  = 100000.0                                      // minimum volume for pair - 100k$
)

func Run() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	ctxLimitter := context.Background()
	defer ctxCancel()

	var (
		limiter = rate.NewLimiter(rate.Limit(4), 4)
		mu      sync.Mutex
		store   []d.Pair
		counter int32
	)

	// load all tokens from mexc
	wl, err := loadWhitelistFile()
	if err != nil {
		log.Panicf("Failed to load whitelist! %v", err)
	}

	err = MexcInit()
	if err != nil {
		log.Panicf("Mexc init: %v", err)
	}
	MexcCompare(&wl)
	tokens := MexcGetTokens()

	// find pairAddress and pool name from dexscreener
	// start worker pool
	pool := pool.New().WithMaxGoroutines(4)

	// иногда количество network в asset больше 1, нужно удалять ненужные в файле mexc.go
	for _, token := range tokens {
		select {
		case <-ctx.Done():
			log.Println("Stopping loading...")
			return

		default:
			pool.Go(func() {
				if err := limiter.Wait(ctxLimitter); err != nil {
					return
				}

				pairs, err := fetchPair(token.NetworkList[0].Network, token.NetworkList[0].Contract)
				if err != nil || len(*pairs) == 0 {
					fmt.Errorf("Fetch DexScreener: %v", err)
					return
				}

				selectedPair := validatePairs(*pairs)
				// if selectedPair if empry, then all provided pairs were bad, then do not save it
				if selectedPair.Volume.H24 == 0 {
					return
				}

				normalizedPair := normalizePair(selectedPair, token.Coin)
				idx := atomic.AddInt32(&counter, 1)
				printReceivedToken(idx, normalizedPair)

				// save pair
				mu.Lock()
				store = append(store, *normalizedPair)
				mu.Unlock()
			})
		}
	}

	pool.Wait()
	savePairs(&store)
}

// methods
func fetchPair(network, tokenAddress string) (*DexScreenerResponse, error) {
	var res DexScreenerResponse = []PairDS{}

	if len(network) == 0 {
		return &res, fmt.Errorf("Network not provied")
	}

	url := dexEnpoint + network + "/" + tokenAddress

	resp, err := http.Get(url)
	if err != nil {
		return &res, fmt.Errorf("failed to fetch data from DexScreener API for '%s', '%s' pair. %v", network, tokenAddress, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &res, fmt.Errorf("failed to read body from DexScreener API for '%s', '%s' pair %v", network, tokenAddress, err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return &res, fmt.Errorf("failed to parse body from DexScreener API for '%s', '%s' pair. %v", network, tokenAddress, err)
	}

	return &res, nil
}

func validatePairs(pairs DexScreenerResponse) *PairDS {
	bestToken := &PairDS{Volume: Volume{H24: 0}} // create empty result
	var curQuoteSymbol string
	var curVolume24 float64

	for _, pair := range pairs {
		curQuoteSymbol = pair.QuoteToken.Symbol
		curVolume24 = pair.Volume.H24

		// filter by quote token. Only SOL, USDC, USDT allowed
		if curQuoteSymbol != "SOL" &&
			curQuoteSymbol != "USDC" &&
			curQuoteSymbol != "USDT" &&
			curQuoteSymbol != "WBNB" {
			continue
		}

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
			bestToken = &pair
		}
	}

	return bestToken
}

func normalizePair(pair *PairDS, symbol string) *d.Pair {
	normalized := &d.Pair{
		BaseToken:         symbol,
		QuoteToken:        pair.QuoteToken.Symbol,
		PairAddress:       pair.PairAddress,
		BaseTokenAddress:  pair.BaseToken.Address,
		QuoteTokenAddress: pair.QuoteToken.Address,
		Network:           pair.ChainID,
		Pool:              pair.DexID,
		Labels:            pair.Labels,
		URL:               pair.URL,
	}

	return normalized
}

func savePairs(pairs *[]d.Pair) {

	payload, err := json.MarshalIndent(pairs, "", "  ")
	if err != nil {
		log.Panicf("[savePairs] Failed to parse 'pairs' to []byte, %v", err)
	}

	workDirPath, err := os.Getwd()
	if err != nil {
		log.Panicf("Get work directory: %v", err)
	}
	path := filepath.Join(workDirPath, "config", "pairs.json")
	err = os.WriteFile(path, payload, 0644)
	if err != nil {
		log.Panicf("Write pairs to file: %v", err)
	}

	log.Println("Saved! Total saved: ", len(*pairs))
}

func loadWhitelistFile() ([]Whitelist, error) {
	workDir, err := utils.GetWorkDirPath()
	if err != nil {
		return []Whitelist{}, err
	}

	path := filepath.Join(workDir, "config", "network_pool_whitelist.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return []Whitelist{}, fmt.Errorf("loading 'config/network_pool_whitelist.json' file: %v", err)
	}

	var res []Whitelist
	err = json.Unmarshal(data, &res)
	if err != nil {
		return []Whitelist{}, fmt.Errorf("unmarshal data from 'config/network_pool_whitelist.json' file: %v", err)
	}

	return res, nil
}

func printReceivedToken(counter int32, pair *d.Pair) {
	fmt.Printf("%d) Token %s/%s | %s\n", counter, pair.BaseToken, pair.QuoteToken, pair.Network)
	fmt.Println(pair.URL)
	fmt.Println()
}
