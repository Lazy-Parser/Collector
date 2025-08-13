package database

import (
	"fmt"
	"os"

	sqlite_custom "github.com/Lazy-Parser/Collector/internal/adapter/sqlite"
	"github.com/Lazy-Parser/Collector/market"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Option func(*cfg)
type cfg struct {
	autoMigrate bool
}

func WithAutoMigrate() Option {
	return func(c *cfg) { c.autoMigrate = true }
}

// Start initializes a new SQLite database connection with the given path and options.
// Creates a new database file if provided path does not exist.
func Start(dbPath string, opts ...Option) (*gorm.DB, error) {
	if !checkIfFileExists(dbPath) {
		if err := createDatabaseFile(dbPath); err != nil {
			return nil, err
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	var c cfg
	for _, opt := range opts {
		opt(&c)
	}

	if c.autoMigrate {
		if err := db.AutoMigrate(&market.Token{}, &market.Pair{}); err != nil {
			sqlDB, _ := db.DB()
			sqlDB.Close()
			return nil, err
		}
	}

	return db, nil
}

func NewTokenRepo(db *gorm.DB) market.TokenRepo { return sqlite_custom.NewTokenRepo(db) }
func NewPairRepo(db *gorm.DB) market.PairRepo   { return sqlite_custom.NewPairRepo(db) }

func checkIfFileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func createDatabaseFile(dbPath string) error {
	file, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	fmt.Printf("Info: Created database file %s\n", file.Name())
	defer file.Close()

	return nil
}
