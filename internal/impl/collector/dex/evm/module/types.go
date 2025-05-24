package module

import (
	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type IEvmModule interface {
	Push(toListen *[]database.Pair, network string)
	setAbi(abi abi.ABI)
	GetAbi() abi.ABI
	GetAllPairs() *map[string]*[]database.Pair
	FindPair(address string) *database.Pair
	GetPairsByNetwork(network string) (*[]database.Pair, bool)
}
type BaseEVMModule struct {
	toListen map[string]*[]database.Pair
	abi      abi.ABI
}

type EVMModuleImplementation interface {
	IEvmModule
	Name() string
	Init() error
	Subscribe(client *ethclient.Client, network string, logs chan types.Log) error
	GetPairs() *map[string]*[]database.Pair
	FindPair(address string) *database.Pair
	GetSwapHash() common.Hash
	HandleSwap(
		vLog types.Log,
		poolName string,
		token0Decimals int,
		token1Decimals int,
		isBaseToken0 bool,
	) (core.CollectorDexResponse, error)
}

type AMM struct {
	IEvmModule
}

type CLMM struct {
	IEvmModule
}
