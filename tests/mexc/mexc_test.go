package mexc_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/config"
	"github.com/Lazy-Parser/Collector/exchange"
	"github.com/Lazy-Parser/Collector/market"
)

// TODO: add metrics to the ws client
func TestMexc(t *testing.T) {
	// first create buffer
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "..", "config", ".env")
	cfg, err := config.NewConfig(path)
	if err != nil {
		t.Fatalf("failed to setup config service. \nError: %v", err)
	}

	t.Log("Initing...")
	mexcApi := api.NewMexcApi(cfg)
	mexcE, err := exchange.NewMexc(mexcApi)
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	// blocking
	ch := make(chan *market.MexcSpotTick, 1024)
	go func() {
		if err := mexcE.ListenSpot(ctx, ch); err != nil {
			t.Error(err)
		}
	}()

	t.Log("Exchange created, try to fetch buffer")
	if err := mexcE.BufferLoop(ctx); err != nil {
		t.Error(err)
	}
	t.Log("Buffer Loop started success")

	coins := map[string]struct{}{}
	ticker := time.NewTicker(time.Second)
	// listen ticks
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			log.Printf("COINS: %d", len(coins))

		case tick := <-ch:
			coins[tick.Symbol] = struct{}{}
		}
	}

	close(ch)
	return
}
