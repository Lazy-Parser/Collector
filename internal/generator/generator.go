package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	// "sync/atomic"

	core "github.com/Lazy-Parser/Collector/internal/core"
	database "github.com/Lazy-Parser/Collector/internal/database"
	"github.com/Lazy-Parser/Collector/internal/utils"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/time/rate"
)

var (
	dexEnpoint = "https://api.dexscreener.com/token-pairs/v1/" // {chainId}/{tokenAddress}
	minVolume  = 100000.0                                      // minimum volume for pair - 100k$
	usdtMap    = map[string]string{
		"ethereum": "0xdAC17F958D2ee523a2206206994597C13D831ec7",
		"bsc":      "0x55d398326f99059fF775485246999027B3197955",
	}
)

func Run() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	ctxLimitter := context.Background()
	defer ctxCancel()

	var (
		limiter       = rate.NewLimiter(rate.Limit(4), 4)
		mu            sync.Mutex
		store         []core.Pair
		quoteStoreMap map[QuoteToken]struct{} // map of unique quote tokens
		// counter int32
		pw = progress.NewWriter()
	)
	quoteStoreMap = make(map[QuoteToken]struct{})

	// load all tokens from mexc
	wl, err := utils.LoadWhitelistFile()
	if err != nil {
		log.Panicf("Failed to load whitelist! %v", err)
	}

	err = MexcInit()
	if err != nil {
		log.Panicf("Mexc init: %v", err)
	}
	MexcCompare(&wl)
	tokens := MexcGetTokens()

	// loading bar
	pw.SetSortBy(progress.SortByPercentDsc)
	pw.SetStyle(progress.StyleDefault)
	pw.SetNumTrackersExpected(1)
	pw.SetAutoStop(true)

	go pw.Render()

	tracker := &progress.Tracker{
		Message: "Loading",
		Total:   int64(len(tokens)),
		Units:   progress.UnitsDefault,
	}
	pw.AppendTracker(tracker)

	// find pairAddress and pool name from dexscreener
	// start worker pool
	pool1 := pool.New().WithMaxGoroutines(4)

	// иногда количество network в asset больше 1, нужно удалять ненужные в файле mexc.go
	for _, token := range tokens {
		select {
		case <-ctx.Done():
			log.Println("Stopping loading...")
			return

		default:
			pool1.Go(func() {
				if err := limiter.Wait(ctxLimitter); err != nil {
					return
				}

				// increment progress bar
				mu.Lock()
				tracker.Increment(1)
				mu.Unlock()

				pairs, err := fetchPairs(token.NetworkList[0].Network, token.NetworkList[0].Contract)
				if err != nil || len(*pairs) == 0 {
					fmt.Errorf("Fetch DexScreener: %v", err)
					return
				}

				selectedPair := validatePairs(*pairs)
				// if selectedPair if empry, then all provided pairs were bad, then do not save it
				if selectedPair.Volume.H24 == -1 {
					return
				}

				normalizedPair := normalizePair(selectedPair, "base")

				quoteToken := QuoteToken{
					Address: normalizedPair.Quote.Address,
					Name:    normalizedPair.Quote.Name,
					Symbol:  normalizedPair.Quote.Symbol,
					Network: normalizedPair.Network,
				}

				// save pair
				mu.Lock()
				store = append(store, *normalizedPair)
				quoteStoreMap[quoteToken] = struct{}{}
				mu.Unlock()
			})
		}
	}

	// We need to cast pairs prices to usdt, so we need to add pairs quoteToken/USDT.
	var quoteList []QuoteToken
	for t := range quoteStoreMap {
		quoteList = append(quoteList, t)
	}

	pool1.Wait()

	list := LoadQuoteChangerPairs(ctx, quoteList)
	for _, elem := range list {
		store = append(store, elem)
	}

	pw.Stop()
	savePairs(&store)
}

