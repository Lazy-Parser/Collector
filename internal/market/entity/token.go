package market

import "strings"

type Token struct {
	Name        string
	Decimal     uint8
	Address     string
	Image_url   string
	WithdrawFee string
	CreateTime  int64
	Network     string
}

// bool indicates if Token was found by provided address or not
func FindTokenByAddress(tokens *[]Token, address string) (Token, bool) {
	for _, token := range *tokens {
		if strings.ToLower(token.Address) == strings.ToLower(address) {
			return token, true
		}
	}

	return Token{}, false
}

// // Method Compare compares provided Token 'a' with 'this' ('a') token according to their 'Address' fields. Returns TRUE if the addressess are the same. Register (a or A) does not matter.
// func (t *Token) Compare(a Pair) bool {
// 	return strings.ToLower(t.Address) == strings.ToLower(a.Address)
// }

// // Method GetCommonAddress return 'common.Address' that generated from 'this'.Address field. Useful for Ethereum like work
// func (t *Token) GetCommonAddress() common.Address {
// 	return common.HexToAddress(t.Address)
// }
