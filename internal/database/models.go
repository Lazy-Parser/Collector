package database

import "gorm.io/gorm"

type Database struct {
	DB           *gorm.DB
	IsInitied    bool
	TokenService *TokenService
	PairService  *PairService
}

type PairService struct {
	db *gorm.DB
}

type TokenService struct {
	db *gorm.DB
}

type Token struct {
	ID       int    `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"column:name"`
	Address  string `gorm:"unique;column:address"`
	Decimals int    `gorm:"column:decimals"`
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
}

type PairQuery struct {
	PairAddress string
	Network     string
	Pool        string
	Label       string
}
