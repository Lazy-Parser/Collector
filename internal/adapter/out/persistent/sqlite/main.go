package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

// Function NewConnection returns Database struct to work with local database. dbPath is a path where to store database
func NewConnection(dbPath string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Create or update schema
	if err := db.AutoMigrate(&Token{}, &Pair{}); err != nil {
		return nil, err
	}

	database := &Database{
		DB: db,
	}

	return database, nil
}
