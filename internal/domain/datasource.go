package domain

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// CEX
// Main interface for different datasource implementation. It define how to implement some "class / methods" for stock
type DataSource interface {
	Name() string // MEXC, OKX, ...
	Connect() error
	Subscribe() error
	Run(ctx context.Context, push func(AggregatorPayload)) // main logic. Must call Ping(), must parse message and pass all data to AggregatorPayload for Joiner
	ListenState() <-chan bool
	SetState(bool)
}

// DEX (Pool liquidity)
type DataSourceDex interface {
	Name() string // pancakeswap, ...
	Init(toListen *[]Pair) error
	Connect() error // // pass list of pairs...
	Subscribe() error
	FetchDecimals(ctx context.Context) (map[common.Address]uint8, error)
	Run(ctx context.Context, consumerCh chan PancakeswapV2Responce) // some data
}
