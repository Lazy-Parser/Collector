package worker_mexc

import (
	config "github.com/Lazy-Parser/Collector/config/service"
	api_mexc "github.com/Lazy-Parser/Collector/internal/api/mexc"
	"github.com/Lazy-Parser/Collector/internal/common/chains"
	logger "github.com/Lazy-Parser/Collector/internal/common/zerolog"
	"context"
	"fmt"
)

type MexcWorker struct {
	mexcApi *api_mexc.MexcAPI
	cfg     *config.Config
	chains  *chains.Chains
}

func NewWorker(mexcApi *api_mexc.MexcAPI, cfg *config.Config, chains *chains.Chains) *MexcWorker {
	return &MexcWorker{mexcApi: mexcApi, cfg: cfg, chains: chains}
}

// Function GetAllToken fetches all coins on Mexc (with additional info like contract, network, withdraw, deposit). It also change network names to the global and remove tokens, which networks are unsupported
func (mexc *MexcWorker) GetAllTokens(ctx context.Context) ([]api_mexc.Asset, error) {
	var res []api_mexc.Asset
	tokens, err := mexc.mexcApi.FetchCurrencyInformation(ctx, mexc.cfg)
	if err != nil {
		return nil, fmt.Errorf("fetch all tokens from mexc: %v", err)
	}

	// Mexc has its own network names. For example: solana - SOL, ethereum - ETH.
	// So we just change mexc network names to the global ones
	for i := range len(tokens) {
		if len(tokens[i].NetworkList) == 0 {
			continue
		}
		tokenInfo := tokens[i].NetworkList[0] // TODO: its not the first one always.

		// Deposit and Withdraw must be true
		if !(tokenInfo.DepositEnable && tokenInfo.WithdrawEnable) {
			continue
		}

		// check if network supported
		if ok := mexc.chains.IsWhitelist(tokenInfo.Network); !ok {
			continue
		}

		// normalize network name
		normalizedNetwork, ok := mexc.chains.Select(tokenInfo.Network).ToBase()
		if !ok {
			// the error in the toBase() method cannot be, because of the checking earlier. But to make shure, we can check it one more time.
			logger.Get().Z.Warn().Msgf("failed to change mexc network name to the base. Network: %s", tokenInfo.Network)
			continue
		}

		tokens[i].NetworkList[0].Network = normalizedNetwork

		res = append(res, tokens[i])
	}

	return res, nil
}

// Function GetAllFutures fetches all coins on futures. From futures we can get only token name, so we need to compare all tokens with futures tokens and pick only appropriate ones
func (mexc *MexcWorker) GetAllFutures(ctx context.Context) ([]api_mexc.ContractDetail, error) {
	return mexc.mexcApi.FetchContractInformation(ctx, mexc.cfg.Mexc.API.CONTRACTS_DETAIL)

}

// second return type (bool) means could this func find token by symbol
func (mexc *MexcWorker) FindContractBySymbol(arr *[]api_mexc.ContractDetail, symbol string) (api_mexc.ContractDetail, bool) {
	for _, contract := range *arr {
		if contract.BaseCoin == symbol {
			return contract, true
		}
	}

	return api_mexc.ContractDetail{}, false
}

// import
// (
// 	config "github.com/Lazy-Parser/Collector/config/service"
// 	"github.com/Lazy-Parser/Collector/internal/adapter/out/webapi/chains"
// 	market "github.com/Lazy-Parser/Collector/internal/market/entity"
// 	"context"
// 	"crypto/hmac"
// 	"crypto/sha256"
// 	"encoding/hex"
// 	"fmt"
// 	"net/url"
// 	"time"

// 	"github.com/go-resty/resty/v2"
// )

// type Mexc struct {
// }

// func NewMexc(chainsService *chains.Chains) *Mexc {
// 	return &Mexc{}
// }

// // First,
// func (mexc *Mexc) GetFutures(ctx context.Context, cfg *config.Config, chainsService *chains.Chains) ([]market.Token, error) {
// 	allCoins, err := mexc.FetchAllCoins(ctx, cfg)
// 	if err != nil {
// 		return nil, fmt.Errorf("get all coins from mexc: %v", err)
// 	}

//
// 	futures, err := mexc.FetchFuturesCoins(ctx, cfg)
// 	if err != nil {
// 		return nil, fmt.Errorf("get futures from mexc: %v", err)
// 	}

// 	// compare all coins and futures to pick appropriate tokens
// 	var res []market.Token
// 	for _, coin := range allCoins {
// 		coinAddInfo := coin.NetworkList[0]

// 		// DEPOSIT && WITHDRAW CHECK. Deposit and Withdraw must be true
// 		if !(coinAddInfo.DepositEnable && coinAddInfo.WithdrawEnable) {
// 			// Pass this token
// 			continue
// 		}

// 		// NETWORK WHITELIST
// 		if !chainsService.IsWhitelist(coinAddInfo.Network) {
// 			// current network is not whitelisted
// 			continue
// 		}

// 		// TRY TO FIND TOKEN IN FUTURES LIST
// 		contract, ok := findContractBySymbol(&futures, coin.Coin)
// 		if !ok {
// 			// not found
// 			continue
// 		}

// 		// save
// 		globalNetwork, ok := chainsService.Select(coinAddInfo.Network).ToBase()
// 		if !ok {
// 			fmt.Printf("Failed to find global network name from mexc for network: %s", coinAddInfo.Network)
// 		}
// 		res = append(res, market.Token{
// 			Name:        coin.Coin,
// 			Decimal:     0, // unknown, will find it later
// 			Address:     coinAddInfo.Contract,
// 			Image_url:   contract.ImageUrl,
// 			WithdrawFee: coinAddInfo.WithdrawFee,
// 			Network:     globalNetwork,
// 			CreateTime:  contract.CreateTime,
// 		})
// 	}

// 	return res, nil
// }

// // Method FetchAllCoins fetches all coins from Mexc with additional info (withdraw, contract, ...)
// func (mexc *Mexc) FetchAllCoins(ctx context.Context, cfg *config.Config) ([]Asset, error) {
// 	qs := url.Values{"timestamp": {fmt.Sprint(time.Now().UnixMilli())}}
// 	mac := hmac.New(sha256.New, []byte(cfg.Mexc.PRIVATE_TOKEN))
// 	mac.Write([]byte(qs.Encode()))
// 	qs.Set("signature", hex.EncodeToString(mac.Sum(nil)))

// 	var res []Asset
// 	_, err := resty.New().
// 		R().
// 		SetContext(ctx).
// 		SetQueryString(qs.Encode()).
// 		SetResult(&res).
// 		SetHeader("X-MEXC-APIKEY", cfg.Mexc.ACCESS_TOKEN).
// 		Get(cfg.Mexc.API.CONFIG_GETALL)
// 	return res, err
// }

// func (mexc *Mexc) FetchFuturesCoins(ctx context.Context, cfg *config.Config) ([]ContractDetail, error) {
// 	var res Contracts
// 	_, err := resty.New().
// 		R().
// 		SetContext(ctx).
// 		SetResult(&res).
// 		Get(cfg.Mexc.API.CONTRACTS_DETAIL)
// 	return res.Data, err
// }
