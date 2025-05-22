package solana

import (
	"bytes"
	"context"
	"io"
	"math/big"
	"net/http"

	//"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"

	d "github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/Lazy-Parser/Collector/internal/utils"
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
	s.mapperID = make(map[uint]mapper, len(*pairs)*2)
	s.baseVaults = make(map[string]*baseVault, len(*pairs)+1)
	s.quoteVaults = make(map[string]*quoteVault, len(*pairs)+1)
	for _, pair := range *pairs {
		fmt.Printf("PAIR: %+v \n", pair)

		// If Lable is "", it means, that current pair use AMM.
		// For AMM we need to fetch vaults from tokens adresses using manager.FetchAndSaveMetadata().
		// For not AMM (CLMM, CPMM), tokens addresses are vaults, so we just use default addresses
		s.baseVaults[pair.BaseToken.Address] = &baseVault{
			price:        big.NewFloat(0.0),
			vault:        utils.TernaryIf(pair.Label == "", pair.BaseToken.Vault, pair.BaseToken.Address),
			quoteAddress: pair.QuoteToken.Address,
		}

		s.quoteVaults[pair.QuoteToken.Address] = &quoteVault{
			price:       big.NewFloat(0.0),
			vault:       utils.TernaryIf(pair.Label == "", pair.QuoteToken.Vault, pair.QuoteToken.Address),
			baseAddress: pair.BaseToken.Address,
		}
	}

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
	// subscribe base tokens
	for address, baseVault := range s.baseVaults {
		if err := s.subscribeVault(baseVault.vault, address, true); err != nil {
			return err
		}
		fmt.Printf("BaseVault: %s\n", baseVault.vault) // TODO: remove
		idCounter++
	}

	// subscribe base tokens
	for address, quoteVault := range s.quoteVaults {
		if err := s.subscribeVault(quoteVault.vault, address, false); err != nil {
			return err
		}
		fmt.Printf("QuotexVault: %s\n", quoteVault.vault) // TODO: remove
		idCounter++
	}

	return nil
}

func (s *Solana) subscribeVault(vault string, address string, isBase bool) error {
	msg := subscribeMsg{
		Jsonrpc: "2.0",
		ID:      idCounter,
		Method:  "accountSubscribe",
		Params: []interface{}{
			vault,
			map[string]interface{}{
				"encoding":   "jsonParsed",
				"commitment": "confirmed", // finalized, processed
			},
		},
	}
	if err := s.conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to subscribe vault '%s' to RpcEndpoint '%s'. %v", vault, s.Name(), err)
	}

	// read first message to save subscribtion ID
	for {
		_, raw, err := s.conn.ReadMessage()
		if err != nil {
			log.Printf("failed to read init message from '%s' RpcEndpoint. %v", s.Name(), err)
			return nil
		}
		var mes LogAck
		if err := json.Unmarshal(raw, &mes); err != nil {
			log.Printf("failed to unmarshal init message from '%s' RpcEndpoint. %v", s.Name(), err)
			return nil
		}
		if mes.Error != nil {
			log.Printf("failed to subscibe vault '%s' to '%s' RpcEndpoint. %v", vault, s.Name(), err)
		}

		s.mapperID[mes.Result] = mapper{address, isBase}
		fmt.Printf("LOG: map: %d - %s\n", mes.Result, address)
		break
	}

	return nil
}

// main listen loop
func (s *Solana) Run(ctx context.Context, consumerCh chan d.CollectorDexResponse) {
	err := s.fetchQuoteVaults()
	if err != nil {
		fmt.Printf("failed to fetch quote vaults: %v\n", err)
		ctx.Done()
	}

	for {
		// read message
		_, raw, err := s.conn.ReadMessage()
		if err != nil {
			log.Printf("failed to read message from '%s' RpcEndpoint. %v", s.Name(), err)
			continue
		}

		// Decode response message
		msg, err := unmarshallPayload(raw)
		if err != nil {
			fmt.Println(err)
		}

		decimals := msg.Params.Result.Value.Data.Parsed.Info.TokenAmount.Decimals
		amount := msg.Params.Result.Value.Data.Parsed.Info.TokenAmount.Amount
		tokenPrice, err := handleAccountUpdate(decimals, amount)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = s.updateTokenPrice(msg.Params.Subscription, tokenPrice)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("Token price: %s\n", tokenPrice.Text('f', 12))

		// price, ok := s.calculatePrice(msg.Params.Subscription)
		// if !ok {
		// 	fmt.Println("Failed to calculate price")
		// 	continue
		// }

		// mapper, ok := s.mapperID[msg.Params.Subscription]
		// if !ok {
		// 	fmt.Println("Failed to fetch address by subID")
		// 	continue
		// }
		// if !mapper.isBase {
		// 	continue
		// }

		// fmt.Printf("New price: %s \n", price.Text('f', 15))
	}
}

