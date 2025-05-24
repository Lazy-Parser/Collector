package module

import (
	"context"
	"errors"
	"fmt"
	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CreateAMM() *AMM {
	base := BaseEVMModule{
		toListen: make(map[string][]database.Pair),
	}
	return &AMM{BaseEVMModule: &base}
}

func (amm *AMM) Name() string { return "AMM" }

func (amm *AMM) Init() error {
	wd, _ := os.Getwd()
	abiJson, err := os.ReadFile(filepath.Join(wd, "contracts", "uniLikeSwapV2-swap.json"))
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', cannot load ABI file!", amm.Name())
	}

	parsedAbi, err := abi.JSON(strings.NewReader(string(abiJson)))
	if err != nil {
		return fmt.Errorf("[PANCAKESWAP][V2][Init] Failed to init '%s', cannot parse ABI file!", amm.Name())
	}

	amm.setAbi(parsedAbi)
	return nil
}

func (amm *AMM) GetSwapHash() common.Hash {
	swapSig := []byte("Swap(address,uint256,uint256,uint256,uint256,address)")
	return crypto.Keccak256Hash(swapSig)
}

func (amm *AMM) Subscribe(client *ethclient.Client, network string, logs chan types.Log) error {
	pairs, ok := amm.GetPairsByNetwork(network)
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
		Topics:    [][]common.Hash{{amm.GetSwapHash()}},
	}

	_, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return errors.New("failed to subscribe to swap event for amm " + network)
	}

	return nil
}

func (amm *AMM) HandleSwap(
	vLog types.Log,
	poolName string,
	token0Decimals int,
	token1Decimals int,
	isBaseToken0 bool,
) (core.CollectorDexResponse, error) {
	// Placeholder for the result we will build.
	var resp core.CollectorDexResponse

	// ------------------------------------------------------------------ decode
	// Swap(uint256 amount0In, uint256 amount1In,
	//      uint256 amount0Out, uint256 amount1Out)
	var ev struct {
		Amount0In, Amount1In   *big.Int
		Amount0Out, Amount1Out *big.Int
	}
	if err := amm.GetAbi().UnpackIntoInterface(&ev, "Swap", vLog.Data); err != nil {
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
			new(big.Float).Mul(amount1, decimals0),
			new(big.Float).Mul(amount0, decimals1),
		)
	} else {
		price = new(big.Float).Quo(
			new(big.Float).Mul(amount0, decimals0),
			new(big.Float).Mul(amount1, decimals1),
		)
	}

	// 5. Verify price sanity
	if price.Sign() <= 0 {
		return resp, fmt.Errorf("[V2][handleSwap] invalid price calculation")
	}

	resp = core.CollectorDexResponse{
		IsBaseToken0: isBaseToken0,
		Timestamp:    time.Now().UnixMilli(),
		Price:        price,
		Address:      vLog.Address.String(),
		From:         amm.Name(),
		Type:         "?",
	}
	return resp, nil
}
