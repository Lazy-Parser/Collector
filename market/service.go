package market

import "context"

type Service struct {
    tokens TokenRepo
    pairs  PairRepo
}

func NewService(t TokenRepo, p PairRepo) *Service {
    return &Service{tokens: t, pairs: p}
}

func (s *Service) ListTokens(ctx context.Context) ([]Token, error) {
    return s.tokens.GetAll(ctx)
}

func (s *Service) PairsFor(ctx context.Context, base string) ([]Pair, error) {
    return s.pairs.ByBase(ctx, base)
}
