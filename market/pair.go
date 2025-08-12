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

// type PairWithPtr struct {
// 	BaseToken  *Token
// 	QuoteToken *Token

// 	Address string
// 	Network string
// 	Pool    string

// 	Label string
// 	URL   string
// 	Type  string
// }

// func (p *PairWithPtr) ToPair() Pair {
// 	return Pair{
// 		BaseToken:  *p.BaseToken,
// 		QuoteToken: *p.QuoteToken,
// 		Address:    p.Address,
// 		Network:    p.Network,
// 		Pool:       p.Pool,
// 		Label:      p.Label,
// 		URL:        p.URL,
// 		Type:       p.Type,
// 	}
// }

// type PairUpdate struct {
// 	BaseToken  *Token
// 	QuoteToken *Token
// 	Address    *string
// 	Network    *string
// 	Pool       *string
// 	Label      *string
// 	URL        *string
// 	Type       *string
// }
