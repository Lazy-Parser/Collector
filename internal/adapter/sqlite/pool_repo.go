package sqlite_custom

import (
	"context"
	"log"

	"github.com/Lazy-Parser/Collector/market"
	"gorm.io/gorm"
)

type poolRepo struct {
	db *gorm.DB
}

func NewPoolRepo(db *gorm.DB) *poolRepo {
	return &poolRepo{db: db}
}

func (r *poolRepo) GetAll(ctx context.Context) ([]market.Pool, error) {
	var pairsdb []PoolDB
	if err := r.db.
		WithContext(ctx).
		Preload("BaseToken").
		Preload("QuoteToken").
		Find(&pairsdb).Error; err != nil {
		return nil, err
	}

	log.Printf("Pools from db: %+v", pairsdb)

	pairs := make([]market.Pool, len(pairsdb))
	for i, pdb := range pairsdb {
		pairs[i] = ToPool(pdb)
	}

	return pairs, nil
}

func (r *poolRepo) Get(ctx context.Context, addr string) (market.Pool, error) {
	var pairdb PoolDB
	if err := r.db.WithContext(ctx).Where("address = ?", addr).First(&pairdb).Error; err != nil {
		return market.Pool{}, err
	}

	return ToPool(pairdb), nil
}

// FindOrCreate finds or creates a pair in the database.
// pair's base and quote tokens can be empty.
func (r *poolRepo) FindOrCreate(ctx context.Context, pool market.Pool, baseId, quoteId uint) (uint, error) {
	p := ToPoolDB(pool)
	p.BaseTokenID = baseId
	p.QuoteTokenID = quoteId
	// base tokens empty, because otherwise gorm will save them (but its supposed that tokens are already saved)
	p.BaseToken = TokenDB{}
	p.QuoteToken = TokenDB{}

	res := r.db.WithContext(ctx).FirstOrCreate(&p, PoolDB{Address: p.Address})
	if res.Error != nil {
		return 0, res.Error
	}

	return p.ID, nil
}

func (r *poolRepo) RemoveAll() error {
	return r.db.Exec("DELETE FROM pools").Error
}

func (r *poolRepo) GetAllRepos() ([]string, error) {
	return r.db.Migrator().GetTables()
}
