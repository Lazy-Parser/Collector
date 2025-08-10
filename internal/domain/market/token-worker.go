package market

import "slices"

type TokenWorker struct {
	tokens []Token
}

func NewTokenWorker() *TokenWorker {
	return &TokenWorker{
		tokens: make([]Token, 0),
	}
}

// Push a COPY of the token to the worker's internal array
func (tw *TokenWorker) Push(token Token) {
	tw.tokens = append(tw.tokens, token)
}

// Push a COPY of the token to the worker's internal array
func (tw *TokenWorker) PushMany(tokens []Token) {
	tw.tokens = append(tw.tokens, tokens...)
}

func (tw *TokenWorker) RemoveByAddress(address string) {
	for i, token := range tw.tokens {
		if token.Address == address {
			tw.tokens = slices.Delete(tw.tokens, i, i+1)
			break
		}
	}
}

// GetAllTokens returns a COPY of the worker's internal array
func (tw *TokenWorker) GetAllTokens() []Token {
	return slices.Clone(tw.tokens)
}

// Find by address (index + ok)
func (tw *TokenWorker) FindIndexByAddress(address string) (int, bool) {
	for i := range tw.tokens {
		if tw.tokens[i].Address == address {
			return i, true
		}
	}
	return -1, false
}

func (tw *TokenWorker) Get(i int) (Token, bool) {
	if i < 0 || i >= len(tw.tokens) {
		return Token{}, false
	}
	return tw.tokens[i], true
}

// Get by address (copy + ok)
func (tw *TokenWorker) FindByAddress(address string) (Token, bool) {
	if i, ok := tw.FindIndexByAddress(address); ok {
		return tw.tokens[i], true
	}
	return Token{}, false
}

func (tw *TokenWorker) Len() int {
	return len(tw.tokens)
}

func (tw *TokenWorker) Update(i int, updates TokenPatch) {
	if i < 0 || i >= len(tw.tokens) {
		return
	}

	if updates.Name != nil {
		tw.tokens[i].Name = *updates.Name
	}
	if updates.Decimal != nil {
		tw.tokens[i].Decimal = *updates.Decimal
	}
	if updates.Address != nil {
		tw.tokens[i].Address = *updates.Address
	}
	if updates.Image_url != nil {
		tw.tokens[i].Image_url = *updates.Image_url
	}
	if updates.Network != nil {
		tw.tokens[i].Network = *updates.Network
	}
	if updates.WithdrawFee != nil {
		tw.tokens[i].WithdrawFee = *updates.WithdrawFee
	}
	if updates.CreateTime != nil {
		tw.tokens[i].CreateTime = *updates.CreateTime
	}
}

func (tw *TokenWorker) LoadFromDB() {
	// TODO
}
