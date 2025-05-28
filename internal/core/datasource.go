package core

import (
	"context"

	db "github.com/Lazy-Parser/Collector/internal/database"
)

// CEX
// Main interface for different datasource implementation. It define how to implement some "class / methods" for stock
type DataSource interface {
	Name() string // MEXC, OKX, ...
	Connect() error
	Subscribe(ctx context.Context, pairs []db.Pair) error
	Run(ctx context.Context, consumerChan chan MexcResponse) // main logic. Must call Ping(), must parse message and pass all data to AggregatorPayload for Joiner
}

// DEX (Pool liquidity)
type DataSourceDex interface {
	Name() string // pancakeswap, ...
	//Init(toListen *[]db.Pair) error
	Init() error
	Connect() error // // pass list of pairs...
	Subscribe() error
	Run(ctx context.Context, consumerCh chan CollectorDexResponse) // some data
}

type MetadataCollector interface {
	PushPairs(pairs *[]db.Pair)
	FetchMetadata() (Metadata, error)
}
