package sqlite_custom

import "github.com/Lazy-Parser/Collector/market"

type PoolDB struct {
	ID uint `gorm:"primarykey"`

	BaseTokenID  uint
	BaseToken    TokenDB `gorm:"foreignKey:BaseTokenID"`
	QuoteTokenID uint
	QuoteToken   TokenDB `gorm:"foreignKey:QuoteTokenID"`

	Address string
	Network string
	Pool    string

	Label string
	URL   string
	Type  string
}

type TokenDB struct {
	ID          uint `gorm:"primarykey"`
	Name        string
	Decimal     uint8
	Address     string
	Image_url   string
	WithdrawFee string
	CreateTime  int64
	Network     string
}

func (PoolDB) TableName() string {
	return "pools"
}
func (TokenDB) TableName() string {
	return "tokens"
}

func ToPoolDB(p market.Pool) PoolDB {
	return PoolDB{
		BaseToken:  ToTokenDB(p.Pair.BaseToken),
		QuoteToken: ToTokenDB(p.Pair.QuoteToken),
		Address:    p.Address,
		Network:    p.Network,
		Pool:       p.Pool,
		Label:      p.Label,
		URL:        p.URL,
		Type:       p.Type,
	}
}

func ToPool(m PoolDB) market.Pool {
	return market.Pool{
		Pair: market.Pair{
			BaseToken:  ToToken(m.BaseToken),
			QuoteToken: ToToken(m.QuoteToken),
		},
		Address: m.Address,
		Network: m.Network,
		Pool:    m.Pool,
		Label:   m.Label,
		URL:     m.URL,
		Type:    m.Type,
	}
}

func ToTokenDB(t market.Token) TokenDB {
	return TokenDB{
		Name:       t.Name,
		Network:    t.Network,
		Address:    t.Address,
		Decimal:    t.Decimal,
		CreateTime: t.CreateTime,
		// TODO: add more fields
	}
}

func ToToken(t TokenDB) market.Token {
	return market.Token{
		Name:       t.Name,
		Network:    t.Network,
		Address:    t.Address,
		Decimal:    t.Decimal,
		CreateTime: t.CreateTime,
	}
}
