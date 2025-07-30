package market_usecase

import market "Cleopatra/internal/market/entity"

type DatabasePairRepo interface {
	SavePair(pair market.Pair) error
	GetAllPairs() ([]market.Pair, error)
}

type DatabaseTokenRepo interface {
	SaveToken(token market.Token) error
	GetAllTokens() ([]market.Token, error)
}
