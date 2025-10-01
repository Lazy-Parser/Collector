package worker_internal

import (
	"context"
	"fmt"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/chains"
	"github.com/Lazy-Parser/Collector/market"
)

type MexcWorker struct {
	api    api.MexcAPI
	chains *chains.Chains
}

func NewMexcWorker(api api.MexcAPI, chains *chains.Chains) *MexcWorker {
	return &MexcWorker{api: api, chains: chains}
}

func (mw *MexcWorker) GetAllTokens(ctx context.Context) ([]market.MexcAsset, error) {
	var res []market.MexcAsset
	tokens, err := mw.api.FetchCurrencyInformation(ctx)
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

		// normalize network name
		normalizedNetwork, ok := mw.chains.Select(tokenInfo.Network).ToBase()
		if !ok {
			// network not supported / not in whitelist
			// logger.Get().Z.Warn().Msgf("failed to change mexc network name to the base. Network: %s", tokenInfo.Network)
			continue
		}

		tokens[i].NetworkList[0].Network = normalizedNetwork

		res = append(res, tokens[i])
	}

	return res, nil
}

func (mw *MexcWorker) GetAllFutures(ctx context.Context) ([]market.MexcContractDetail, error) {
	return mw.api.FetchContractInformation(ctx)
}

func (mw *MexcWorker) FindContractBySymbol(arr *[]market.MexcContractDetail, symbol string) (market.MexcContractDetail, bool) {
	for _, contract := range *arr {
		if contract.BaseCoin == symbol {
			return contract, true
		}
	}

	return market.MexcContractDetail{}, false
}
