package exchange

import (
	"github.com/Lazy-Parser/Collector/chains"
	"github.com/Lazy-Parser/Collector/market"
)

// Compare - returns an array of futures with all info that exchange can give.
//
// Problem: You fetched all tokens from exchange api (it provides: name (symbol), contract, deposit, withdraw, ...).
// Then fetched all futures tokens. But this endpoint gives you only tokens names.
// So to get futures tokens with all info, you need to filter spots tokens (from the 1 request) with tokens from the futures (second requests) and return that matched
func CompareFutures(tokens *[]market.Token, futures *[]market.Token) []market.Token {
	var res []market.Token
	for _, token := range *tokens {
		_, ok := findToken(token.Name, futures)
		if !ok {
			continue
		}

		res = append(res, token)
	}

	return res
}
func findToken(name string, tokens *[]market.Token) (market.Token, bool) {
	for idx := range *tokens {
		if (*tokens)[idx].Name == name {
			return (*tokens)[idx], true
		}
	}

	return market.Token{}, false
}

// Return an array of filtered tokens by allowed networks. Allowed networks are in chainsService
func FilterByNetworks(tokens *[]market.Token, chainsService *chains.Chains) []market.Token {
	var res []market.Token
	for _, token := range *tokens {
		if ok := chainsService.IsWhitelist(token.Network); ok {
			res = append(res, token)
		}
	}

	return res
}
