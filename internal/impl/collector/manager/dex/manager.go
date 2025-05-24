package manager_dex

import (
	"context"
	"errors"
	"fmt"
	"github.com/Lazy-Parser/Collector/internal/ui"
	"log"

	"github.com/Lazy-Parser/Collector/internal/core"
	database "github.com/Lazy-Parser/Collector/internal/database"
	"github.com/Lazy-Parser/Collector/internal/utils"
)

func New() *ManagerDex {
	return &ManagerDex{
		list:  []*core.DataSourceDex{},
		pairs: map[string][]database.Pair{},
	}
}

func (m *ManagerDex) Push(collector core.DataSourceDex) error {
	if collector == nil {
		return errors.New("cannot push a nil collector")
	}

	// save collector and pairs for this collector
	m.list = append(m.list, &collector)

	return nil
}

// do not start in new goroutine! Method run make every provided collector run in seperate goroutine
func (m *ManagerDex) Run(ctx context.Context) error {
	// ASSIGN PAIRS TO THE CORRESPONDING COLLECTOR
	//allPairs, _ := database.GetDB().PairService.GetAllPairs()
	consumerChan := make(chan core.CollectorDexResponse, 1000)

	// load whitelist (allowed networks / pools)
	_, err := utils.LoadWhitelistFile()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// start all collectors. Hardcode, change
	for _, collector := range m.list {
		go startCollector(ctx, collector, consumerChan)
	}

	ui.GetUI().ShowCollectorPrices(consumerChan)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Manager...")
			return nil
			//case message := <-consumerChan:
			// just log
			//pair := findPair(&allPairs, message.Address)
			//fmt.Printf(
			//	"[%s]: %s/%s - %s\n",
			//	message.From, pair.BaseToken.Name, pair.QuoteToken.Name, message.Price.Text('f', 12),
			//)
		}
	}
}

func startCollector(
	ctx context.Context,
	collector *core.DataSourceDex,
	consumerChan chan core.CollectorDexResponse,
) {
	fmt.Println("Starting collector....")
	err := (*collector).Init()
	if err != nil {
		log.Fatalf("failed to init '%s' collector. %v", (*collector).Name(), err)
	}
	fmt.Println("Inited!")

	err = (*collector).Connect()
	if err != nil {
		log.Fatalf("failed to connect '%s' collector. %v", (*collector).Name(), err)
	}
	fmt.Println("Connected!")

	fmt.Println("Subscribing...")
	err = (*collector).Subscribe()
	if err != nil {
		log.Fatalf("failed to subscribe '%s' collector. %v", (*collector).Name(), err)
	}
	fmt.Println("Subscribed!")

	fmt.Println("Running...")
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

// Fetch decimals for all tokens and vaults for solana and save all to the database.
// It work with ALL collectors, not only with provided
func (m *ManagerDex) FetchAndSaveMetadata(fetchers *[]core.MetadataCollector) error {
	for _, collector := range *fetchers {
		metadata, err := collector.FetchMetadata()
		if err != nil {
			return fmt.Errorf("failed to fetch metadata. %v", err)
		}

		if err := saver(metadata); err != nil {
			return fmt.Errorf("failed to save metadata. %v", err)
		}
	}

	return nil
}
func saver(metadata core.Metadata) error {
	switch metadata.ToSave {
	case "decimals":

		saveDecimals(&metadata.Decimals)
		break
	case "vaults":
		saveVaults(&metadata.Vaults)
		break
	case "all":
		saveVaults(&metadata.Vaults)
		saveDecimals(&metadata.Decimals)
		break
	default:
		return errors.New("invalid metadata to_save: '" + metadata.ToSave + "'. (Do not know what to save. Only 'decimals', 'vaults', 'all' available)")
		break
	}

	return nil
}
func saveDecimals(decimals *map[string]uint8) {
	for address, decimal := range *decimals {
		database.GetDB().TokenService.UpdateDecimals(&database.Token{Address: address}, decimal)
	}
}
func saveVaults(vaults *map[string]string) {
	for address, vault := range *vaults {
		database.GetDB().TokenService.UpdateVault(&database.Token{Address: address}, vault)
	}
}

// GetCollectorByName returns a pointer to the collector with the given name.
func (m *ManagerDex) getCollectorByName(collectorName string) (*core.DataSourceDex, error) {
	for i := range m.list {
		if (*m.list[i]).Name() == collectorName {
			return m.list[i], nil // Return the pointer to the collector
		}
	}
	return nil, fmt.Errorf("collector with name '%s' not found!", collectorName)
}

func (m *ManagerDex) Stop() {

}
