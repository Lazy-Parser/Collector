package logic

import (
	"github.com/Lazy-Parser/Collector/internal/core"
)

func GetDatabaseTokens() []core.Token {
	tokens := []core.Token{
		{
			Name:    "BTC",
			Decimal: 8,
			Address: "0xsjkdhgf34yrsudihfuk2g3er98ydf",
		},
		{
			Name:    "ETH",
			Decimal: 16,
			Address: "0x548t9eoiuhrg9834hfeduhfsdjfn",
		},
		{
			Name:    "USDT",
			Decimal: 6,
			Address: "0x3485yn3v495yn3984cv6903vy46wr",
		},
	}

	return tokens
}