func LoadQuoteChangerPairs(ctx context.Context, quoteList []QuoteToken) []core.Pair {
	var mu sync.Mutex
	list := make([]core.Pair, 100)
	ctxLimitter := context.Background()
	var limiter = rate.NewLimiter(rate.Limit(4), 4)

	pool2 := pool.New().WithMaxGoroutines(4)
	for _, token := range quoteList {
		select {
		case <-ctx.Done():
			log.Println("Stopping loading...")
			return nil

		default:
			pool2.Go(func() {
				if err := limiter.Wait(ctxLimitter); err != nil {
					return
				}

				// skip those tokens, they are already usdt. (usdc equals to usdt, the difference is ~0.004%, so we can ignore it)
				if token.Symbol == "USDC" || token.Symbol == "USDT" || token.Symbol == "USD1" {
					return
				}

				res, err := FetchQuotePair(token.Network, token.Address, usdtMap[token.Network])
				if err != nil {
					fmt.Println(err)
					return
				}
				if len(res) == 0 {
					fmt.Printf("DexScreener API return nothing for '%s', '%s', '%s'", token.Network, token.Address, usdtMap[token.Network])
					return
				}

				normalizedPair := normalizePair(&res[0], "quote")
				fmt.Printf("%+v\n", normalizedPair)

				//save a quote pair
				mu.Lock()
				list = append(list, *normalizedPair)
				mu.Unlock()
			})
		}
	}

	pool2.Wait()

	return list
}

// methods
func fetchPairs(network, tokenAddress string) (*DexScreenerResponse, error) {
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
	bestToken := &PairDS{Volume: Volume{H24: -1}} // create empty result
	//var curQuoteSymbol string
	var curVolume24 float64

	for _, pair := range pairs {
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
			bestToken = &pair
		}
	}

	return bestToken
}

func normalizePair(pair *PairDS, pairType string) *core.Pair {
	normalized := &core.Pair{
		Base: core.Token{
			Name:     pair.BaseToken.Symbol,
			Address:  pair.BaseToken.Address,
			Decimals: -1,
			Symbol:   pair.BaseToken.Symbol,
		},
		Quote: core.Token{
			Name:     pair.QuoteToken.Symbol,
			Address:  pair.QuoteToken.Address,
			Decimals: -1,
			Symbol:   pair.QuoteToken.Symbol,
		},
		PairAddress: pair.PairAddress,
		Network:     pair.ChainID,
		Pool:        pair.DexID,
		Labels:      pair.Labels,
		URL:         pair.URL,
		Type:        pairType,
		PriceNative: pair.PriceNative,
		PriceUsd:    pair.PriceUSD,
	}

	return normalized
}

func FetchQuotePair(network string, address string, usdtAddress string) (DexScreenerResponse, error) {
	var res DexScreenerResponse = []PairDS{}

	url := "https://api.dexscreener.com/tokens/v1/" + network + "/" + address + "," + usdtAddress

	resp, err := http.Get(url)
	if err != nil {
		return res, fmt.Errorf("failed to fetch data from DexScreener API for '%s', '%s', '%s' pair. %v", address, usdtAddress, network, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, fmt.Errorf("failed to read body from DexScreener API for '%s', '%s', '%s' pair %v", address, usdtAddress, network, err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return res, fmt.Errorf("failed to parse body from DexScreener API for '%s', '%s', '%s' pair. %v", network, usdtAddress, network, err)
	}

	return res, nil
}

func savePairs(pairs *[]core.Pair) error {
	// NEW - save to sqlite
	db := database.GetDB()
	if !db.IsInitied {
		log.Panic("Trying to use not inited database in generator.go!")
	}

	for _, p := range *pairs {
		// save base token. Check if exists, if yes - set found token as new (baseToken)
		payload := database.Token{Name: p.Base.Name, Address: p.Base.Address, Decimals: -1}
		baseToken, err := database.GetDB().TokenService.SaveOrFind(&payload)
		if err != nil {
			return fmt.Errorf("database error: %v", err)
		}

		// save quote token. Check if exists, if yes - set found token as new (quoteToken)
		payload = database.Token{Name: p.Quote.Name, Address: p.Quote.Address, Decimals: -1}
		quoteToken, err := database.GetDB().TokenService.SaveOrFind(&payload)
		if err != nil {
			return fmt.Errorf("database error: %v", err)
		}

		pair := database.Pair{
			BaseTokenID:  baseToken.ID,
			QuoteTokenID: quoteToken.ID,
			PairAddress:  p.PairAddress,
			Network:      p.Network,
			Pool:         p.Pool,
			URL:          p.URL,
			Label:        getLabel(p.Labels),
			Type:         p.Type,
		}
		err = db.PairService.SavePair(&pair)
		if err != nil {
			return fmt.Errorf("database error: %v", err)
		}
	}

	log.Println("Saved! Total saved: ", len(*pairs))

	return nil
}

func getLabel(arr []string) string {
	if len(arr) == 0 {
		return ""
	}

	return arr[0]
}
