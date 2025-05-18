package solana

import (
	"context"
	//"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	d "github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/gorilla/websocket"
	"log"
)

var (
	// rpcEndpointWs   = flag.String("rpcEndpintWs", "wss://solana-mainnet.core.chainstack.com/171bc800908f187df7686f3f75c3080f", "Chainstack rpc endpoint for listening swaps")
	rpcEndpointHttp = flag.String("rpcEndpointHttp", "http://solana-mainnet.core.chainstack.com/171bc800908f187df7686f3f75c3080f", "Chainstack rpc endpoint for fething decimals")
	rpcEndpoint     = "wss://solana-mainnet.core.chainstack.com/171bc800908f187df7686f3f75c3080f"
	idCounter       = 1
)

const (
	offsetBaseVaultV4    = 336  // For Raydium V4 (752 bytes)
	offsetQuoteVaultV4   = 368  // For Raydium V4 (752 bytes)
	offsetBaseVaultCLMM  = 1360 // For CLMM Raydium (1544 bytes)
	offsetQuoteVaultCLMM = 1392 // For CLMM Raydium (1544 bytes)
	offsetBaseVault637   = 256  // For 637 bytes
	offsetQuoteVault637  = 288  // For 637 bytes
	pubkeyLen            = 32
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

// TODO
// Я уже могу получать vaults и decimals в одном запросе, но не везде я могу получит deciaml сразу, так что надо создать две фунции, 1 - получение decimals (уже есть на гитхабе), 2 - получение vaults - уже есть тут
// так же реализовать хранение vaults в базе данных (обновить бд)
