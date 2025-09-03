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

func (r *pairRepo) GetAll(ctx context.Context) ([]market.Pool, error) {
	var pairsdb []PoolDB
	if err := r.db.WithContext(ctx).Find(&pairsdb).Error; err != nil {
		return nil, err
	}

	pairs := make([]market.Pool, len(pairsdb))
	for i, pdb := range pairsdb {
		pairs[i] = ToPool(pdb)
	}

	return pairs, nil
}

func (r *pairRepo) Get(ctx context.Context, addr string) (market.Pool, error) {
	var pairdb PoolDB
	if err := r.db.WithContext(ctx).Where("address = ?", addr).First(&pairdb).Error; err != nil {
		return market.Pool{}, err
	}

	return ToPool(pairdb), nil
}

// FindOrCreate finds or creates a pair in the database.
// pair's base and quote tokens can be empty.
func (r *pairRepo) FindOrCreate(ctx context.Context, pair market.Pool, baseId, quoteId uint) (uint, error) {
	p := ToPoolDB(pair)
	p.BaseTokenID = baseId
	p.QuoteTokenID = quoteId

	res := r.db.WithContext(ctx).FirstOrCreate(&p, PoolDB{Address: p.Address})
	if res.Error != nil {
		return 0, res.Error
	}

	return p.ID, nil
}

func (r *pairRepo) RemoveAll() error {
	return r.db.Exec("DELETE FROM pair_dbs").Error
}
