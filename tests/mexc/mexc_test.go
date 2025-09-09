package mexc_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/config"
	"github.com/Lazy-Parser/Collector/exchange"
)

func TestMexc(t *testing.T) {
	// first create buffer
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "..", "config", ".env")
	cfg, err := config.NewConfig(path)
	if err != nil {
		t.Fatalf("failed to setup config service. \nError: %v", err)
	}

	mexcApi := api.NewMexcApi(cfg)
	mexcE, err := exchange.NewMexc(mexcApi)
	if err != nil {
		t.Error(err)
	}

	t.Log("Exchange created, try to fetch buffer")
	if err := mexcE.BufferLoop(context.Background()); err != nil {
		t.Error(err)
	}
	t.Log("Fetched")

	for {
		i := 1
		for symbol, meta := range *mexcE.GetBuffer() {
			t.Logf("%d) Symbol: %s | Meta: %+v", i, symbol, meta)
			i++
		}
		time.Sleep(time.Second * 10)
	}
}
