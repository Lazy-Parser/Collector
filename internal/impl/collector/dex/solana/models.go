package solana

import (
	"encoding/json"
	"math/big"

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
	subID    map[int]string
	conn     *websocket.Conn
	vaults   map[string]*vaultState
}

type vaultState struct {
	Latest *big.Int
	Prev   *big.Int
	Pair   *database.Pair
	IsBase bool
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
