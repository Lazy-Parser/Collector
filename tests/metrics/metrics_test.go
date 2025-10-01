package metrics_test

import (
	"testing"

	"github.com/Lazy-Parser/Collector/metrics"
)

func TestMetricks(t *testing.T) {
	t.Log("Start metrics...")
	if err := metrics.StartMetrics(); err != nil {
		t.Error(err)
	}
}
