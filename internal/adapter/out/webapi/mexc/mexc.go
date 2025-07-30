package mexc

import (
	config "Cleopatra/config/service"
	"Cleopatra/internal/adapter/out/webapi/chains"
	market "Cleopatra/internal/market/entity"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

type Mexc struct {
}

func NewMexc(chainsService *chains.Chains) *Mexc {
	return &Mexc{}
}

// First, fetch all coins on Mexc (with additional info like contract, network, withdraw, deposit)
func (mexc *Mexc) GetFutures(ctx context.Context, cfg *config.Config, chainsService *chains.Chains) ([]market.Token, error) {
	allCoins, err := mexc.FetchAllCoins(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("get all coins from mexc: %v", err)
	}

	// Fetch all coins on futures. From futures we can get only token name, so we then we need to compare all tokens with futures tokens and pick only appropriate ones
	futures, err := mexc.FetchFuturesCoins(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("get futures from mexc: %v", err)
	}

	// compare all coins and futures to pick appropriate tokens
	var res []market.Token
	for _, coin := range allCoins {
		coinAddInfo := coin.NetworkList[0]

		// DEPOSIT && WITHDRAW CHECK. Deposit and Withdraw must be true
		if !(coinAddInfo.DepositEnable && coinAddInfo.WithdrawEnable) {
			// Pass this token
			continue
		}

		// NETWORK WHITELIST
		if !chainsService.IsWhitelist(coinAddInfo.Network) {
			// current network is not whitelisted
			continue
		}

		// TRY TO FIND TOKEN IN FUTURES LIST
		contract, ok := findContractBySymbol(&futures, coin.Coin)
		if !ok {
			// not found
			continue
		}

		// save
		globalNetwork, ok := chainsService.Select(coinAddInfo.Network).ToBase()
		if !ok {
			fmt.Printf("Failed to find global network name from mexc for network: %s", coinAddInfo.Network)
		}
		res = append(res, market.Token{
			Name:        coin.Coin,
			Decimal:     0, // unknown, will find it later
			Address:     coinAddInfo.Contract,
			Image_url:   contract.ImageUrl,
			WithdrawFee: coinAddInfo.WithdrawFee,
			Network:     globalNetwork,
			CreateTime:  contract.CreateTime,
		})
	}

	return res, nil
}

// Method FetchAllCoins fetches all coins from Mexc with additional info (withdraw, contract, ...)
func (mexc *Mexc) FetchAllCoins(ctx context.Context, cfg *config.Config) ([]Asset, error) {
	qs := url.Values{"timestamp": {fmt.Sprint(time.Now().UnixMilli())}}
	mac := hmac.New(sha256.New, []byte(cfg.Mexc.PRIVATE_TOKEN))
	mac.Write([]byte(qs.Encode()))
	qs.Set("signature", hex.EncodeToString(mac.Sum(nil)))

	var res []Asset
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetQueryString(qs.Encode()).
		SetResult(&res).
		SetHeader("X-MEXC-APIKEY", cfg.Mexc.ACCESS_TOKEN).
		Get(cfg.Mexc.API.CONFIG_GETALL)
	return res, err
}

func (mexc *Mexc) FetchFuturesCoins(ctx context.Context, cfg *config.Config) ([]ContractDetail, error) {
	var res Contracts
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetResult(&res).
		Get(cfg.Mexc.API.CONTRACTS_DETAIL)
	return res.Data, err
}
