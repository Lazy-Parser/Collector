package solana

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Lazy-Parser/Collector/internal/database"
	d "github.com/Lazy-Parser/Collector/internal/domain"
	"github.com/gorilla/websocket"
)

var (
	// rpcEndpointWs   = flag.String("rpcEndpintWs", "wss://solana-mainnet.core.chainstack.com/171bc800908f187df7686f3f75c3080f", "Chainstack rpc endpoint for listening swaps")
	rpcEndpointHttp = flag.String("rpcEndpointHttp", "http://solana-mainnet.core.chainstack.com/171bc800908f187df7686f3f75c3080f", "Chainstack rpc endpoint for fething decimals")
	rpcEndpoint     = "wss://solana-mainnet.core.chainstack.com/171bc800908f187df7686f3f75c3080f"
	idCounter       = 1
)

func (s *Solana) Name() string {
	return "Solana"
}

func (s *Solana) Init(pairs *[]database.Pair) error {
	flag.Parse()

	if len(*pairs) == 0 {
		return errors.New("in '" + s.Name() + "' init failed: provided pairs list is empty!")
	}
	s.toListen = pairs

	// init
	s.vaults = make(map[string]*vaultState, len(*pairs)*2+10)
	s.subID = make(map[int]string, len(*pairs)*2+10)

	return nil
}

func (s *Solana) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(rpcEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to RpcEndpoint '%s'. %v", s.Name(), err)
	}

	s.conn = conn

	return nil
}

func (s *Solana) Subscribe() error {
	var vault string
	for _, pair := range *s.toListen {
		// subscribe on base vault
		vault = pair.BaseToken.Address
		if err := s.subscribeVault(vault); err != nil {
			return err
		}
		s.subID[idCounter] = vault
		s.vaults[vault] = &vaultState{
			Latest: nil,
			Prev:   nil,
			Pair:   &pair,
			IsBase: true,
		}
		idCounter++

		// subscribe on quote vault
		vault = pair.QuoteToken.Address
		if err := s.subscribeVault(vault); err != nil {
			return err
		}
		s.subID[idCounter] = vault
		s.vaults[vault] = &vaultState{
			Latest: nil,
			Prev:   nil,
			Pair:   &pair,
			IsBase: false,
		}
		idCounter++

	}

	return nil
}

func (s *Solana) subscribeVault(vault string) error {
	msg := subscribeMsg{
		Jsonrpc: "2.0",
		ID:      idCounter,
		Method:  "accountSubscribe",
		Params: []interface{}{
			vault,
			map[string]interface{}{
				"encoding":   "base64",
				"commitment": "confirmed", // finalized, processed
			},
		},
	}
	if err := s.conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to subscribe to RpcEndpoint '%s'. %v", s.Name(), err)
	}

	return nil
}

// main listen loop
func (s *Solana) Run(ctx context.Context, consumerCh chan d.CollectorDexResponse) {
	for {
		// read message
		_, raw, err := s.conn.ReadMessage()
		if err != nil {
			log.Fatalf("failed to read message from '%s' RpcEndpoint. %v", s.Name(), err)
			continue
		}

		// Encode response message
		var mes Log
		if err := json.Unmarshal(raw, &mes); err != nil {
			log.Fatalf("failed to unmarshal message from '%s' RpcEndpoint. %v", s.Name(), err)
			continue
		}

		// subscription ACK
		// if m.Result != nil && m.ID != 0 {
		// 	var subID int
		// 	json.Unmarshal(m.Result, &subID)
		// 	active[subID] = pending[m.ID]
		// 	delete(pending, m.ID)
		// 	continue
		// }

		// logs
		if mes.Method == "logsNotification" && mes.Params != nil {
			// do here
		}
	}
}

func (s *Solana) Unsubscribe() {
	// TODO
	// s.conn.WriteJSON()
}

// fetch decimlas for all provided tokens. All pairs must be from Solana network!
// Return array, where Key - mint address, value - decimal for address
func (s *Solana) FetchDecimals(pairs *[]database.Pair) (map[string]uint8, error) {
	res := make(map[string]uint8, len(*pairs)*2)
	// create array of unqiue tokens addresses
	var mints []string
	set := map[string]struct{}{}
	for _, pair := range *pairs {
		set[pair.BaseToken.Address] = struct{}{}
		set[pair.QuoteToken.Address] = struct{}{}
	}
	for address := range set {
		mints = append(mints, address)
	}

	req := RpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getMultipleAccounts",
		Params: []interface{}{
			mints,
			map[string]interface{}{"encoding": "base64", "commitment": "finalized"},
		},
	}
	payload, _ := json.Marshal(req)

	response, err := http.Post(*rpcEndpointHttp, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decimals in '%s', %v", s.Name(), err)
	}
	defer response.Body.Close()

	var body d.SolanaRpcResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode decimals responce in '%s', %v", s.Name(), err)
	}

	for idx, resp := range body.Result.Value {
		if resp.Error != nil || len(resp.Data) == 0 {
			continue
		}

		blob64 := resp.Data[0].(string)
		raw, _ := base64.StdEncoding.DecodeString(blob64)
		if len(raw) < 45 { // mint layout is 82 bytes, offset 44 holds decimals
			continue
		}

		decimal := raw[44]
		res[mints[idx]] = decimal
	}

	return res, nil
}

func (s *Solana) Disconnect() {
	for _, m := range *s.toListen {
		idCounter++

		msg := subscribeMsg{
			Jsonrpc: "2.0",
			ID:      idCounter, // can be any client-chosen request ID
			Method:  "logsUnsubscribe",
			Params:  []interface{}{subscriptionID},
		}
	}

	s.conn.Close()
}
