package mexc

// ---------- mexc ----------
// get all coins from mexc
type Network struct {
	Network        string `json:"netWork"`
	Contract       string `json:"contract"` // 0x…
	DepositEnable  bool   `json:"depositEnable"`
	WithdrawEnable bool   `json:"withdrawEnable"`
	WithdrawFee    string `json:"withdrawFee"`
}
type Asset struct {
	Coin        string    `json:"coin"`        // "ETH"
	NetworkList []Network `json:"networkList"` // all chains
}

// future pairs
type Contracts struct {
	Data []ContractDetail `json:"data"`
}
type ContractDetail struct {
	BaseCoin   string `json:"baseCoin"`
	Symbol     string `json:"symbol"` // "BTC_USDT"
	CreateTime int64  `json:"createTime"`
	ImageUrl   string `json:"baseCoinIconUrl"`
}

// second return type (bool) means could this func find token by symbol
func findContractBySymbol(arr *[]ContractDetail, symbol string) (ContractDetail, bool) {
	for _, contract := range *arr {
		if contract.BaseCoin == symbol {
			return contract, true
		}
	}

	return ContractDetail{}, false
}