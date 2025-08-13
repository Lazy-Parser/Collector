package market

import "context"

// Keep interfaces small and focused.
type TokenRepo interface {
	Get(ctx context.Context, addr string) (Token, error)
	GetAll(ctx context.Context) ([]Token, error)
	Save(ctx context.Context, token Token) error
}

type PairRepo interface {
	Get(ctx context.Context, addr string) (Pair, error)
	GetAll(ctx context.Context) ([]Pair, error)
	Save(ctx context.Context, pair Pair) error
}
