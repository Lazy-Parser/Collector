package mexc_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/config"
	"github.com/Lazy-Parser/Collector/exchange"
	"github.com/Lazy-Parser/Collector/market"
)

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

	// listen ticks
	for {
		select {
		case <-ctx.Done():
			return

		case tick := <-ch:
			t.Logf(
				"%s: ASK %s | BID %s | Withdraw %t | Deposit %t",
				tick.Symbol, tick.AskPrice, tick.BidPrice, tick.Withdraw, tick.Deposit,
			)
		}
	}

	close(ch)
	return
}
