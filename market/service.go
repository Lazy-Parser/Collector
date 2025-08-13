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

func (s *Service) ListPairs(ctx context.Context) ([]Pair, error) {
	return s.pairs.GetAll(ctx)
}
