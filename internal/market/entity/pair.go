package market

type Pair struct {
	BaseToken  Token
	QuoteToken Token

	Address string
	Network string
	Pool    string

	Label string
	URL   string
	Type  string
}

// dexscreener.PairDS -> market.PairCandidat -> market.Pair
type PairCandidat struct {
	BaseToken  Token
	QuoteToken Token

	Address   string
	Network   string
	Pool      string
	Volume    float64
	Liquidity Liquidity

	Label string
	URL   string
	Type  string
}

type Liquidity struct {
	USD   float64
	Base  float64
	Quote float64
}

// Methods
func PairCandidatToPair(candidat PairCandidat) Pair {
	return Pair{
		BaseToken: Token{
			Name:    candidat.BaseToken.Name,
			Address: candidat.BaseToken.Address,
			Decimal: candidat.BaseToken.Decimal,
		},
		QuoteToken: Token{
			Name:    candidat.QuoteToken.Name,
			Address: candidat.QuoteToken.Address,
			Decimal: candidat.QuoteToken.Decimal,
		},
		Address: candidat.Address,
		Network: candidat.Network,
		Pool:    candidat.Pool,

		Label: candidat.Label,
		URL:   candidat.URL,
	}
}

// Function SelectBest selects PairCandidat with the biggest Volume. Return false if all Volumes in Pairs in 'data' < 'minVolume' (default 70k$)
func SelectBest(data *[]PairCandidat, minVolume float64) (Pair, bool) {
	bestToken := PairCandidat{Volume: -1} // create empty result
	ok := false

	//var curQuoteSymbol string
	var curVolume24 float64

	for _, pair := range *data {
		//curQuoteSymbol = pair.QuoteToken.Symbol
		curVolume24 = pair.Volume

		// filter by quote token. Only SOL, USDC, USDT allowed
		//if curQuoteSymbol != "SOL" &&
		//	curQuoteSymbol != "USDC" &&
		//	curQuoteSymbol != "USDT" &&
		//	curQuoteSymbol != "WBNB" {
		//	continue
		//}

		// filter by volume
		if curVolume24 < minVolume {
			continue
		}

		// TODO: maybe add filter by liquidity
		// filter by liquidity
		// if pair.Liquidity.USD < 10000 {
		// 	continue
		// }

		// TODO: add filter by allowed pools!!!!! VERY IMPORTANT

		// select pair with the biggest volume
		if curVolume24 > bestToken.Volume {
			bestToken = pair
			ok = true
		}
	}

	// mapping PairCandidat -> Pair
	return PairCandidatToPair(bestToken), ok
}

// // Method Compare compares provided Pair 'a' with 'this' ('p') pair according to their 'Address' fields. Returns TRUE if the addressess are the same. Register (a or A) does not matter
// func (p *Pair) Compare(a Pair) bool {
// 	return strings.ToLower(p.Address) == strings.ToLower(a.Address)
// }

// // Method GetCommonAddress return 'common.Address' that generated from 'this'.Address field. Useful for Ethereum like work
// func (p *Pair) GetCommonAddress() common.Address {
// 	return common.HexToAddress(p.Address)
// }
