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
	var pairsdb []PairDB
	if err := r.db.WithContext(ctx).Find(&pairsdb).Error; err != nil {
		return nil, err
	}

	pairs := make([]market.Pair, len(pairsdb))
	for i, pdb := range pairsdb {
		pairs[i] = ToPair(pdb)
	}

	return pairs, nil
}

func (r *pairRepo) Get(ctx context.Context, addr string) (market.Pair, error) {
	var pairdb PairDB
	if err := r.db.WithContext(ctx).Where("address = ?", addr).First(&pairdb).Error; err != nil {
		return market.Pair{}, err
	}

	return ToPair(pairdb), nil
}

func (r *pairRepo) Save(ctx context.Context, pair market.Pair) error {
	pairdb := ToPairDB(pair)
	if err := r.db.WithContext(ctx).Save(&pairdb).Error; err != nil {
		return err
	}
	return nil
}
