package sqlite_custom

import (
	"context"

	"github.com/Lazy-Parser/Collector/market"
	"gorm.io/gorm"
)

type pairRepo struct {
	db *gorm.DB
}

func NewPairRepo(db *gorm.DB) *pairRepo {
	return &pairRepo{db: db}
}

func (r *pairRepo) GetAll(ctx context.Context) ([]market.Pair, error) {
	var pairs []market.Pair
	// if err := r.db.Find(&tokens).Error; err != nil {
	// 	return nil, err
	// }
	return pairs, nil
}

func (r *pairRepo) Get(ctx context.Context, addr string) (market.Pair, error) {
	var pair market.Pair
	// if err := r.db.WithContext(ctx).Where("address = ?", addr).First(&token).Error; err != nil {
	// 	return market.Token{}, err
	// }
	return pair, nil
}

func (r *pairRepo) Save(ctx context.Context, pair market.Pair) error {
	// if err := r.db.WithContext(ctx).Save(&token).Error; err != nil {
	// 	return err
	// }
	return nil
}
