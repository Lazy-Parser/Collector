package manager

import (
	"context"
	"log"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/impl/aggregator"
)

type CollectorManager struct {
	collectors []core.DataSource
}

// TODO: maybe add context / logger
func CreateManager() *CollectorManager {
	return &CollectorManager{
		collectors: []core.DataSource{},
	}
}

// Add new Collector to collectors array
func (m *CollectorManager) NewCollector(newCollector core.DataSource) {
	m.collectors = append(m.collectors, newCollector)
}

// You should start it in a new GoRoutine. This func iterate over each collector and start it.
func (m *CollectorManager) Run(joiner *aggregator.Joiner) {
	for _, collector := range m.collectors {
		go func() {
			log.Printf("Starting %s collector...\n", collector.Name())

			// try to connect to the server (cex) / node (dex)
			err := collector.Connect()
			if err != nil {
				log.Panicf("Failed to connect to the '%s' collector. %v \n", collector.Name(), err)
				return // go to the next collector
			}

			// try to subscribe to some events
			err = collector.Subscribe()
			if err != nil {
				log.Panicf("Failed to subscribe to the event in the '%s' collector. %v \n", collector.Name(), err)
				return // go to the next collector
			}

			// start execution of current collector
			collector.Run(context.Background(), joiner.Push)
		}()
	}
}
