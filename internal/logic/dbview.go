package logic

import (
	"github.com/Lazy-Parser/Collector/internal/database"
)

func GetDatabaseTokens() ([]database.Token, error) {
	return database.GetDB().TokenService.GetAllTokens()
}

func GetDatabasePairs() []database.Pair {
	return nil
}
