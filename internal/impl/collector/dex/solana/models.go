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

type SolanaPayload struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      uint16        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"` // SolanaPayloadOptions should be in 'Params' as the last value
}
type SolanaPayloadOptions struct {
	Encoding   string `json:"encoding"`
	Commitment string `json:"commitment"` // "processed" | "confirmed" | "finalized"
}

// RPCResponse represents the top‐level JSON-RPC response.
type SolanaResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  Result `json:"result"`
	ID      int    `json:"id"`
}

// Result holds the context and the account value.
type Result struct {
	Context Context `json:"context"`
	Value   []Value `json:"value"`
}

// Context gives RPC version and slot.
type Context struct {
	APIVersion string `json:"apiVersion"`
	Slot       uint64 `json:"slot"`
}

// Value contains the account data and metadata.
type Value struct {
	Data       []string `json:"data"`       // [base64Data, "base64"]
	Executable bool     `json:"executable"` // always false for non-program accounts
	Lamports   uint64   `json:"lamports"`
	Owner      string   `json:"owner"`
	RentEpoch  uint64   `json:"rentEpoch"`
	Space      uint64   `json:"space"`
}
