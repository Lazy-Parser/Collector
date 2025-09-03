package exchange_internal

import (
	"context"
	"fmt"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/market"
)

type Mexc struct {
	api api.MexcAPI
}

func NewMexc(api api.MexcAPI) *Mexc {
	return &Mexc{api: api}
}

func (m *Mexc) Name() string {
	return "Mexc"
}

// filter - is an option that filter assets by depositEnable and withdrawEnable.
// limit = "-1" = no limit
func (m *Mexc) Spots(ctx context.Context, filter bool, limit int) ([]market.Token, error) {
	assets, err := m.api.FetchCurrencyInformation(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spots in %s exchange: %v", m.Name(), err)
	}

	var tokens []market.Token
	for _, asset := range assets {
		// Only for internal tests
		if len(tokens) >= limit && limit != -1 {
			break
		}

		for _, net := range asset.NetworkList {
			if filter {
				if !(net.DepositEnable && net.WithdrawEnable) {
					continue
				}
			}
			tokens = append(tokens, normalizeTokenSpot(asset.Coin, net))
		}
	}

	return tokens, nil
}
func normalizeTokenSpot(name string, info market.MexcNetwork) market.Token {
	return market.Token{
		Name:        name,
		Address:     info.Contract,
		Network:     info.Network,
		Decimal:     0, // unknown,
		WithdrawFee: info.WithdrawFee,
		Withdraw:    info.WithdrawEnable,
		Deposit:     info.DepositEnable,
	}
}

func (m *Mexc) Futures(ctx context.Context) ([]market.Token, error) {
	res, err := m.api.FetchContractInformation(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch futures from %s exchange, reason: %v", m.Name(), err)
	}

	futures := make([]market.Token, 0, len(res))
	for _, detail := range res {
		futures = append(futures, normalizeTokenFuture(detail))
	}

	return futures, nil
}
func normalizeTokenFuture(detail market.MexcContractDetail) market.Token {
	return market.Token{
		Name:        detail.BaseCoin,
		Address:     "",
		Network:     "",
		Decimal:     0,
		WithdrawFee: "",
		Withdraw:    false,
		Deposit:     false,
	}
}
