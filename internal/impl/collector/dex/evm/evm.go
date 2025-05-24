package evm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	core "github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/impl/collector/dex/evm/module"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EVM struct {
	logs    chan types.Log
	clients map[string]*ethclient.Client
	sub     ethereum.Subscription
	modules *[]module.EVMModuleImplementation
}

// TODO: load from file conf in future
var (
	networkURL = map[string]string{
		"ethereum": "wss://ethereum-mainnet.core.chainstack.com/2db495046ce3bdf84a3074a3bff733e5",
		"bsc":      "wss://bsc-mainnet.core.chainstack.com/d467926f7a436f3d20ad07c7c65dab08",
	}
)

func (p *EVM) Name() string {
	return "EVM"
}

// toListen - all pairs, that this pool will listen
func (p *EVM) Init(modules *[]module.EVMModuleImplementation) error {
	if len(*modules) == 0 {
		return errors.New("provided modules arr is empty in " + p.Name())
	}
	p.modules = modules

	return nil
}

func (p *EVM) Connect() error {
	// get all networks that we will use
	var networks map[string]struct{}
	for _, m := range *p.modules {
		for network := range *m.GetPairs() {
			networks[network] = struct{}{}
		}
	}

	// for all networks create clients
	for network := range networks {
		client, err := ethclient.Dial(networkURL[network])
		if err != nil {
			return errors.New("failed to connect to '" + network + "' network")
		}

		p.clients[network] = client
	}

	return nil
}

func (p *EVM) Subscribe() error {
	p.logs = make(chan types.Log)

	// subscribe to all network clients
	for _, m := range *p.modules {
		for network := range *m.GetPairs() {
			if err := m.Subscribe(p.clients[network], network, p.logs); err != nil {
				return fmt.Errorf("failed to subcribe '%s' to '%s' network", m.Name(), network)
			}

			fmt.Printf("'%s' subscribed to '%s' sucessful/n", m.Name(), network)
		}

	}

	return nil
}

func (p *EVM) Run(ctx context.Context, consumerCh chan core.CollectorDexResponse) {
	for {
		select {
		case vLog := <-p.logs:
			// извлекаем и обрабатываем даныне
			for _, m := range *p.modules {
				if m.GetSwapHash() == vLog.Topics[0] {
					p.handleSwap(&m, vLog, consumerCh)
					break
				}
			}
		}
	}
}

func (p *EVM) handleSwap(m *module.EVMModuleImplementation, vLog types.Log, consumerCh chan core.CollectorDexResponse) {
	curPair := (*m).FindPair(vLog.Address.String())
	token0, _ := sortTokens(
		common.HexToAddress(curPair.BaseToken.Address),
		common.HexToAddress(curPair.QuoteToken.Address),
	)
	isBaseToken0 := isBaseToken(token0, common.HexToAddress(curPair.BaseToken.Address))

	res, err := (*m).HandleSwap(
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
