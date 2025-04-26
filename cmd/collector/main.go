package main

import (
	m "github.com/Lazy-Parser/Collector/internal/impl/collector/manager"
	mexc "github.com/Lazy-Parser/Collector/internal/impl/collector/cex"
	"github.com/Lazy-Parser/Collector/internal/impl/aggregator"
)

func main() {
	aggregator.InitJoiner()
	joiner := aggregator.GetJoiner()

	manager := m.CreateManager()
	manager.NewCollector(&mexc.MexcSource{})

	go manager.Run(joiner)
}
