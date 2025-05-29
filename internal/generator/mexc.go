package generator

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	d "github.com/Lazy-Parser/Collector/internal/core"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

var (
	stream    chan Asset
	assets    []Asset
	contracts []ContractDetail
	store     []Asset
)

func MexcInit() error {
	ctx := context.Background()
	stream = make(chan Asset, 5000)

	// remove in prod
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	// remove in prod

	// load tokens from https://api.mexc.com/api/v3/capital/config/getall
	accessToken := os.Getenv("MEXC_ACCESS_TOKEN")
	privateToken := os.Getenv("MEXC_PRIVATE_TOKEN")
	assets, err = getAssets(ctx, accessToken, privateToken)
	if err != nil {
		return fmt.Errorf("fetching all tokens from MEXC [generator/mexc.go/getAssets]. %v", err)
	}
	log.Println("All tokens loaded from mexc succsessful!")

	// load tokens only from mexc
	contracts, err = getContractsDetail(ctx)
	if err != nil {
		return fmt.Errorf("fetching pairs from futures / MEXC [generator/mexc.go/getContractsDetail]. %v", err)
	}
	log.Println("All pairs loaded from futures / mexc succsessful!")

	return nil
}

// compare tokens from mexc and return only tokens from futures with addressed. Also filter tokens based on whitelist
func MexcCompare(wl *[]d.Whitelist) {
	for _, asset := range assets {
		if asset.containsContract() && asset.containsWhitelist(wl) {
			store = append(store, asset)
		}
	}
}

func MexcGetTokens() []Asset {
	return store
}

// check if current asset (coin) is future (contains some token from futures list (contracts))
func (a *Asset) containsContract() bool {
	for _, c := range contracts {
		if len(a.NetworkList[0].Contract) == 0 {
			continue
		}

		if a.Coin == c.BaseCoin {
			return true
		}
	}

	return false
}

func (a *Asset) containsWhitelist(wl *[]d.Whitelist) bool {
	for _, allowed := range *wl {
		for idx, n := range a.NetworkList {
			if allowed.NetworkShort == n.NetworkShort {
				// add full name of network
				a.NetworkList[idx].Network = allowed.Network
				return true
			}
		}
	}

	return false
}

// get all coins (spot / futures) with addresses
func getAssets(ctx context.Context, key, secret string) ([]Asset, error) {
	qs := url.Values{"timestamp": {fmt.Sprint(time.Now().UnixMilli())}}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(qs.Encode()))
	qs.Set("signature", hex.EncodeToString(mac.Sum(nil)))

	var res []Asset
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetQueryString(qs.Encode()).
		SetResult(&res).
		SetHeader("X-MEXC-APIKEY", key).
		Get("https://api.mexc.com/api/v3/capital/config/getall")
	return res, err
}

// get all coins from futures
func getContractsDetail(ctx context.Context) ([]ContractDetail, error) {
	var res Contracts
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetResult(&res).
		Get("https://contract.mexc.com/api/v1/contract/detail")
	return res.Data, err
}
