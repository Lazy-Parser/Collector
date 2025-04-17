package dex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	m "Collector/internal/models" // models
)

// идея для декса
// c мекса берем все монеты -> каждую монету парсим для апи декса (btc/usdt) ->
// запрос на получение блокчейна и всей инфы ->
// запрос на ифну монеты

func Run(ctx context.Context) {
	fmt.Println("Getting data about MAJOR/USDT")

	var pairs = []string{
		"ETH/USDT",
		"BTC/USDT",
		"SOL/USDT",
		"AVAX/USDT",
		"BNB/USDT",
		"ARB/USDT",
		"OP/USDT",
		"MATIC/USDT",
		"LINK/USDT",
		"ADA/USDT",
		"DOT/USDT",
		"FTM/USDT",
		"NEAR/USDT",
		"ATOM/USDT",
		"INJ/USDT",
		"APE/USDT",
		"UNI/USDT",
		"SUSHI/USDT",
		"AAVE/USDT",
		"DYDX/USDT",
		"XRP/USDT",
		"TRX/USDT",
		"RNDR/USDT",
		"DOGE/USDT",
		"SHIB/USDT",
		"PEPE/USDT",
		"FLOKI/USDT",
		"CAKE/USDT",
		"1INCH/USDT",
		"GMX/USDT",
		"CRV/USDT",
		"SNX/USDT",
		"RPL/USDT",
		"LDO/USDT",
		"GRT/USDT",
		"MASK/USDT",
		"STG/USDT",
		"JOE/USDT",
		"VRA/USDT",
		"WOO/USDT",
	}

	resChan := make(chan m.DexInfoResponse)
	// 1000 * 60 ms / 50 = 1500ms.
	for i := 0; i < 40; i++ {
		go GetInfo("https://api.dexscreener.com/latest/dex/search", pairs[i], resChan)
		res := <-resChan

		fmt.Printf(
			"%s/%s. ChainId: %s | PairAddress: %s\n",
			res.Pairs[0].BaseToken.Symbol, res.Pairs[0].QuoteToken.Symbol, res.Pairs[0].ChainID, res.Pairs[0].PairAddress,
		)

		time.Sleep(time.Millisecond * 1500)
	}
}

// Returns chainId, dexId, pairAddress, liquidity
func GetInfo(url string, coin string, res chan m.DexInfoResponse) error {
	resp, err := http.Get(url + "?q=" + coin)
	if err != nil {
		return fmt.Errorf("error trying to get info about coin -> %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error trying to read Body -> %w", err)
	}
	var body m.DexInfoResponse
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return fmt.Errorf("error trying to unmarchall json -> %w", err)
	}

	res <- body
	return nil
}

func delayHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Printf("➡️ %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		fmt.Printf("✅ Обработано за %v\n", time.Since(start))
	})
}
