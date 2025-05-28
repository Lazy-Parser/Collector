package module

import (
	"github.com/Lazy-Parser/Collector/internal/database"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

func (m *BaseEVMModule) Push(toListen []database.Pair, network string) {
	m.toListen[network] = toListen
}

func (m *BaseEVMModule) setAbi(abi abi.ABI) {
	m.abi = abi
}

func (m *BaseEVMModule) GetAbi() abi.ABI {
	return m.abi
}

func (m *BaseEVMModule) GetAllPairs() *map[string][]database.Pair {
	return &m.toListen
}

func (m *BaseEVMModule) FindPair(address string) *database.Pair {
	var res *database.Pair

	for _, pairs := range m.toListen {
		for _, pair := range pairs {
			if pair.PairAddress == address {
				res = &pair
			}
		}
	}

	return res
}

func (m *BaseEVMModule) GetPairsByNetwork(network string) (*[]database.Pair, bool) {
	pairs, ok := m.toListen[network]
	return &pairs, ok
}
