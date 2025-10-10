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
	QuoteCoin  string `json:"quoteCoin"`
	Symbol     string `json:"symbol"` // "BTC_USDT"
	CreateTime int64  `json:"createTime"`
	ImageUrl   string `json:"baseCoinIconUrl"`
}

type MexcContractTicksRes struct {
	Success bool               `json:"success"`
	Code    int                `json:"code"`
	Data    []MexcContractTick `json:"data"`
}

type MexcContractTick struct {
	Symbol      string  `json:"symbol"`
	Bid1        float64 `json:"bid1"`
	Ask1        float64 `json:"ask1"`
	Volume24H   float64 `json:"volume24"`
	MaxBidPrice float64 `json:"maxBidPrice"`
	MinAskPrice float64 `json:"minAskPrice"`
}

// 24h ticker stats
type MexcTickerStats struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	AskPrice string `json:"askPrice"`
	Volume   string `json:"volume"`
}

type MexcOrderBookTick struct {
	Symbol   string `json:"symbol"` // "BTCUSDT"
	BidPrice string `json:"bidPrice"`
	BidQty   string `json:"bidQty"`
	AskPrice string `json:"askPrice"`
	AskQty   string `json:"askQty"`
}

type MexcSpotTick struct {
	Symbol string

	Volume   string
	Deposit  bool
	Withdraw bool
	// always empty
	Contract string // maybe???

	BidPrice string
	BidQty   string
	AskPrice string
	AskQty   string
}

// Only for internal exchange work, use [MexcFutureTick] instead
type MexcFutureTickWS struct {
	Channel string `json:"channel"`
	Data    struct {
		Asks [][]float32 `json:"asks"`
		Bids [][]float32 `json:"bids"`
	} `json:"data"`
	Symbol string `json:"symbol"`
}

type MexcFutureTick struct {
	Symbol      string  `json:"symbol"`
	Bid1        float64 `json:"bidPrice"`
	Ask1        float64 `json:"askPrice"`
	MaxBidPrice float64 `json:"maxBidPrice"`
	MinAskPrice float64 `json:"minAskPrice"`

	Volume      string
	Deposit     bool
	WithdrawFee string
	Withdraw    bool
	Contract    string
}

type MexcTokenMeta struct {
	Volume string

	Deposit     bool
	WithdrawFee string
	Withdraw    bool
	Contract    string
}

type MexcTokenMetaUpdate struct {
	Volume *string

	Deposit     *bool
	WithdrawFee *string
	Withdraw    *bool
	Contract    *string
}

type MexcExchangeInfoRes struct {
	Symbols []MexcExchangeInfo `json:"symbols"`
}

type MexcExchangeInfo struct {
	Symbol     string `json:"symbol"`
	BaseAsset  string `json:"baseAsset"`
	QuoteAsset string `json:"quoteAsset"`
}
