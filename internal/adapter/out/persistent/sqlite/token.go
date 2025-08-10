package database

import market "github.com/Lazy-Parser/Collector/internal/domain/market"

func (db *Database) SaveToken(token market.Token) error {
	toSave := toDBToken(token)
	return db.DB.Create(&toSave).Error
}

func (db *Database) SaveOrFindToken(token market.Token) (market.Token, error) {
	return market.Token{}, nil
}

func (db *Database) UpdateDecimal(addr string, decimal uint8) error {
	return nil
}

func (db *Database) GetAllTokens() ([]market.Token, error) {
	var tokens []Token
	resp := db.DB.Find(&tokens)

	// change to entity.Token
	out := make([]market.Token, len(tokens))
	for i, token := range tokens {
		out[i] = toToken(token)
	}

	return out, resp.Error
}
