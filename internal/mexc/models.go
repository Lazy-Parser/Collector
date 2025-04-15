package mexc

type ConfType string

const (
	Futures ConfType = "FUTURES"
	Spot ConfType = "SPOT"
)

type MexcConf struct {
	Type		ConfType
	URL         string
	Subscribe   map[string]interface{}
	Unsubscribe map[string]interface{}
	ParseFunc 	func([]byte)  
}

type SpotMiniTickersResponse struct {
	Data     []SpotMiniTicker `json:"d"`
	Channel  string           `json:"c"` // "spot@public.miniTickers.v3.api@UTC+8"
	Ts       int64            `json:"t"` // Global timestamp
}

type SpotMiniTicker struct {
	Symbol     string `json:"s"`      // Trading pair
	Price      string `json:"p"`      // Last price
	Change     string `json:"r"`      // 24h change %
	TrueChange string `json:"tr"`     // True 24h change %
	High       string `json:"h"`      // 24h high
	Low        string `json:"l"`      // 24h low
	VolumeUSDT string `json:"v"`      // Quote volume
	VolumeBase string `json:"q"`      // Base volume
	LastRT     string `json:"lastRT"` // Possibly latency (always -1?)
	MT         string `json:"MT"`     // Market type ("0")
	NV         string `json:"NV"`     // Possibly placeholder ("--")
}
