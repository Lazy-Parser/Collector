package market

type Token struct {
	Name        string
	Address     string
	Network     string
	Decimal     uint8
	
	Image_url   string
	CreateTime  int64
	
	WithdrawFee string
	Withdraw    bool
	Deposit     bool
}
