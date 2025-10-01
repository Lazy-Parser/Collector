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
	StopAll(ctx context.Context)
	BufferLoop(ctx context.Context) error
	ListenSpot(ctx context.Context, ch chan *market.MexcSpotTick) error
	ListenFutures(ctx context.Context, ch chan *market.MexcFutureTick) error
}

func NewMexc(api api.MexcAPI) (Exchange, error) {
	return exchange_internal.NewMexc(api)
}

func NewGate() {
	// TODO
}

func NewBitget() {
	// TODO
}
