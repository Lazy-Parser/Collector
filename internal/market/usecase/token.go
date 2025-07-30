package market_usecase

import (
	market "Cleopatra/internal/market/entity"
	"Cleopatra/internal/port"
)

type TokenService struct {
	repo   DatabaseTokenRepo
	logger port.Logger
}

func NewTokenService(db DatabaseTokenRepo, logger port.Logger) *TokenService {
	return &TokenService{
		repo:   db,
		logger: logger,
	}
}

func (service *TokenService) Save(token market.Token) error {
	return service.repo.SaveToken(token)
}

func (service *TokenService) GetAll() ([]market.Token, error) {
	return service.repo.GetAllTokens()
}
