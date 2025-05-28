package manager_cex

import (
	"context"
	"fmt"
	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/database"
)

type CollectorManager struct {
	collectors  []core.DataSource
	pairsMapper map[string][]database.Pair
}

func CreateManager() *CollectorManager {
	return &CollectorManager{
		pairsMapper: make(map[string][]database.Pair),
		collectors:  []core.DataSource{},
	}
}

// Add new Collector to collectors array
func (m *CollectorManager) NewCollector(newCollector core.DataSource, pairs []database.Pair) {
	m.collectors = append(m.collectors, newCollector)
	m.pairsMapper[(newCollector).Name()] = pairs
}

// You should not start it in a new GoRoutine. This func iterate over each collector and start it.
func (m *CollectorManager) Run(ctx context.Context, consumerChan chan core.MexcResponse) {
	// mexc
	for _, collector := range m.collectors {
		if err := collector.Connect(); err != nil {
			fmt.Println(err.Error())
			//ui.GetUI().LogsView(err.Error())
		}
		//ui.GetUI().LogsView("Connected")

		if err := collector.Subscribe(ctx, m.pairsMapper[collector.Name()]); err != nil {
			fmt.Println(err.Error())
			//ui.GetUI().LogsView(err.Error())
		}
		//ui.GetUI().LogsView("Subscribed")

		go (collector).Run(ctx, consumerChan)
	}
}
