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

	database = &Database{DB: db, IsInitied: true}

	return nil
}

func GetDB() *Database {
	return database
}

// -------
func (db *Database) SaveToken(token *Token) error {
	res := db.DB.Create(token)
	return res.Error
}

func (db *Database) SavePair(token *Pair) error {
	res := db.DB.Create(token)
	return res.Error
}

// ---------
func (db *Database) GetAllTokens() ([]Token, error) {
	var tokens []Token
	res := db.DB.Find(&tokens)
	return tokens, res.Error
}

func (db *Database) GetAllPairs() ([]Pair, error) {
	var pairs []Pair
	res := db.DB.Find(&pairs)
	return pairs, res.Error
}

// ---------
func (db *Database) ClearTokens() error {
	// var token Token
	res := db.DB.Exec("DELETE FROM `tokens`")
	return res.Error
}

func (db *Database) ClearPairs() error {
	// var token Token
	res := db.DB.Exec("DELETE FROM `pairs`")
	return res.Error
}
