package manager_dex

import (
	core "github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"math/big"
)

type ManagerDex struct {
	list       []*core.DataSourceDex
	pairs      map[string][]database.Pair
	quotePairs map[string]*big.Float
}
