package exchange_internal

import (
	"context"
	"sync"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/market"
)

type Mexc struct {
	mu     sync.RWMutex
	api    api.MexcAPI
	buffer []market.BufferTick
	// TODO: websocket conn
}

func NewMexc(api api.MexcAPI) *Mexc {
	return &Mexc{api: api}
}

func (m *Mexc) Name() string {
	return "Mexc"
}

func (m *Mexc) Fetch24hTickerStats(ctx context.Context) ([]market.MexcTickerStats, error) {
	return m.api.Fetch24hTickerStats(ctx)
}

func (m *Mexc) FetchConfigAll(ctx context.Context) ([]market.MexcAsset, error) {
	return m.api.FetchCurrencyInformation(ctx)
}

func (m *Mexc) FetchContractsDetails(ctx context.Context) ([]market.MexcContractDetail, error) {
	return m.api.FetchContractInformation(ctx)
}

func (m *Mexc) UpdateBuffer(update market.BufferTickUpdate) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := m.findInBuffer(update.Name)
	if idx == -1 {
		m.appendTick(update)
		return
	}

	m.updateTickIdx(idx, update)
}

// private
func (m *Mexc) findInBuffer(name string) int {
	for i, tick := range m.buffer {
		if tick.Name == name {
			return i
		}
	}
	return -1
}
func (m *Mexc) updateTickIdx(idx int, update market.BufferTickUpdate) {
	if update.Withdraw != nil {
		m.buffer[idx].Withdraw = *update.Withdraw
	}
	if update.WithdrawFee != nil {
		m.buffer[idx].WithdrawFee = *update.WithdrawFee
	}
	if update.Deposit != nil {
		m.buffer[idx].Deposit = *update.Deposit
	}
	if update.Volume != nil {
		m.buffer[idx].Volume = *update.Volume
	}
}
func (m *Mexc) appendTick(tick market.BufferTickUpdate) {
	// BufferTickUpdate cast to the BufferTick
	t := market.BufferTick{}
	t.Name = tick.Name // name is mandatory
	if tick.Withdraw != nil {
		t.Withdraw = *tick.Withdraw
	}
	if tick.WithdrawFee != nil {
		t.WithdrawFee = *tick.WithdrawFee
	}
	if tick.Deposit != nil {
		t.Deposit = *tick.Deposit
	}
	if tick.Volume != nil {
		t.Volume = *tick.Volume
	}

	m.buffer = append(m.buffer, t)
}

func (m *Mexc) ListenSpot(ch chan market.MexcSpotTick) {

}

func (m *Mexc) ListenFutures(ch chan market.MexcFutureTick) {

}


// filter - is an option that filter assets by depositEnable and withdrawEnable.
// limit = "-1" = no limit
// func (m *Mexc) Spots(ctx context.Context, filter bool, limit int) ([]market.Token, error) {
// 	assets, err := m.api.FetchCurrencyInformation(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch spots in %s exchange: %v", m.Name(), err)
// 	}

// 	var tokens []market.Token
// 	for _, asset := range assets {
// 		// Only for internal tests
// 		if len(tokens) >= limit && limit != -1 {
// 			break
// 		}

// 		for _, net := range asset.NetworkList {
// 			if filter {
// 				if !(net.DepositEnable && net.WithdrawEnable) {
// 					continue
// 				}
// 			}
// 			tokens = append(tokens, normalizeTokenSpot(asset.Coin, net))
// 		}
// 	}

// 	return tokens, nil
// }
// func normalizeTokenSpot(name string, info market.MexcNetwork) market.Token {
// 	return market.Token{
// 		Name:        name,
// 		Address:     info.Contract,
// 		Network:     info.Network,
// 		Decimal:     0, // unknown,
// 		WithdrawFee: info.WithdrawFee,
// 		Withdraw:    info.WithdrawEnable,
// 		Deposit:     info.DepositEnable,
// 	}
// }

// func (m *Mexc) Futures(ctx context.Context) ([]market.Token, error) {
// 	res, err := m.api.FetchContractInformation(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch futures from %s exchange, reason: %v", m.Name(), err)
// 	}

// 	futures := make([]market.Token, 0, len(res))
// 	for _, detail := range res {
// 		futures = append(futures, normalizeTokenFuture(detail))
// 	}

// 	return futures, nil
// }
// func normalizeTokenFuture(detail market.MexcContractDetail) market.Token {
// 	return market.Token{
// 		Name:        detail.BaseCoin,
// 		Address:     "",
// 		Network:     "",
// 		Decimal:     0,
// 		WithdrawFee: "",
// 		Withdraw:    false,
// 		Deposit:     false,
// 	}
// }
