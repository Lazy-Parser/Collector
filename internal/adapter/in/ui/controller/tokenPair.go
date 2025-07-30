package controller

import (
	market "Cleopatra/internal/market/entity"
	market_usecase "Cleopatra/internal/market/usecase"
)

type TokenPairService struct {
	repo market_usecase.TokenService
}

func NewTokenService(tokenService market_usecase.TokenService) *TokenPairService {
	return &TokenPairService{
		repo: tokenService,
	}
}

func (service *TokenPairService) GetAllTokens() ([]market.Token, error) {
	return service.repo.GetAll()
}
