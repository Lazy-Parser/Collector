package manager_dex

import (
	"context"
	"errors"
	"fmt"
	"log"

	database "github.com/Lazy-Parser/Collector/internal/database"
	d "github.com/Lazy-Parser/Collector/internal/domain"
	"github.com/Lazy-Parser/Collector/internal/utils"
	"github.com/ethereum/go-ethereum/common"
)

func New() *ManagerDex {
	return &ManagerDex{
		list: []*d.DataSourceDex{},
	}
}

func (m *ManagerDex) Push(collector d.DataSourceDex) error {
	if collector == nil {
		return errors.New("cannot push a nil collector")
	}
	m.list = append(m.list, &collector)
	fmt.Println("Pushed!")

	return nil
}

// do not start in new goroutine! Method run make every provided collector run in seperate goroutine
func (m *ManagerDex) Run(ctx context.Context) error {
	// ASSIGN PAIRS TO THE CORRESPONDING COLLECTOR
	var pancakeswapV3Pairs []database.Pair
	consumerChan := make(chan d.PancakeswapV2Responce, 1000)

	// load whitelist (allowed networks / pools)
	_, err := utils.LoadWhitelistFile()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// this is hardcode. Change in future
	allPairs, err := database.GetDB().PairService.GetAllPairs()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Tokens to listen:")
	for _, pair := range allPairs {
		if pair.Pool == "pancakeswap" && pair.Label == "v3" {
			pancakeswapV3Pairs = append(pancakeswapV3Pairs, pair)
			fmt.Printf("%s/%s\n", pair.BaseToken.Name, pair.QuoteToken.Name)
		}
	}
	fmt.Printf("TOTAL: %d", len(pancakeswapV3Pairs))

	// start all collectors. Hardcode, change
	for _, collector := range m.list {
		go startCollector(ctx, collector, pancakeswapV3Pairs, consumerChan)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Manager...")
			return nil
		case message := <-consumerChan:
			// just log
			pair := findPair(&allPairs, message.Hex)
			if pair == nil {
				fmt.Println("New message got, but didnt find appropriate pair! Message:")
				fmt.Printf(
					"HEX: %s\t|\tPOOL: %s\t|\tPRICE: %s\t| \n",
					message.Hex, message.Pool, message.Price.Text('f', 12),
				)
				continue
			}

			fmt.Printf(
				"%s/%s - %s\n",
				pair.BaseToken.Name, pair.QuoteToken.Name, message.Price.Text('f', 12),
			)
		}
	}
}

func startCollector(
	ctx context.Context,
	collector *d.DataSourceDex,
	toListen []database.Pair,
	consumerChan chan d.PancakeswapV2Responce,
) {
	fmt.Println("Starting collector....")
	err := (*collector).Init(&toListen)
	if err != nil {
		log.Fatalf("failed to init '%s' collector. %v", (*collector).Name(), err)
	}
	fmt.Println("Inited!")

	err = (*collector).Connect()
	if err != nil {
		log.Fatalf("failed to connect '%s' collector. %v", (*collector).Name(), err)
	}
	fmt.Println("Connected!")

	err = (*collector).Subscribe()
	if err != nil {
		log.Fatalf("failed to subscribe '%s' collector. %v", (*collector).Name(), err)
	}
	fmt.Println("Subscribed!")

	(*collector).Run(ctx, consumerChan)
}

func findPair(pairs *[]database.Pair, address string) *database.Pair {
	var res *database.Pair

	for _, pair := range *pairs {
		if pair.PairAddress == address {
			res = &pair
		}
	}

	return res
}

func (m *ManagerDex) FetchDecimals(collectorName string, toListen *[]database.Pair) (map[common.Address]uint8, error) {
	// select collector
	collector, err := m.getCollectorByName(collectorName)
	if err != nil {
		return nil, err
	}

	if collector == nil {
		return nil, errors.New("Collector with name '" + collectorName + "' did not found!")
	}

	err = (*collector).Init(toListen)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize collector: %v", err)
	}

	err = (*collector).Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to collector: %v", err)
	}

	res, err := (*collector).FetchDecimals(context.Background())
	if err != nil {
		return nil, fmt.Errorf("collector with name '%s' failed to fetchDecimals!, %v", (*collector).Name(), err)
	}

	return res, nil
}

// GetCollectorByName returns a pointer to the collector with the given name.
func (m *ManagerDex) getCollectorByName(collectorName string) (*d.DataSourceDex, error) {
	for i := range m.list {
		if (*m.list[i]).Name() == collectorName {
			return m.list[i], nil // Return the pointer to the collector
		}
	}
	return nil, fmt.Errorf("collector with name '%s' not found!", collectorName)
}
