// TODO: creating a buffer arr to then cast from it is not memory efficient. So its better to stream rows
package sqlite_custom

import (
	"context"

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
	var tokensdb []TokenDB
	if err := r.db.Find(&tokensdb).Error; err != nil {
		return nil, err
	}

	// cast
	tokens := make([]market.Token, len(tokensdb))
	for i, t := range tokensdb {
		tokens[i] = ToToken(t)
	}

	return tokens, nil
}

func (r *tokenRepo) Get(ctx context.Context, addr string) (market.Token, error) {
	var tokendb TokenDB
	if err := r.db.WithContext(ctx).Where("address = ?", addr).First(&tokendb).Error; err != nil {
		return market.Token{}, err
	}

	return ToToken(tokendb), nil
}

// Update if already exists (by address), insert otherwise
// Returns the ID of the saved token and error
// ID returns always: if token already exists, or inserted
func (r *tokenRepo) FindOrCreate(ctx context.Context, token market.Token) (uint, error) {
	t := ToTokenDB(token)

	// res := r.db.WithContext(ctx).FirstOrCreate(&t, TokenDB{Address: t.Address})
	res := r.db.WithContext(ctx).Create(&t)
	if res.Error != nil {
		return 0, res.Error
	}

	return t.ID, nil
}

func (r *tokenRepo) RemoveAll() error {
	return r.db.Exec("DELETE FROM token_dbs").Error
}
