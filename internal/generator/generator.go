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
	"time"

	"github.com/Lazy-Parser/Collector/internal/domain"
	"github.com/Lazy-Parser/Collector/internal/impl/aggregator"
	"github.com/Lazy-Parser/Collector/internal/utils"
	"golang.org/x/time/rate"
)

// TODO: добавтиь WaitGroup что бы код не завершался
var (
	fetchPairUrl = "https://api.dexscreener.com/latest/dex/search"
)

// generate JSON file of all pairs from provided DataSource, also add info like pairAddress, pull name, networm name from dex api or coinGeko api
func Run(collector domain.DataSource) {
	log.Println("Setup generation...")
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
	ctxLimitter := context.Background()
	defer ctxCancel()

	// load whitelist json
	whitelist, err := loadWhitelistFile()
	if err != nil {
		log.Panicf("Loading whitelist file failed: %v", err)
	}

	var wg sync.WaitGroup

	// create joiner, because collector can put data only in aggregator
	aggregator.InitJoiner()
	joiner := aggregator.GetJoiner()

	// try to start collector. For example it will be mexc, then extract all data, and via DexcScreener api collect all data
	if err := collector.Connect(); err != nil {
		log.Panicf("Connect to %s collector. %v", collector.Name(), err)
	}
	if err := collector.Subscribe(); err != nil {
		log.Panicf("Subscribe to %s collectors events. %v", collector.Name(), err)
	}

	go collector.Run(ctx, joiner.Push, joiner.SetState)
	log.Println("Start analyzing")

	// get data
	// IMPORTANT! 'payload.Data' in aggregator is stored as interface{}. So we do not know what type excectly we get.
	// I will get data from MEXC, so i will cast to Mexc
	limitter := rate.NewLimiter(290, 290) // max - 300, but i use a little less
	var toSave []PairNormalized
	pairChan := make(chan *PairNormalized, 1000)
	counter := 1

	// Консумер
	go func() {
		for pair := range pairChan {
			toSave = append(toSave, *pair)
		}
	}()

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		// listen state, stop this goroutine when state is false (it means that collector stops)
		go func() {
			for {
				switch {
				case (!<-joiner.ListenState()):
					wg.Done()
					return
				}
			}
		}()

		for payload := range joiner.Stream() {
			symbols := payload.Symbol

			// i will try to get all info about pair from dexscreener api. It has limit 300 requests per second. So i use limitter
			wg.Add(1) // ДО запуска запроса
			go func(symbols string) {
				defer wg.Done()

				// fetch info
				if err := limitter.Wait(ctxLimitter); err != nil {
					log.Printf("Limiter error: %v", err)
					return
				}

				res, err := fetchPair(symbols)
				if err != nil {
					fmt.Errorf("fetching symbol from dexscreener api, %v", err)
					return
				}

				// find one pair from all copies from dexscreener resp
				pair := getOriginalPair(res, &whitelist)

				counter++
				printReceivedToken(counter, symbols)

				pairChan <- normalizePair(pair, symbols)
			}(symbols)
		}
	}(ctx)

	log.Println("Waiting for all requests...")
	wg.Wait()

	log.Println("Saving collected pairs...")
	savePairs(&toSave)
}

func fetchPair(symbols string) (*DexScreenerResponse, error) {
	var res DexScreenerResponse
	url := fetchPairUrl + "?q=" + symbols
	resp, err := http.Get(url)
	if err != nil {
		return &res, fmt.Errorf("failed to fetch data from DexScreener API for '%s' pair", symbols)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &res, fmt.Errorf("failed to read body from DexScreener API for '%s' pair", symbols)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return &res, fmt.Errorf("failed to parse body from DexScreener API for '%s' pair", symbols)
	}

	return &res, nil
}

// return pair with the biggest liquidity / volume
func getOriginalPair(data *DexScreenerResponse, wl *[]Whitelist) *Pair {
	if len(data.Pairs) == 0 {
		return &Pair{}
	}

	maxPair := &Pair{Volume: Volume{H24: 0}}
	for _, pair := range data.Pairs {
		// do not add pair with volume < 1M
		if pair.Volume.H24 < 1000000 {
			continue
		}

		if !checkPair(wl, pair.ChainID, pair.DexID) { // check if current pair located in our set of pools
			continue
		}

		if pair.Volume.H24 > maxPair.Volume.H24 { //&& pair.Volume.H24 > maxPair.Volume.H24 {
			maxPair = &pair
		}
	}

	utils.ClearConsole()
	log.Printf("Selected: %s, %s", maxPair.ChainID, maxPair.DexID)

	return maxPair
}

func checkPair(wl *[]Whitelist, network string, pool string) bool {
	for i := 0; i < len(*wl); i++ {
		if (*wl)[0].Network == network {
			if len((*wl)[i].Pools) == 0 {
				return true
			}
			for j := 0; j < len((*wl)[i].Pools); j++ {
				if (*wl)[i].Pools[j] == pool {
					return true
				}
			}
		}
	}

	return false
}

func normalizePair(pair *Pair, symbols string) *PairNormalized {
	normalized := &PairNormalized{
		Pair:              symbols,
		PairAddress:       pair.PairAddress,
		BaseTokenAddress:  pair.BaseToken.Address,
		QuoteTokenAddress: pair.QuoteToken.Address,
		Network:           pair.ChainID,
		Pull:              pair.DexID,
		URL:               pair.URL,
	}

	return normalized
}

func savePairs(pairs *[]PairNormalized) {

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

func printReceivedToken(counter int, token string) {
	log.Printf("%d) Token %s received", counter, token)
}
