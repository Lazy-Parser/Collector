package module

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/Lazy-Parser/Collector/internal/ui"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func CreateCLMM() *CLMM {
	base := BaseEVMModule{
		toListen: make(map[string][]database.Pair),
	}
	return &CLMM{BaseEVMModule: &base}
}

func (clmm *CLMM) Name() string { return "CLMM" }

func (clmm *CLMM) Init() error {
	wd, _ := os.Getwd()
	abiJson, err := os.ReadFile(filepath.Join(wd, "contracts", "uniLikeSwapV3-swap.json"))
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', cannot load ABI file!", clmm.Name())
	}

	parsedAbi, err := abi.JSON(strings.NewReader(string(abiJson)))
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', cannot parse ABI file!", clmm.Name())
	}

	clmm.setAbi(parsedAbi)

	return nil
}

func (clmm *CLMM) GetSwapHash() common.Hash {
	swapSig := []byte("Swap(address,address,int256,int256,uint160,uint128,int24,uint128,uint128)")
	return crypto.Keccak256Hash(swapSig)
}

func (clmm *CLMM) Subscribe(client *ethclient.Client, network string, logs chan types.Log) error {
	pairs, ok := clmm.GetPairsByNetwork(network)
	if !ok {
		return errors.New("pairs for provided network '" + network + "' doesnt exits")
	}

	var poolAddresses []common.Address
	for _, pair := range *pairs {
		address := common.HexToAddress(pair.PairAddress)
		poolAddresses = append(poolAddresses, address)
	}

	// Фильтр на событие Swap
	query := ethereum.FilterQuery{
		Addresses: poolAddresses,
		Topics:    [][]common.Hash{{clmm.GetSwapHash()}},
	}

	_, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return errors.New("failed to subscribe to swap event for CLMM " + network)
	}

	return nil
}

func (clmm *CLMM) HandleSwap(
	vLog types.Log,
	poolName string,
	decimal0 int,
	decimal1 int,
	isBaseToken0 bool,
) (core.CollectorDexResponse, error) {
	// Placeholder for the result we will build.
	var resp core.CollectorDexResponse

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

	if err := clmm.GetAbi().UnpackIntoInterface(&ev, "Swap", vLog.Data); err != nil {
		// fmt.Println("Failed to parse logs")
		ui.GetUI().LogsView("Failed to parse logs", "error")
		return resp, fmt.Errorf("[V3][handleSwap]: decode: %w", err)
	}

	// calculate price
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
	if !isBaseToken0 {
		adjustedPrice = new(big.Float).Quo(
			new(big.Float).SetInt(big.NewInt(1)),
			adjustedPrice,
		)
	}

	resp = core.CollectorDexResponse{
		IsBaseToken0: isBaseToken0,
		Timestamp:    time.Now().UnixMilli(),
		Price:        adjustedPrice,
		Address:      vLog.Address.String(),
		From:         clmm.Name(),
		Type:         "?",
	}
	return resp, nil
}
