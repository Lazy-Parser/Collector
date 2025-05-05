package domain

import "context"

// Main interface for different datasource implementation. It define how to implement some "class / methods" for stock
type DataSource interface {
	Name() string // MEXC, OKX, ...
	Connect() error
	Subscribe() error
	Run(ctx context.Context, push func(AggregatorPayload)) // main logic. Must call Ping(), must parse message and pass all data to AggregatorPayload for Joiner
	ListenState() <- chan bool
	SetState(bool)
}
