package database

import market "Cleopatra/internal/market/entity"

type Token struct {
	ID      int    `gorm:"primaryKey;autoIncrement"`
	Name    string `gorm:"column:name"`
	Address string `gorm:"unique;column:address"`
	Decimal int    `gorm:"column:decimals"`

	// From entity.Token
	// Name string
	// Decimal uint8
	// Address string
}

type Pair struct {
	ID int `gorm:"primaryKey;autoIncrement"`

	BaseTokenID int
	BaseToken   Token `gorm:"foreignKey:BaseTokenID;references:ID"`

	QuoteTokenID int
	QuoteToken   Token `gorm:"foreignKey:QuoteTokenID;references:ID"`

	PairAddress string `gorm:"unique;column:pair_address"`
	Network     string `gorm:"column:network"`
	Pool        string `gorm:"column:pool"`
	Label       string `gorm:"column:label"`
	URL         string `gorm:"column:url"`
	Type        string `gorm:"column:type"`

	// FROM entity.Pair
	// BaseToken Token
	// QuoteToken Token

	// Address string
	// Network string
	// Pool string

	// Label string
	// URL string
	// Type string
}

// Mapping methods

// Method toToken returns entity.Token from provided sqlite.Token
func toToken(dbToken Token) market.Token {
	return market.Token{
		Name:    dbToken.Name,
		Address: dbToken.Address,
		Decimal: uint8(dbToken.Decimal),
	}
}

// Method toDBToken returns sqlite.Token from provided entity.Token
func toDBToken(token market.Token) Token {
	// ID IS NOT SPECIFIED!!!
	return Token{
		Name:    token.Name,
		Address: token.Address,
		Decimal: int(token.Decimal),
	}
}

// Method toPair returns entity.Pair from provided sqlite.Pair
func toPair(dbPair Pair) market.Pair {
	return market.Pair{
		BaseToken:  toToken(dbPair.BaseToken),
		QuoteToken: toToken(dbPair.QuoteToken),

		Address: dbPair.PairAddress,
		Network: dbPair.Network,
		Pool:    dbPair.Pool,
		Label:   dbPair.Label,
		URL:     dbPair.URL,
		Type:    dbPair.Type,
	}
}

// Method toDBPair returns sqlite.Pair from provided entity.Pair
func toDBPair(pair market.Pair) Pair {
	return Pair{
		BaseToken:  toDBToken(pair.BaseToken),
		QuoteToken: toDBToken(pair.QuoteToken),

		PairAddress: pair.Address,
		Network:     pair.Network,
		Pool:        pair.Pool,
		Label:       pair.Label,
		URL:         pair.URL,
		Type:        pair.Type,
	}
}
