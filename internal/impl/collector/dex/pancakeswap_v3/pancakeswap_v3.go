package pancakeswap_v3

import (
	"context"
	"errors"
	"fmt"
	"log"

	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	database "github.com/Lazy-Parser/Collector/internal/database"
	d "github.com/Lazy-Parser/Collector/internal/domain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type PancakeswapV3 struct {
	logs     chan types.Log
	toListen *[]database.Pair
	abi      abi.ABI // listen price
	client   *ethclient.Client
	sub      ethereum.Subscription
}

var (
	wd, _         = os.Getwd()
	erc20AbiPath  = filepath.Join(wd, "internal", "impl", "collector", "dex", "pancakeswap_v2", "abi", "erc20-decimals.json")
	multicallPath = filepath.Join(wd, "internal", "impl", "collector", "dex", "pancakeswap_v2", "abi", "multicall.json")
	mcAddress     = common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")
)

func (p *PancakeswapV3) Name() string {
	return "PancakeswapV3"
}

// toListen - all pairs, that this pool will listen
func (p *PancakeswapV3) Init(toListen *[]database.Pair) error {
	if len(*toListen) == 0 {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', provided 'toListen' array is empty!", p.Name())
	}
	p.toListen = toListen

	// load ABI
	wd, _ := os.Getwd()
	abiJson, err := os.ReadFile(filepath.Join(wd, "internal", "impl", "collector", "dex", "pancakeswap_v3", "abi", "PancakeswapV3-swap.json"))
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', cannot load ABI file!", p.Name())
	}

	parsedAbi, err := abi.JSON(strings.NewReader(string(abiJson)))
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', cannot parse ABI file!: %v", p.Name(), err)
	}

	p.abi = parsedAbi

	return nil
}

func (p *PancakeswapV3) Connect() error {
	client, err := ethclient.Dial("wss://bsc-mainnet.core.chainstack.com/d467926f7a436f3d20ad07c7c65dab08")
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Connect] Failed to connect '%s', %v", p.Name(), err)
	}

	p.client = client

	return nil
}

func (p *PancakeswapV3) Subscribe() error {
	var poolAddresses []common.Address
	for _, pair := range *p.toListen {
		address := common.HexToAddress(pair.PairAddress)
		poolAddresses = append(poolAddresses, address)
	}

	// Хэш события Swap
	swapSig := []byte("Swap(address,address,int256,int256,uint160,uint128,int24,uint128,uint128)")
	swapTopic := crypto.Keccak256Hash(swapSig)

	// Фильтр на событие Swap
	query := ethereum.FilterQuery{
		Addresses: poolAddresses,
		Topics:    [][]common.Hash{{swapTopic}},
	}

	p.logs = make(chan types.Log)

	sub, err := p.client.SubscribeFilterLogs(context.Background(), query, p.logs)
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Subscribe] Failed to subscribe on swap event'%s', %v", p.Name(), err)
	}

	p.sub = sub

	return nil
}

func (p *PancakeswapV3) Run(ctx context.Context, consumerCh chan d.PancakeswapV2Responce) {
	for {
		select {
		case err := <-p.sub.Err():
			p.sub.Unsubscribe()
			p.client.Close()
			log.Fatal("[PANCAKESWAP][V2][Run]: Subscribtion error: %v", err)
		case vLog := <-p.logs:
			fmt.Println("Swap событие получено!")

			res, err := handleSwap(
				p.abi,
				vLog,
				p.Name(),
			)
			if err != nil {
				log.Fatal("[HandleSwap] error handleSwap: %v", err)
			}

			consumerCh <- res
		}
	}
}

func handleSwap(
	pairABI abi.ABI,
	vLog types.Log,
	poolName string,
) (d.PancakeswapV2Responce, error) {
	// Placeholder for the result we will build.
	var resp d.PancakeswapV2Responce

	// ------------------------------------------------------------------ decode
	var ev struct {
		Amount0            *big.Int `abi:"amount0"`
		Amount1            *big.Int `abi:"amount1"`
		SqrtPriceX96       *big.Int `abi:"sqrtPriceX96"`
		Liquidity          *big.Int `abi:"liquidity"`
		Tick               *big.Int `abi:"tick"`
		ProtocolFeesToken0 *big.Int `abi:"protocolFeesToken0"`
		ProtocolFeesToken1 *big.Int `abi:"protocolFeesToken1"`
	}
	if err := pairABI.UnpackIntoInterface(&ev, "Swap", vLog.Data); err != nil {
		fmt.Println("Failed to parse logs")
		return resp, fmt.Errorf("[V3][handleSwap]: decode: %w", err)
	}

	fmt.Printf("\n")
	resp = d.PancakeswapV2Responce{
		Pool:      poolName,
		Timestamp: time.Now().Local().String(), // use Unix ms for easier math
		Price:     new(big.Float),
		Hex:       vLog.Address.String(),
	}
	return resp, nil
}

func (p *PancakeswapV3) FetchDecimals(ctx context.Context) (map[common.Address]uint8, error) {
	if len(*p.toListen) == 0 {
		return nil, errors.New("empty token list")
	}

	fmt.Printf("Provided list toListen: %d\n", len(*p.toListen))
	// ------------------------------------------------  уникальный список
	set := map[common.Address]struct{}{}
	for _, t := range *p.toListen {
		set[common.HexToAddress(t.BaseToken.Address)] = struct{}{}
		set[common.HexToAddress(t.QuoteToken.Address)] = struct{}{}
	}
	list := make([]common.Address, 0, len(set))
	for a := range set {
		list = append(list, a)
	}
	fmt.Printf("Provided list toListen: %d\n", len(list))

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

	raw, err := p.client.CallContract(ctx,
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

	out := make(map[common.Address]uint8, len(list))
	for i, r := range returns {
		if r.Success && len(r.ReturnData) >= 32 {
			var decimal uint8
			err := erc.UnpackIntoInterface(&decimal, "decimals", r.ReturnData)
			if err != nil {
				out[list[i]] = 18 // Default value (could be set to 18 if decoding fails)
			} else {
				out[list[i]] = decimal
			}
		} else {
			out[list[i]] = 0 // не удалось — caller решает, что делать (обычно 18)
		}
	}

	return out, nil
}

// func findPair(pairs *[]database.Pair, pairAddress string) *database.Pair {
// 	var res *database.Pair

// 	for _, pair := range *pairs {
// 		if pair.PairAddress == pairAddress {
// 			res = &pair
// 		}
// 	}

// 	return res
// }
