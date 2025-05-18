package solana

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/mr-tron/base58"
)

type SolanaHelper struct {
	pairs *[]database.Pair
}

func (h *SolanaHelper) PushPairs(pairs *[]database.Pair) {
	h.pairs = pairs
}

func (h *SolanaHelper) FetchMetadata() (core.Metadata, error) {
	metadata := core.Metadata{}

	vaults, err := fetchVaults(h.pairs)
	if err != nil {
		return metadata, fmt.Errorf("failed to fetch metadata: %v", err)
	}

	decimals, err := fetchDecimals(h.pairs)
	if err != nil {
		return metadata, fmt.Errorf("failed to fetch metadata: %v", err)
	}

	metadata.Vaults = vaults
	metadata.Decimals = decimals
	metadata.ToSave = "all"

	return metadata, nil
}

// fetch decimlas for all provided tokens. All pairs must be from Solana network!
// Return array, where Key - token address, Value - vault
func fetchVaults(pairs *[]database.Pair) (map[string]string, error) {
	res := make(map[string]string, len(*pairs)*2)

	// create array of unique pairs addresses
	var addresses []string
	for _, pair := range *pairs {
		addresses = append(addresses, pair.PairAddress)
	}

	req := SolanaPayload{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getMultipleAccounts",
		Params: []interface{}{
			addresses,
			map[string]interface{}{"encoding": "base64"},
		},
	}
	payload, _ := json.Marshal(req)

	response, err := http.Post(*rpcEndpointHttp, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decimals in '%s', %v", "solana", err)
	}
	defer response.Body.Close()

	var body SolanaResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode decimals responce in %v", err)
	}

	for idx, value := range body.Result.Value {
		base64Str := value.Data[0]
		decodedStr, err := base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64Str getMultipleAccounts. %v", err)
		}

		// check raydium version
		var offsetBaseVault = 0
		var offsetQuoteVault = 0
		switch len(decodedStr) {
		case 752: // v4
			offsetBaseVault = offsetBaseVaultV4
			offsetQuoteVault = offsetQuoteVaultV4
			break
		case 1544: // CLMM
			offsetBaseVault = offsetBaseVaultCLMM
			offsetQuoteVault = offsetQuoteVaultCLMM
			break
		case 637: // Other
			offsetBaseVault = offsetBaseVault637
			offsetQuoteVault = offsetQuoteVault637
			break
		default:
			fmt.Printf("decoded string from getMultipleAccounts (%d) has %d bytes, but only 752 accepted (Raydium V4) \n", idx, len(decodedStr))
			break
		}

		if offsetBaseVault == 0 || offsetQuoteVault == 0 {
			continue
		}

		// decode and save vaults in appropriate index
		baseVault, quoteVault := decodeSolanaRPCResponse(decodedStr, offsetBaseVault, offsetQuoteVault)
		curPair := (*pairs)[idx]
		res[curPair.BaseToken.Address] = baseVault
		res[curPair.QuoteToken.Address] = quoteVault
	}

	return res, nil
}

func decodeSolanaRPCResponse(raw []byte, offsetBase int, offsetQuote int) (string, string) {
	baseVault := base58.Encode(raw[offsetBase : offsetBase+pubkeyLen])
	quoteVault := base58.Encode(raw[offsetQuote : offsetQuote+pubkeyLen])

	return baseVault, quoteVault
}

func fetchDecimals(pairs *[]database.Pair) (map[string]uint8, error) {
	res := make(map[string]uint8, len(*pairs)*2)
	// create array of unique tokens addresses
	var mints []string
	set := map[string]struct{}{}
	for _, pair := range *pairs {
		set[pair.BaseToken.Address] = struct{}{}
		set[pair.QuoteToken.Address] = struct{}{}
	}
	for address := range set {
		mints = append(mints, address)
	}

	req := SolanaPayload{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getMultipleAccounts",
		Params: []interface{}{
			mints,
			map[string]interface{}{"encoding": "base64"},
		},
	}
	payload, _ := json.Marshal(req)

	response, err := http.Post(*rpcEndpointHttp, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decimals in '%s', %v", "solana", err)
	}
	defer response.Body.Close()

	var body core.SolanaRpcResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode decimals responce in '%s', %v", "solana", err)
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
