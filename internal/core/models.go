package core

type Token struct {
	Name    string
	Decimal uint8
	Address string
}

type Pair struct {
	BaseToken  Token
	QuoteToken Token
	Address    string
	Pool       string
	Network    string
}
