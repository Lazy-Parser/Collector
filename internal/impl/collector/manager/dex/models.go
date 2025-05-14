package manager_dex

import (
	"github.com/Lazy-Parser/Collector/internal/database"
	d "github.com/Lazy-Parser/Collector/internal/domain"
)

type ManagerDex struct {
	list []*d.DataSourceDex
	pairs map[string][]database.Pair
}
