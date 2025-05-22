package solana

import (
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
	toListen    *[]database.Pair
	conn        *websocket.Conn
	mapperID    map[uint]mapper        // map subscribtion ID -> dexscreener address
	baseVaults  map[string]*baseVault  // map dexscreener address -> vault
	quoteVaults map[string]*quoteVault // map dexscreener address -> vault
}

type mapper struct {
	address string
	isBase  bool
}

type baseVault struct {
	price        *big.Float
	vault        string // base address
	quoteAddress string // quote address from dexscrenner. Its not a vault!
}

type quoteVault struct {
	price       *big.Float
	vault       string // quote address
	baseAddress string // base address from dexscrenner. Its not a vault!
}

type Account struct {
	Data struct {
		Parsed struct {
			Info struct {
				TokenAmount struct {
					Amount   string `json:"amount"`   // raw on‐chain units
					Decimals int    `json:"decimals"` // their decimal scale
				} `json:"tokenAmount"`
			} `json:"info"`
		} `json:"parsed"`
	} `json:"data"`
}

// MultiAccountResponse is the shape of the HTTP getMultipleAccounts call:
//
//	resp.Result.Value → []Account
type LogMultiple struct {
	Result struct {
		Value []Account `json:"value"`
	} `json:"result"`
}

// AccountNotification is the WS “accountNotification” envelope.
//
//	Params.Subscription → your subID
//	Params.Value        → the Account payload
type Log struct {
	Params struct {
		Subscription uint `json:"subscription"`
		Result       struct {
			Value Account `json:"value"`
		} `json:"result"`
	} `json:"params"`
}

type LogAck struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  uint   `json:"result"`
	ID      uint64 `json:"id"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type SolanaPayload struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
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
