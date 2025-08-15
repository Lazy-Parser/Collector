package chains

type ChainMeta struct {
	Id        string    `json:"id"` // global name of network, that will be used in the whole app
	Providers Providers `json:"providers"`
}

type Providers struct {
	Dexscreener string `json:"dexscreener"`
	Coingecko   string `json:"coingecko"`
	Mexc        string `json:"mexc"`
}
