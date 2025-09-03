package market

import "context"

// Keep interfaces small and focused.
type TokenRepo interface {
	Get(ctx context.Context, addr string) (Token, error)
	GetAll(ctx context.Context) ([]Token, error)
	FindOrCreate(ctx context.Context, token Token) (uint, error)
	RemoveAll() error
}

type PairRepo interface {
	Get(ctx context.Context, addr string) (Pool, error)
	GetAll(ctx context.Context) ([]Pool, error)
	FindOrCreate(ctx context.Context, pair Pool, baseId, quoteId uint) (uint, error)
	RemoveAll() error
}
