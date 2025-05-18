package pancakeswap_helper

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	wd, _         = os.Getwd()
	erc20AbiPath  = filepath.Join(wd, "internal", "impl", "collector", "dex", "pancakeswap", "abi", "erc20-decimals.json")
	multicallPath = filepath.Join(wd, "internal", "impl", "collector", "dex", "pancakeswap", "abi", "multicall.json")
	mcAddress     = common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")
)

type PancakeswapHelper struct {
	pairs *[]database.Pair
}

func (h *PancakeswapHelper) PushPairs(pairs *[]database.Pair) {
	h.pairs = pairs
}

func (h *PancakeswapHelper) FetchMetadata() (core.Metadata, error) {
	metadata := core.Metadata{}

	decimals, err := fetchDecimals(h.pairs)
	if err != nil {
		return metadata, fmt.Errorf("failed to fetch decimals for Pancakeswap: %v", err)
	}

	metadata.Decimals = decimals
	metadata.ToSave = "decimals"
	return metadata, nil
}

func fetchDecimals(pairs *[]database.Pair) (map[string]uint8, error) {
	if len(*pairs) == 0 {
		return nil, errors.New("empty token list")
	}

	// init connection
	client, err := ethclient.Dial("wss://bsc-mainnet.core.chainstack.com/d467926f7a436f3d20ad07c7c65dab08")
	if err != nil {
		return nil, fmt.Errorf("[PANCAKESWAP][V2][Connect] Failed to connect '%s', %v", "panckaswap", err)
	}
	defer client.Close()

	// ------------------------------------------------  уникальный список
	set := map[common.Address]struct{}{}
	for _, t := range *pairs {
		set[common.HexToAddress(t.BaseToken.Address)] = struct{}{}
		set[common.HexToAddress(t.QuoteToken.Address)] = struct{}{}
	}
	list := make([]common.Address, 0, len(set))
	for a := range set {
		list = append(list, a)
	}
	fmt.Printf("Provided tokens to fetch decimals (%s): %d\n", "pancakeswap", len(list))

	// ------------------------------------------------  ABI helpers
	// load ABIs
	erc20Bytes, err := os.ReadFile(erc20AbiPath)
	if err != nil {
		return nil, err
	}

	multicallBytes, err := os.ReadFile(multicallPath)
	if err != nil {
		return nil, err
	}

	erc, err := abi.JSON(strings.NewReader(string(erc20Bytes)))
	if err != nil {
		return nil, err
	}

	mc, err := abi.JSON(strings.NewReader(string(multicallBytes)))
	if err != nil {
		return nil, err
	}

	decSig, _ := erc.Pack("decimals") // 0x313ce567

	// ------------------------------------------------  build Call[]
	type call struct {
		Target   common.Address
		CallData []byte
	}
	calls := make([]call, len(list))
	for i, t := range list {
		calls[i] = call{t, decSig}
	}

	// ------------------------------------------------  pack & call
	payload, _ := mc.Pack("tryAggregate", false, calls)

	raw, err := client.CallContract(context.Background(),
		ethereum.CallMsg{To: &mcAddress, Data: payload},
		nil,
	)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------  decode result
	var returns []struct {
		Success    bool
		ReturnData []byte
	}
	if err := mc.UnpackIntoInterface(&returns, "tryAggregate", raw); err != nil {
		return nil, err
	}

	out := make(map[string]uint8, len(list))
	for i, r := range returns {
		if r.Success && len(r.ReturnData) >= 32 {
			var decimal uint8
			err := erc.UnpackIntoInterface(&decimal, "decimals", r.ReturnData)
			if err != nil {
				out[list[i].String()] = 18 // Default value (could be set to 18 if decoding fails)
			} else {
				out[list[i].String()] = decimal
			}
		} else {
			out[list[i].String()] = 0 // не удалось — caller решает, что делать (обычно 18)
		}
	}

	return out, nil
}
