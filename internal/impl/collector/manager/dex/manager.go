package manager_dex

import (
	"encoding/json"
	// "fmt"
	"os"
	"path/filepath"

	d "github.com/Lazy-Parser/Collector/internal/domain"
)

var (
	MAX_COLLECTORS = 10
)

func New() *ManagerDex {
	return &ManagerDex{
		list: make([]d.DataSourceDex, MAX_COLLECTORS),
	}
}

func (m *ManagerDex) Push(collector d.DataSourceDex) {
	m.list = append(m.list, collector)
}

// do not start in new goroutine! Method run make every provided collector run in seperate goroutine
func (m *ManagerDex) Run() error {
	pairs, err := loadPairs()
	if err != nil {
		return err
	}

	// filter pairs for pancakeswap v2
	var filteredPairs []d.Pair
	for _, pair := range *pairs {
		if pair.Pool == "pancakeswap" && len(pair.Labels) != 0 && pair.Labels[0] == "v2" {
			filteredPairs = append(filteredPairs, pair)
			// fmt.Printf("Pair to listen: %s/%s, pair: %s\n", pair.BaseToken, pair.QuoteToken, pair.PairAddress)
		}
	}

	for _, collector := range m.list {
		go startCollector(&collector, filteredPairs)
	}

	return nil
}

func startCollector(collector *d.DataSourceDex, toListen []d.Pair) error {
	// implement something here
	return nil
}

func loadPairs() (*[]d.Pair, error) {
	var pairs []d.Pair

	wd, _ := os.Getwd()
	bytes, err := os.ReadFile(filepath.Join(wd, "config", "pairs.json"))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &pairs)
	if err != nil {
		return nil, err
	}

	return &pairs, err
}
