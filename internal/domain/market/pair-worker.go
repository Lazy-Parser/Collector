package market

type PairWorker struct {
	pairs []PairWithPtr
	// mu sync.RWMutex
}

func NewPairWorker() *PairWorker {
	return &PairWorker{
		pairs: []PairWithPtr{},
	}
}

func (pw *PairWorker) PushPair(pair PairWithPtr) {
	pw.pairs = append(pw.pairs, pair)
}

func (pw *PairWorker) GetPairs() []Pair {
	var pairs []Pair
	for i := range pw.pairs {
		pairs = append(pairs, pw.pairs[i].ToPair())
	}
	return pairs
}

func (pw *PairWorker) GetPairByAddress(address string) (Pair, bool) {
	for i := range pw.pairs {
		if pw.pairs[i].Address == address {
			return pw.pairs[i].ToPair(), true
		}
	}
	return Pair{}, false
}

func (pw *PairWorker) GetAllTokens() []Token {
	var tokens []Token
	for i := range pw.pairs {
		tokens = append(tokens, *pw.pairs[i].BaseToken)
		tokens = append(tokens, *pw.pairs[i].QuoteToken)
	}
	return tokens
}

func (pw *PairWorker) GetQuoteTokens() []Token {
	var tokens []Token
	for i := range pw.pairs {
		tokens = append(tokens, *pw.pairs[i].QuoteToken)
	}
	return tokens
}

func (pw *PairWorker) Update(address string, updates PairUpdate) bool {
	for i := range pw.pairs {
		if pw.pairs[i].Address == address {
			if updates.BaseToken != nil {
				pw.pairs[i].BaseToken = updates.BaseToken
			}
			if updates.QuoteToken != nil {
				pw.pairs[i].QuoteToken = updates.QuoteToken
			}
			if updates.Address != nil {
				pw.pairs[i].Address = *updates.Address
			}
			if updates.Label != nil {
				pw.pairs[i].Label = *updates.Label
			}
			if updates.Network != nil {
				pw.pairs[i].Network = *updates.Network
			}
			if updates.Pool != nil {
				pw.pairs[i].Pool = *updates.Pool
			}
			if updates.Type != nil {
				pw.pairs[i].Type = *updates.Type
			}
			if updates.URL != nil {
				pw.pairs[i].URL = *updates.URL
			}
			return true
		}
	}
	return false
}
