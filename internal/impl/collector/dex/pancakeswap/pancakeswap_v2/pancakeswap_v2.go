package pancakeswap_v2

import (
	"context"
	"fmt"
	"log"

	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	d "github.com/Lazy-Parser/Collector/internal/core"
	database "github.com/Lazy-Parser/Collector/internal/database"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type PancakeswapV2 struct {
	logs     chan types.Log
	toListen *[]database.Pair
	abi      abi.ABI // listen price
	client   *ethclient.Client
	sub      ethereum.Subscription
}

func (p *PancakeswapV2) Name() string {
	return "PancakeswapV2"
}

// toListen - all pairs, that this pool will listen
func (p *PancakeswapV2) Init(toListen *[]database.Pair) error {
	if len(*toListen) == 0 {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', provided 'toListen' array is empty!", p.Name())
	}
	p.toListen = toListen

	// load ABI
	wd, _ := os.Getwd()
	abiJson, err := os.ReadFile(filepath.Join(wd, "internal", "impl", "collector", "dex", "pancakeswap", "pancakeswap_v2", "abi", "pancakeswapV2-swap.json"))
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

func (p *PancakeswapV2) Connect() error {
	client, err := ethclient.Dial("wss://bsc-mainnet.core.chainstack.com/d467926f7a436f3d20ad07c7c65dab08")
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Connect] Failed to connect '%s', %v", p.Name(), err)
	}

	p.client = client

	return nil
}

func (p *PancakeswapV2) Subscribe() error {
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

func (p *PancakeswapV2) Run(ctx context.Context, consumerCh chan d.CollectorDexResponse) {
	for {
		select {
		case err := <-p.sub.Err():
			p.sub.Unsubscribe()
			p.client.Close()
			log.Fatal("[PANCAKESWAP][V2][Run]: Subscribtion error: %v", err)
		case vLog := <-p.logs:
			// извлекаем и обрабатываем даныне
			curPair := findPair(p.toListen, vLog.Address.String())
			token0, _ := sortTokens(
				common.HexToAddress(curPair.BaseToken.Address),
				common.HexToAddress(curPair.QuoteToken.Address),
			)
			isBaseToken0 := isBaseToken(token0, common.HexToAddress(curPair.BaseToken.Address))

			res, err := handleSwap(
				p.abi,
				vLog,
				p.Name(),
				curPair.BaseToken.Decimals,
				curPair.QuoteToken.Decimals,
				isBaseToken0,
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
	token0Decimals int,
	token1Decimals int,
	isBaseToken0 bool,
) (d.CollectorDexResponse, error) {
	// Placeholder for the result we will build.
	var resp d.CollectorDexResponse

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

	// 2. Validate swap amounts
	if (ev.Amount0In.Sign() > 0 && ev.Amount1In.Sign() > 0) ||
		(ev.Amount0Out.Sign() > 0 && ev.Amount1Out.Sign() > 0) {
		return resp, fmt.Errorf("[V2][handleSwap] invalid swap amounts")
	}

	// 3. Calculate decimal adjustment factors
	decimals0 := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token0Decimals)), nil))
	decimals1 := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token1Decimals)), nil))

	// 4. Determine swap direction and calculate price
	var price *big.Float
	var amount1, amount0 *big.Float
	switch {
	// Case 1: Token0 -> Token1 (sell Token0, buy Token1) - Sell
	case ev.Amount0In.Sign() > 0 && ev.Amount1Out.Sign() > 0:
		amount0 = new(big.Float).SetInt(ev.Amount0In)
		amount1 = new(big.Float).SetInt(ev.Amount1Out)

	// Case 2: Token1 -> Token0 (sell Token1, buy Token0) - Buy
	case ev.Amount1In.Sign() > 0 && ev.Amount0Out.Sign() > 0:
		amount1 = new(big.Float).SetInt(ev.Amount1In)
		amount0 = new(big.Float).SetInt(ev.Amount0Out)

	default:
		return resp, fmt.Errorf("[V2][handleSwap] invalid swap direction")
	}

	// calculate price
	if isBaseToken0 {
		price = new(big.Float).Quo(
			new(big.Float).Mul(amount0, decimals0),
			new(big.Float).Mul(amount1, decimals1),
		)
	} else {
		price = new(big.Float).Quo(
			new(big.Float).Mul(amount1, decimals0),
			new(big.Float).Mul(amount0, decimals1),
		)
	}

	// 5. Verify price sanity
	if price.Sign() <= 0 {
		return resp, fmt.Errorf("[V2][handleSwap] invalid price calculation")
	}

	resp = d.CollectorDexResponse{
		Timestamp: time.Now().UnixMilli(),
		Price:     price,
		Address:   vLog.Address.String(),
		From:      poolName,
		Type:      "?",
	}
	return resp, nil
}

func findPair(pairs *[]database.Pair, pairAddress string) *database.Pair {
	var res *database.Pair

	for _, pair := range *pairs {
		if pair.PairAddress == pairAddress {
			res = &pair
		}
	}

	return res
}

// Правильная сортировка токенов как в контрактах Uniswap/PancakeSwap
func sortTokens(tokenA, tokenB common.Address) (token0, token1 common.Address) {
	a := new(big.Int).SetBytes(tokenA.Bytes())
	b := new(big.Int).SetBytes(tokenB.Bytes())

	if a.Cmp(b) > 0 {
		return tokenA, tokenB
	}
	return tokenB, tokenA
}

func isBaseToken(token, baseToken common.Address) bool {
	return strings.EqualFold(token.Hex(), baseToken.Hex())
}
