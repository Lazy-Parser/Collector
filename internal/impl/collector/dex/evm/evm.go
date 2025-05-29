package evm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	core "github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/impl/collector/dex/evm/module"
	"github.com/Lazy-Parser/Collector/internal/ui"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EVM struct {
	logs    chan types.Log
	clients map[string]*ethclient.Client
	sub     ethereum.Subscription
	modules []module.EVMModuleImplementation
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

func (p *EVM) Push(modules []module.EVMModuleImplementation) error {
	p.modules = modules
	return nil
}

// toListen - all pairs, that this pool will listen
func (p *EVM) Init() error {
	p.clients = make(map[string]*ethclient.Client)

	for _, m := range p.modules {
		m.Init()
	}

	return nil
}

func (p *EVM) Connect() error {
	// get all networks that we will use
	networks := make(map[string]struct{})
	for _, m := range p.modules {
		for network := range *m.GetAllPairs() {
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
	for _, m := range p.modules {
		for network := range *m.GetAllPairs() {
			if err := m.Subscribe(p.clients[network], network, p.logs); err != nil {
				return fmt.Errorf("failed to subcribe '%s' to '%s' network", m.Name(), network)
			}

			ui.GetUI().LogsView(fmt.Sprintf("'%s' subscribed to '%s' sucessful \n", m.Name(), network), "log")
		}

	}

	return nil
}

func (p *EVM) Run(ctx context.Context, consumerCh chan core.CollectorDexResponse) {
	for {
		select {
		case vLog := <-p.logs:
			// извлекаем и обрабатываем даныне
			for _, m := range p.modules {
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
	if curPair == nil {
		msg := fmt.Sprintf("no such pair '%s'", vLog.Address.String())
		ui.GetUI().LogsView(msg, "error")
		return
	}

	// Получаем адреса токенов из пары (как в контракте)
	tokenA := common.HexToAddress(curPair.BaseToken.Address)
	tokenB := common.HexToAddress(curPair.QuoteToken.Address)

	// Сортируем токены по адресам, как это делает Uniswap
	token0, _ := sortTokens(tokenA, tokenB)

	// Определяем, является ли токен0 базовым в системе
	isBaseToken0 := isBaseToken(token0, tokenA)

	res, err := (*m).HandleSwap(
		vLog,
		p.Name(),
		curPair.BaseToken.Decimals,
		curPair.QuoteToken.Decimals,
		isBaseToken0,
	)
	if err != nil {
		msg := fmt.Sprintf("[HandleSwap] error handleSwap: %v", err)
		ui.GetUI().LogsView(msg, "error")
		return
	}

	consumerCh <- res
}

// Правильная сортировка токенов как в контрактах Uniswap/PancakeSwap
func sortTokens(tokenA, tokenB common.Address) (token0, token1 common.Address) {
	comp := tokenA.Big().Cmp(tokenB.Big())

	if comp < 0 {
		return tokenA, tokenB
	}
	return tokenB, tokenA
}

func isBaseToken(token, baseToken common.Address) bool {
	return bytes.EqualFold(token.Bytes(), baseToken.Bytes())
}