// we need to call it befor starting main loop function, because quote tokenes updates rarely, so to not wait
// for updated both tokens, we fetch quote tokens one time
func (s *Solana) fetchQuoteVaults() error {
	type vaultMap struct {
		address string
		vault   string
	}

	var vaultsMap []vaultMap
	var vaults []string
	for address, quoteVault := range s.quoteVaults {
		vaultsMap = append(vaultsMap, vaultMap{address: address, vault: quoteVault.vault})
		vaults = append(vaults, quoteVault.vault)
	}

	fmt.Printf("Quote tokens to fetch: %+v\n", vaults)

	payload := SolanaPayload{
		Jsonrpc: "2.0",
		ID:      idCounter,
		Method:  "getMultipleAccounts",
		Params: []interface{}{
			vaults,
			SolanaPayloadOptions{
				Encoding:   "jsonParsed",
				Commitment: "confirmed",
			},
		},
	}
	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal http getMultipleAccounts request: %v", err)
	}

	resp, err := http.Post(*rpcEndpointHttp, "application/json", bytes.NewBuffer(payloadRaw))
	if err != nil {
		return fmt.Errorf("faild to fetch getMultipleAccounts: %v", err)
	}
	defer resp.Body.Close()

	var msg LogMultiple
	body, err := io.ReadAll(resp.Body)
	// fmt.Printf("HTTP RES: %+v\n", string(body))
	if err != nil {
		return fmt.Errorf("faild to read body from getMultipleAccounts request: %v", err)
	}

	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("faild to unmarshall response from getMultipleAccounts request: %v", err)
	}

	fmt.Printf("%+v", msg)

	for idx, data := range msg.Result.Value {
		decimals := data.Data.Parsed.Info.TokenAmount.Decimals
		amount := data.Data.Parsed.Info.TokenAmount.Amount

		vaultType := vaultsMap[idx]
		price, err := handleAccountUpdate(decimals, amount)
		if err != nil {
			return fmt.Errorf("failed to handle account update from getMultipleAccounts request: %v", err)
		}

		fmt.Printf("Quote price: %s", price.Text('f', 12))
		// update price
		s.quoteVaults[vaultType.address].price = price
	}

	return nil
}

func unmarshallPayload(raw []byte) (Log, error) {
	var mes Log
	if err := json.Unmarshal(raw, &mes); err != nil {
		return Log{}, fmt.Errorf("failed to unmarshal message from 'solana' RpcEndpoint. %v", err)
	}

	return mes, nil
}

func handleAccountUpdate(decimals int, amount string) (*big.Float, error) {
	amountInt := big.NewInt(0)
	amountInt, ok := amountInt.SetString(amount, 10)
	if !ok {
		return nil, errors.New("failed to parse amount string field from response from RpcEndpoint for 'solana' vault")
	}

	amountBig := new(big.Float).SetInt(amountInt)
	decimalsBig := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	price := new(big.Float).Quo(
		amountBig,
		decimalsBig,
	)

	return price, nil
}

// updates price in baseVaults map or quoteVaults map
func (s *Solana) updateTokenPrice(subID uint, price *big.Float) error {
	mapper, ok := s.mapperID[subID]
	if !ok {
		return errors.New("Unknown subscribtion ID: " + fmt.Sprint(subID) + "!")
	}

	if mapper.isBase {
		s.baseVaults[mapper.address].price = price
	} else {
		s.quoteVaults[mapper.address].price = price
	}

	return nil
}

func (s *Solana) calculatePrice(subID uint) (*big.Float, bool) {
	mapper, ok := s.mapperID[subID]
	if !ok {
		return nil, false
	}

	var price0 *big.Float
	var price1 *big.Float

	var address string
	if mapper.isBase {
		baseVault := s.baseVaults[mapper.address]
		address = baseVault.vault

		price1 = baseVault.price
		price0 = s.quoteVaults[baseVault.quoteAddress].price
	} else {
		quoteVault := s.quoteVaults[mapper.address]
		address = quoteVault.vault

		price0 = quoteVault.price
		price1 = s.baseVaults[quoteVault.baseAddress].price
	}

	zero := big.NewFloat(0.0)
	if price0.Cmp(zero) == 0 {
		fmt.Println("cannot calculate price, beacause vault0 is not updated. Please, wait!")
		fmt.Printf("%s - %s", address, price1.Text('f', 12))

		return nil, false
	} else if price1.Cmp(zero) == 0 {
		fmt.Println("cannot calculate price, beacause vault1 is not updated. Please, wait!")
		fmt.Printf("%s - %s", address, price0.Text('f', 12))

		return nil, false
	}

	return new(big.Float).Quo(price0, price1), true
}

func (s *Solana) Unsubscribe() {
	// TODO
	// s.conn.WriteJSON()
}
