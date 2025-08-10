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

type TokenPatch struct {
	Name        *string
	Decimal     *uint8
	Address     *string
	Image_url   *string
	WithdrawFee *string
	CreateTime  *int64
	Network     *string
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
