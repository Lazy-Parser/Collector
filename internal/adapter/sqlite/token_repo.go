package sqlite_custom

import (
	"context"
	"time"

	"github.com/Lazy-Parser/Collector/market"
	"gorm.io/gorm"
)

type tokenRepo struct {
	db *gorm.DB
}

func NewTokenRepo(db *gorm.DB) *tokenRepo {
	return &tokenRepo{db: db}
}

func (r *tokenRepo) GetAll(ctx context.Context) ([]market.Token, error) {
	var tokens []market.Token
	// if err := r.db.Find(&tokens).Error; err != nil {
	// 	return nil, err
	// }

	tokens = []market.Token{
		market.Token{
			Name:        "BTC",
			Address:     "0xjhwefi8349fhief",
			Decimal:     16,
			Network:     "Solana",
			CreateTime:  time.Now().UnixMilli() - time.Hour.Milliseconds()*5, // now - 5 hours
			WithdrawFee: "10",
		},
		market.Token{
			Name:        "ETH",
			Address:     "0xCNOUhuefhuio4f",
			Decimal:     8,
			Network:     "BCS",
			CreateTime:  time.Now().UnixMilli(),
			WithdrawFee: "10",
		},
	}

	return tokens, nil
}

func (r *tokenRepo) Get(ctx context.Context, addr string) (market.Token, error) {
	var token market.Token
	// if err := r.db.WithContext(ctx).Where("address = ?", addr).First(&token).Error; err != nil {
	// 	return market.Token{}, err
	// }
	return token, nil
}

func (r *tokenRepo) Save(ctx context.Context, token market.Token) error {
	// if err := r.db.WithContext(ctx).Save(&token).Error; err != nil {
	// 	return err
	// }
	return nil
}
