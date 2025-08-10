package api_mexc

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
