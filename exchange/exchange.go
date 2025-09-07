package exchange

import (
	"context"

	"github.com/Lazy-Parser/Collector/api"
	exchange_internal "github.com/Lazy-Parser/Collector/internal/exchange"
	"github.com/Lazy-Parser/Collector/market"
)

// Every Exchange will have its own Buffer. Buffer - its just a storage for frequently changed data (volume, deposit / withdraw, ...).
//
// Element for Buffer is [BufferTick].
type Exchange interface {
	Name() string
	Fetch24hTickerStats(ctx context.Context) ([]market.MexcTickerStats, error)
	FetchConfigAll(ctx context.Context) ([]market.MexcAsset, error)
	FetchContractsDetails(ctx context.Context) ([]market.MexcContractDetail, error)
	UpdateBuffer(update market.BufferTickUpdate)
	ListenSpot(ch chan market.MexcSpotTick)
	ListenFutures(ch chan market.MexcFutureTick)
}

func NewMexc(api api.MexcAPI) Exchange {
	return exchange_internal.NewMexc(api)
}

func NewGate() {
	// TODO
}

func NewBitget() {
	// TODO
}
