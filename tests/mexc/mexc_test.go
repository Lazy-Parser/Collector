package mexc_test

import (
	// "context"
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	// "testing"
	// "time"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/config"
	"github.com/Lazy-Parser/Collector/exchange"
	"github.com/Lazy-Parser/Collector/market"

	// "github.com/Lazy-Parser/Collector/market"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite

	mexcE exchange.Exchange
}

func TestMexcSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "..", "config", ".env")
	cfg, err := config.NewConfig(path)
	if err != nil {
		log.Fatal(err)
	}
	s.mexcE, err = exchange.NewMexc(api.NewMexcApi(cfg))
	if err != nil {
		log.Fatal(err)
	}

	s.Require().NoError(err, "failed to create mexc exchange")
}

//func (s *TestSuite) TestMexcSpot() {
//	s.T().Log("Booting mexc up...")
//
//	ctx := context.Background()
//
//	// buffer updates
//	s.mexcE.BufferLoop(ctx)
//
//	// blocking
//	ch := make(chan *market.MexcSpotTick, 1024)
//	go func() {
//		if err := s.mexcE.ListenSpot(ctx, ch); err != nil {
//			s.T().Errorf("listen spot error: %v", err)
//		}
//	}()
//	s.T().Log("Done!")
//
//	// listen ticks
//	ticker := time.NewTicker(time.Second * 2)
//	symbols := map[string]*market.MexcSpotTick{}
//	for {
//		select {
//		case <-ticker.C:
//			s.T().Logf("FARTCOIN: %+v", symbols["FARTCOINUSDT"])
//		case <-ctx.Done():
//			return
//		case tick := <-ch:
//			symbols[tick.Symbol] = tick
//		}
//	}
//}

// TODO: does not stream futures. Find if token exists on futures
func (s *TestSuite) TestMexcFutures() {
	s.T().Log("Booting mexc up...")

	ctx := context.Background()

	// fetch buffer first and subscribe
	s.mexcE.BufferLoop(ctx)

	// blocking
	ch := make(chan *market.MexcFutureTick, 1024)
	go func() {
		if err := s.mexcE.ListenFutures(ctx, ch); err != nil {
			s.T().Errorf("listen spot error: %v", err)
		}
	}()
	s.T().Log("Done!")

	// listen ticks
	ticker := time.NewTicker(time.Second * 2)
	symbols := map[string]*market.MexcFutureTick{}
	for {
		select {
		case <-ticker.C:
			s.T().Logf("Symbols: %d", len(symbols))
			s.T().Logf("FARTCOINUSDT: %+v", symbols["FARTCOINUSDT"])
		case <-ctx.Done():
			return
		case tick := <-ch:
			symbols[tick.Symbol] = tick
		}
	}
}
