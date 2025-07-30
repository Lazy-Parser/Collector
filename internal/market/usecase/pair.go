package market_usecase

import (
	market "Cleopatra/internal/market/entity"
	"Cleopatra/internal/port"
)

type PairService struct {
	repo   DatabasePairRepo
	logger port.Logger
}

func NewPairService(db DatabasePairRepo, logger port.Logger) *PairService {
	return &PairService{
		repo:   db,
		logger: logger,
	}
}

func (service *PairService) Save(pair market.Pair) error {
	return service.repo.SavePair(pair)
}

func (service *PairService) GetAll() ([]market.Pair, error) {
	return service.repo.GetAllPairs()
}
