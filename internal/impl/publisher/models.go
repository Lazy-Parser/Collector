package publisher

import "github.com/nats-io/nats.go"

type Publisher struct {
	nc *nats.Conn
}

type DexTick struct {
	Network     string `json:"network"`
	Pool        string `json:"Pool"`
	BaseToken   string `json:"base_token"`
	QuoteToken   string `json:"quote_token"`
	Price       string `json:"price"`
	PairAddress string `json:"pair_address"`
	Timestamp   int64  `json:"timestamp"`
}

type CexTick struct {
	Symbols   string `json:"symbols"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	Timestamp int64  `json:"timestamp"`
}