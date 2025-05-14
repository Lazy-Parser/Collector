package solana

import (
	"encoding/json"

	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/gorilla/websocket"
)

type subscribeMsg struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type Solana struct {
	toListen *[]database.Pair
	subID    []int
	conn     *websocket.Conn
}

type Log struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Params  *struct {
		Subscription int `json:"subscription"`
		Result       struct {
			Value struct {
				Signature string   `json:"signature"`
				Err       string   `json:"err"`
				Logs      []string `json:"logs"`
			} `json:"value"`
		} `json:"result"`
	} `json:"params,omitempty"`
}

type RpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}