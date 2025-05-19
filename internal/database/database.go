package database

import (
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	database *Database
)

func NewConnection() error {
	wd, _ := os.Getwd()
	dbPath := filepath.Join(wd, "store", "collector.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	// Create or update schema
	if err := db.AutoMigrate(&Token{}, &Pair{}); err != nil {
		return err
	}

	tokenService := &TokenService{db: db}
	pairService := &PairService{db: db}
	database = &Database{
		DB:           db,
		IsInitied:    true,
		TokenService: tokenService,
		PairService:  pairService,
	}

	return nil
}

func GetDB() *Database {
	return database
}

// GLOBAL
func (db *Database) GloabalQuery(pair *Pair, token *Token) ([]Pair, error) {
	var pairs []Pair

	queryBuilder := db.DB.Preload("BaseToken").Preload("QuoteToken").Table("pairs").
		Joins("JOIN tokens AS base_token ON base_token.id = pairs.base_token_id").
		Joins("JOIN tokens AS quote_token ON quote_token.id = pairs.quote_token_id")

	if pair.Network != "" {
		queryBuilder = queryBuilder.Where("pairs.network = ?", pair.Network)
	}
	if pair.Pool != "" {
		queryBuilder = queryBuilder.Where("pairs.pool = ?", pair.Pool)
	}
	if pair.Label != "" {
		queryBuilder = queryBuilder.Where("pairs.label = ?", pair.Label)
	}
	if pair.PairAddress != "" {
		queryBuilder = queryBuilder.Where("pairs.pair_address = ?", pair.PairAddress)
	}

	if token.Address != "" {
		queryBuilder = queryBuilder.Where("base_token.address = ? OR quote_token.address = ?", token.Address, token.Address)
	}
	if token.Name != "" {
		queryBuilder = queryBuilder.Where("base_token.name = ? OR quote_token.name = ?", token.Name, token.Name)
	}
	if token.Decimals != 0 {
		queryBuilder = queryBuilder.Where("base_token.decimals = ? OR quote_token.decimals = ?", token.Decimals, token.Decimals)
	}

	res := queryBuilder.Find(&pairs)
	return pairs, res.Error
}

// TOKEN
func (db *TokenService) SaveToken(token *Token) error {
	res := db.db.Create(token)
	return res.Error
}

func (db *TokenService) SaveOrFind(token *Token) (Token, error) {
	var t Token
	res := db.db.FirstOrCreate(&t, token)
	return t, res.Error
}

func (db *TokenService) GetAllTokens() ([]Token, error) {
	var tokens []Token
	res := db.db.Find(&tokens)
	return tokens, res.Error
}

func (db *TokenService) ClearTokens() error {
	// var token Token
	res := db.db.Exec("DELETE FROM `tokens`")
	return res.Error
}

// return only ONE found token
func (db *TokenService) FindTokenByQuery(query *Token) ([]Token, error) {
	var token []Token
	res := db.db.Find(&token, query)
	return token, res.Error
}

// return array of found tokens
func (db *TokenService) FindTokensByQuery(query *Token) ([]Token, error) {
	var tokens []Token
	res := db.db.First(&tokens, query)
	return tokens, res.Error
}

// update token decimals by address
func (db *TokenService) UpdateDecimals(query *Token, decimals uint8) error {
	res := db.db.Model(&Token{}).Where("tokens.address = ?", query.Address).Update("decimals", int(decimals))
	return res.Error
}

// add vault by address
func (db *TokenService) UpdateVault(query *Token, vault string) error {
	res := db.db.Model(&Token{}).Where("tokens.address = ?", query.Address).Update("vault", vault)
	return res.Error
}

// PAIR
func (db *PairService) SavePair(pair *Pair) error {
	res := db.db.Create(pair)
	return res.Error
}

func (db *PairService) SaveOrFind(pair *Pair) (Pair, error) {
	var p Pair
	res := db.db.FirstOrCreate(&p, pair)
	return p, res.Error
}

func (db *PairService) GetAllPairs() ([]Pair, error) {
	var pairs []Pair
	res := db.db.Preload("BaseToken").
		Preload("QuoteToken").
		Find(&pairs)

	return pairs, res.Error
}

// pass NULL if your want to fetch empty fields.
func (db *PairService) GetAllPairsByQuery(query PairQuery) ([]Pair, error) {
	var pairs []Pair

	queryBuilder := db.db.Preload("BaseToken").Preload("QuoteToken")

	if query.Network != "" {
		queryBuilder = queryBuilder.Where("pairs.network = ?", query.Network)
	}

	if query.Pool != "" {
		queryBuilder = queryBuilder.Where("pairs.pool = ?", query.Pool)
	}

	if query.Label != "" {
		if query.Label == "NULL" {
			queryBuilder = queryBuilder.Where(`pairs.label = ""`)
		} else {
			queryBuilder = queryBuilder.Where(`pairs.label = ?`, query.Label)
		}
	}

	if query.PairAddress != "" {
		queryBuilder = queryBuilder.Where("pairs.pair_address = ?", query.PairAddress)
	}

	if query.Limit != 0 {
		queryBuilder = queryBuilder.Limit(query.Limit)
	}

	res := queryBuilder.Find(&pairs)
	return pairs, res.Error
}

func (db *PairService) FindPair(query *Pair) (Pair, error) {
	var pair Pair

	queryBuilder := db.db.Preload("BaseToken").Preload("QuoteToken")

	if query.Network != "" {
		queryBuilder = queryBuilder.Where("pairs.network = ?", query.Network)
	}

	if query.PairAddress != "" {
		queryBuilder = queryBuilder.Where("pairs.pair_address = ?", query.PairAddress)
	}

	if query.Pool != "" {
		queryBuilder = queryBuilder.Where("pairs.pool = ?", query.Pool)
	}

	if query.Label != "" {
		queryBuilder = queryBuilder.Where("pairs.label = ?", query.Label)
	}

	res := queryBuilder.First(&pair)
	return pair, res.Error
}

func (db *PairService) ClearPairs() error {
	// var token Token
	res := db.db.Exec("DELETE FROM `pairs`")
	return res.Error
}
