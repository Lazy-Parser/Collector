package pancakeswap_v2

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

	d "github.com/Lazy-Parser/Collector/internal/domain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type PairMeta struct {
	Name         string         // e.g. "WBNB/BUSD"
	Addr         common.Address // pair contract
	BaseAddr     common.Address // token we quote against (WBNB)
	QuoteAddr    common.Address // token whose price we output (BUSD)
	BaseIsToken0 bool
	DecBase      uint8 // filled automatically
	DecQuote     uint8 // filled automatically
}

type PancakeswapV2 struct {
	logs     chan types.Log
	toListen *[]d.Pair
	abi      abi.ABI // listen price
	client   *ethclient.Client
	sub      ethereum.Subscription
	meta     map[common.Address]PairMeta
}

var (
	wd, _         = os.Getwd()
	erc20AbiPath  = filepath.Join(wd, "internal", "impl", "collector", "dex", "pancakeswap_v2", "erc20-decimals.json")
	multicallPath = filepath.Join(wd, "internal", "impl", "collector", "dex", "pancakeswap_v2", "multicall.json")
	mcAddress     = common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")
)

func (p PancakeswapV2) Name() string {
	return "PancakeswapV2"
}

// toListen - all pairs, that this pool will listen
func (p PancakeswapV2) Init(toListen *[]d.Pair) error {
	if len(*toListen) == 0 {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', provided 'toListen' array is empty!", p.Name())
	}
	p.toListen = toListen

	// load ABI
	wd, _ := os.Getwd()
	abiJson, err := os.ReadFile(filepath.Join(wd, "internal", "impl", "collector", "dex", "pancakeswap_v2", "uniswapV2-swap.json"))
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', cannot load ABI file!", p.Name())
	}

	parsedAbi, err := abi.JSON(strings.NewReader(string(abiJson)))
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', cannot parse ABI file!", p.Name())
	}

	p.abi = parsedAbi

	return nil
}

func (p PancakeswapV2) Connect() error {
	client, err := ethclient.Dial("wss://bsc-mainnet.core.chainstack.com/4984a03359f19068c2334839ea14acd0")
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Connect] Failed to connect '%s', %v", p.Name(), err)
	}

	p.client = client

	return nil
}

func (p PancakeswapV2) Subscribe() error {
	var poolAddresses []common.Address
	for _, pair := range *p.toListen {
		address := common.HexToAddress(pair.PairAddress)
		poolAddresses = append(poolAddresses, address)
	}

	// Хэш события Swap
	swapSig := []byte("Swap(address,uint256,uint256,uint256,uint256,address)")
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

func (p PancakeswapV2) Run(ctx context.Context, consumerCh chan d.PancakeswapV2Responce) {
	for {
		select {
		case err := <-p.sub.Err():
			p.sub.Unsubscribe()
			p.client.Close()
			log.Fatal("[PANCAKESWAP][V2][Run]: Subscribtion error: %v", err)
		case vLog := <-p.logs:
			fmt.Println("Swap событие получено!")

			// Декодируем данные события
			eventData := make(map[string]interface{})
			err := p.abi.UnpackIntoMap(eventData, "Swap", vLog.Data)
			if err != nil {
				log.Fatal("[PANCAKESWAP][V2][Run]: swap message decoding error: %v", err)
			}

			// извлекаем и обрабатываем даныне
			res, err := handleSwap(p.abi, vLog)
			if err != nil {
				log.Fatal("[HandleSwap] error handleSwap: %v", err)
			}

			consumerCh <- res
		}
	}
}

func handleSwap(pairABI abi.ABI, vLog types.Log) (d.PancakeswapV2Responce, error) {
	baseIsToken0 := true
	// Placeholder for the result we will build.
	var resp d.PancakeswapV2Responce

	// ------------------------------------------------------------------ decode
	// Swap(uint256 amount0In, uint256 amount1In,
	//      uint256 amount0Out, uint256 amount1Out)
	var ev struct {
		Amount0In, Amount1In   *big.Int
		Amount0Out, Amount1Out *big.Int
	}
	if err := pairABI.UnpackIntoInterface(&ev, "Swap", vLog.Data); err != nil {
		return resp, fmt.Errorf("[V2][handleSwap]: decode: %w", err)
	}

	// ------------------------------------------------------------------ choose amounts
	// Convert *big.Int → *big.Float just once (all 18-dec, so no scaling).
	toF := func(i *big.Int) *big.Float { return new(big.Float).SetInt(i) }

	var wbnbPerBase *big.Float

	switch {
	//----------------------------------------------------------------------
	// token0  ->  token1
	//----------------------------------------------------------------------
	case ev.Amount0In.Sign() > 0:
		if baseIsToken0 {
			// sold BASE, received WBNB
			wbnbPerBase = new(big.Float).Quo(
				toF(ev.Amount1Out), // WBNB out
				toF(ev.Amount0In),  // BASE in
			)
		} else {
			// sold WBNB, bought BASE  → invert
			wbnbPerBase = new(big.Float).Quo(
				toF(ev.Amount0In),  // WBNB in
				toF(ev.Amount1Out), // BASE out
			)
		}

	//----------------------------------------------------------------------
	// token1  ->  token0
	//----------------------------------------------------------------------
	case ev.Amount1In.Sign() > 0:
		if baseIsToken0 {
			// bought BASE for WBNB
			wbnbPerBase = new(big.Float).Quo(
				toF(ev.Amount1In),  // WBNB in
				toF(ev.Amount0Out), // BASE out
			)
		} else {
			// sold BASE, received WBNB
			wbnbPerBase = new(big.Float).Quo(
				toF(ev.Amount0Out), // WBNB out
				toF(ev.Amount1In),  // BASE in
			)
		}

	default:
		return resp, fmt.Errorf("zero amounts in Swap")
	}

	// ------------------------------------------------------------------ response
	price, _ := wbnbPerBase.Float64()
	resp = d.PancakeswapV2Responce{
		Timestamp: time.Now().Local().String(), // use Unix ms for easier math
		Price:     price,
		Hex:       vLog.Address.String(),
	}
	return resp, nil
}

func (p PancakeswapV2) FetchDecimals(ctx context.Context) (map[common.Address]uint8, error) {
	if len(*p.toListen) == 0 {
		return nil, errors.New("empty token list")
	}

	// ------------------------------------------------  уникальный список
	set := map[common.Address]struct{}{}
	// for _, t := range *p.toListen {
	// 	// set[common.HexToAddress(t.BaseTokenAddress)] = struct{}{}
	// 	// set[common.HexToAddress(t.QuoteTokenAddress)] = struct{}{}
	// }
	list := make([]common.Address, 0, len(set))
	for a := range set {
		list = append(list, a)
	}

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
			out[list[i]] = uint8(r.ReturnData[len(r.ReturnData)-1])
		} else {
			out[list[i]] = 0 // не удалось — caller решает, что делать (обычно 18)
		}
	}

	return out, nil
}
