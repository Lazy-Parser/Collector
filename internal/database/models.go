package database

import "gorm.io/gorm"

type Database struct {
	DB        *gorm.DB
	IsInitied bool
}

type Token struct {
	ID       int `gorm:"primaryKey;autoIncrement"`
	Name     string
	Address  string
	Decimals int
}

type Pair struct {
	ID int `gorm:"primaryKey;autoIncrement"`

	BaseTokenID int
	BaseToken   Token `gorm:"foreignKey:BaseTokenID;references:ID"`

	QuoteTokenID int
	QuoteToken   Token `gorm:"foreignKey:QuoteTokenID;references:ID"`

	PairAddress string
	Network     string
	Pool        string
	Label       string
	URL         string
}
