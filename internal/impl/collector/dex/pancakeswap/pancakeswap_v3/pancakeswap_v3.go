package pancakeswap_v3

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

func (p *PancakeswapV3) Run(ctx context.Context, consumerCh chan d.CollectorDexResponse) {
	for {
		select {
		case err := <-p.sub.Err():
			p.sub.Unsubscribe()
			p.client.Close()
			log.Fatal("[PANCAKESWAP][V2][Run]: Subscribtion error: %v", err)
		case vLog := <-p.logs:
			pair := findPair(p.toListen, vLog.Address.String())
			token0, _ := sortTokens(
				common.HexToAddress(pair.BaseToken.Address),
				common.HexToAddress(pair.QuoteToken.Address),
			)
			isBaseToken0 := isBaseToken(token0, common.HexToAddress(pair.BaseToken.Address))

			res, err := handleSwap(
				p.abi,
				vLog,
				p.Name(),
				pair.BaseToken.Decimals,
				pair.QuoteToken.Decimals,
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
	decimal0 int,
	decimal1 int,
	isBaseToken0 bool,
) (d.CollectorDexResponse, error) {
	// Placeholder for the result we will build.
	var resp d.CollectorDexResponse

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

	// caclulate price
	// priceRatio = sqrtPrice / 2^92
	priceRatio := new(big.Float).Quo(
		new(big.Float).SetInt(ev.SqrtPriceX96),
		new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil)),
	)
	decimalsDelta := int64(decimal1 - decimal0)
	// adjustedPrice = (priceRatio^2) / 10^(decimal0 - decimal1)
	// (priceRatio^2) = (priceRatio * priceRatio)
	adjustedPrice := new(big.Float).Quo(
		new(big.Float).Mul(priceRatio, priceRatio),
		new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(decimalsDelta), nil)),
	)

	// by default calculus are token1 / token0. So if Token0 is base, we should reverse price -> 1 / price
	if isBaseToken0 {
		adjustedPrice = new(big.Float).Quo(
			new(big.Float).SetInt(big.NewInt(1)),
			adjustedPrice,
		)
	}

	resp = d.CollectorDexResponse{
		Timestamp: time.Now().UnixMilli(),
		Price:     adjustedPrice,
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
