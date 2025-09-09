package market

type MexcNetwork struct {
	Coin           string `json:"coin"`
	Network        string `json:"netWork"`
	Contract       string `json:"contract"` // 0x…
	DepositEnable  bool   `json:"depositEnable"`
	WithdrawEnable bool   `json:"withdrawEnable"`
	WithdrawFee    string `json:"withdrawFee"`
}
type MexcAsset struct {
	Coin        string        `json:"coin"`        // "ETH"
	NetworkList []MexcNetwork `json:"networkList"` // all chains
}

// future pairs
type MexcContracts struct {
	Data []MexcContractDetail `json:"data"`
}
type MexcContractDetail struct {
	BaseCoin   string `json:"baseCoin"`
	Symbol     string `json:"symbol"` // "BTC_USDT"
	CreateTime int64  `json:"createTime"`
	ImageUrl   string `json:"baseCoinIconUrl"`
}

// 24h ticker stats
type MexcTickerStats struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	AskPrice string `json:"askPrice"`
	Volume   string `json:"volume"`
}

type MexcSpotTick struct {
	Symbol   string
	BidPrice string
	BidQty   string
	AskPrice string
	AskQty   string
}

type MexcFutureTick struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	AskPrice string `json:"askPrice"`
}

type MexcTokenMeta struct {
	Volume      string
	Deposit     bool
	WithdrawFee string
	Withdraw    bool
}
