package manager_dex

import (
	d "github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
)

type ManagerDex struct {
	list  []*d.DataSourceDex
	pairs map[string][]database.Pair
}
