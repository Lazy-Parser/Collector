package manager_dex

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/Lazy-Parser/Collector/internal/core"
	database "github.com/Lazy-Parser/Collector/internal/database"
	"github.com/Lazy-Parser/Collector/internal/generator"
	"github.com/Lazy-Parser/Collector/internal/ui"
	"github.com/Lazy-Parser/Collector/internal/utils"
)

func New() *ManagerDex {
	return &ManagerDex{
		quotePairs: map[string]*big.Float{}, // wbnb -> price in usdt
		list:       []*core.DataSourceDex{},
		pairs:      map[string][]database.Pair{},
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

// Preload quote changer pairs
func (m *ManagerDex) Init(dbPairs []database.Pair) bool {
	// quote changer pairs
	payload := make([]generator.QuoteToken, len(dbPairs))
	for i := 0; i < len(dbPairs); i++ {
		payload[i] = generator.QuoteToken{
			Address: dbPairs[i].BaseToken.Address,
			Name:    dbPairs[i].BaseToken.Name,
			Symbol:  dbPairs[i].BaseToken.Name,
			Network: dbPairs[i].Network,
		}
	}

	res := generator.LoadQuoteChangerPairs(context.Background(), payload)
	if len(res) == 0 {
		return false
	}

	// add res to the local quotePairs list
	for _, elem := range res {
		price, _ := new(big.Float).SetString(elem.PriceUsd)

		m.quotePairs[elem.Base.Address] = price
	}

	return true
}

// do not start in new goroutine! Method run make every provided collector run in seperate goroutine
func (m *ManagerDex) Run(ctx context.Context, dashboardChan chan core.CollectorDexResponse) error {
	// ASSIGN PAIRS TO THE CORRESPONDING COLLECTOR
	allPairs, _ := database.GetDB().PairService.GetAllPairs()
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

	// append all quoteChanger pairs to ui dashboard
	//for address, price := range m.quotePairs {
	//	payload := core.CollectorDexResponse{
	//		IsBaseToken0: true,
	//		From:         "Pre init",
	//		Timestamp:    time.Now().UnixMilli(),
	//		Price:        price,
	//		Address:      address,
	//		Type:         "quote",
	//	}
	//
	//	dashboardChan <- payload
	//}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Manager...")
			return nil
		case msg := <-consumerChan:
			pair := findPair(&allPairs, msg.Address)
			if pair.Type == "quote" { // save quoteChanger price. for example, wbnb -> 0.123
				m.quotePairs[pair.BaseToken.Address] = msg.Price
				dashboardChan <- msg
				continue
			}

			// Cast pair price to usdt (MBOX/WBNB -> MBOX/USDT). But first check if the current pair already contains usdt. (usdc and usdt are the same, the difference is about ~0.004%)
			if pair.QuoteToken.Name == "USDT" || pair.QuoteToken.Name == "USDC" || pair.QuoteToken.Name == "USD1" {
				dashboardChan <- msg
				continue
			}

			// find the quoteChangerToken to cast. (msg = MBOX/WBNB, quoteChangerToken must have the same WBNB.address)
			priceChanger := findQuoteChangerToken(&m.quotePairs, *pair)
			if priceChanger == nil {
				// if not found, do nothing, just log
				msg := fmt.Sprintf("'%s' didnt found in QuoteChanger arr to cast '%s' -> 'USDT'", pair.QuoteToken.Name, pair.QuoteToken.Name)
				ui.GetUI().LogsView(msg, "warning")
				continue
			}

			msg.Price = new(big.Float).Mul(msg.Price, priceChanger)
			dashboardChan <- msg
		}
	}
}

func startCollector(
	ctx context.Context,
	collector *core.DataSourceDex,
	consumerChan chan core.CollectorDexResponse,
) {
	err := (*collector).Init()
	if err != nil {
		msg := fmt.Sprintf("failed to init '%s' collector. %v", (*collector).Name(), err)
		ui.GetUI().LogsView(msg, "error")
	}
	ui.GetUI().LogsView("DEX Inited!", "log")

	err = (*collector).Connect()
	if err != nil {
		msg := fmt.Sprintf("failed to connect '%s' collector. %v", (*collector).Name(), err)
		ui.GetUI().LogsView(msg, "error")
	}
	ui.GetUI().LogsView("DEX Connected!", "log")

	err = (*collector).Subscribe()
	if err != nil {
		msg := fmt.Sprintf("failed to subscribe '%s' collector. %v", (*collector).Name(), err)
		ui.GetUI().LogsView(msg, "error")
	}
	ui.GetUI().LogsView("DEX Subscribed!", "log")

	ui.GetUI().LogsView("DEX Running...", "log")
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

func findQuoteChangerToken(mapper *map[string]*big.Float, pair database.Pair) *big.Float {
	for token, priceChanger := range *mapper {
		if token == pair.QuoteToken.Address {
			return priceChanger
		}
	}

	return nil
}

// Fetch decimals for all tokens and vaults for solana and save all to the database.
// It work with ALL collectors, not only with provided
func (m *ManagerDex) FetchAndSaveMetadata(fetchers *[]core.MetadataCollector) {
	for _, collector := range *fetchers {
		metadata, err := collector.FetchMetadata()
		if err != nil {
			msg := fmt.Errorf("failed to fetch metadata. %v", err).Error()
			ui.GetUI().LogsView(msg, "error")
			return
		}

		if err := saver(metadata); err != nil {
			msg := fmt.Errorf("failed to save metadata. %v", err).Error()
			ui.GetUI().LogsView(msg, "error")
			return
		}
	}
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
