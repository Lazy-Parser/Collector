package exchange

import (
	"context"

	"github.com/Lazy-Parser/Collector/api"
	exchange_internal "github.com/Lazy-Parser/Collector/internal/exchange"
	"github.com/Lazy-Parser/Collector/market"
)

type Exchange interface {
	Name() string
	Spots(ctx context.Context, filter bool, limit int) ([]market.Token, error)
	Futures(ctx context.Context) ([]market.Token, error)
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
