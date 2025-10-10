package exchange_internal

import (
	"context"
	"fmt"
	"github.com/Lazy-Parser/Collector/api"
	httpclient "github.com/Lazy-Parser/Collector/internal/adapter/http"
	"github.com/Lazy-Parser/Collector/market"
	"log"
	"strconv"
	"strings"
	"time"
)

// Mexc
// Update Volume - every 5 min
// Update Deposit/Withdraw - every 10 min
// TODO: make function, that make a buffer to update (not by timer)
type Mexc struct {
	client *httpclient.Client
	api    api.MexcAPI

	symbols map[string]struct{}
	quotes  map[string]struct{}

	buffer map[string]*market.MexcTokenMeta
}

func (m *Mexc) Name() string {
	return "Mexc"
}

func (m *Mexc) StopAll(ctx context.Context) {

}

// ---- START INITIALIZATION ----

func NewMexc(api api.MexcAPI) (*Mexc, error) {
	symbols, quotes, err := fetchAllSymbols(api)
	if err != nil {
		return nil, err
	}

	return &Mexc{
		client:  httpclient.New(60, 1, time.Second*5), // 1 request per second
		api:     api,
		symbols: symbols,
		quotes:  quotes,
		buffer:  make(map[string]*market.MexcTokenMeta),
	}, nil
}

func fetchAllSymbols(api api.MexcAPI) (map[string]struct{}, map[string]struct{}, error) {
	symbolsRes, err := api.FetchExchangeInfo(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to created mexc exchange: %v", err)
	}
	symbols := make(map[string]struct{})
	quotes := make(map[string]struct{})
	for _, s := range symbolsRes {
		symbols[s.Symbol] = struct{}{}
		quotes[s.QuoteAsset] = struct{}{}
	}

	return symbols, quotes, nil
}

// ---- END INITIALIZATION

// ---- START BUFFER ----

func (m *Mexc) BufferLoop(ctx context.Context) error {
	// init requests
	err := m.fetchVolumeAndUpdate(ctx)
	if err != nil {
		return fmt.Errorf("buffer loop error: %v", err)
	}
	err = m.fetchDepositWithdrawAndUpdate(ctx)
	if err != nil {
		return fmt.Errorf("buffer loop error: %v", err)
	}

	log.Println("First queue burst!")

	go func() {
		ticker := time.NewTicker(time.Minute * 5)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.fetchVolumeAndUpdate(ctx); err != nil {
					// throw error somehow fmt.Errorf("buffer loop error: %v", err)
					// stop buffer if error
					//ctx.Done() - call ctx.Done() in parent service in error handler
				}
				// don't forget to run the burst method for the changes (update/delete/create sub) to take effect.
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Minute * 10)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.fetchDepositWithdrawAndUpdate(ctx); err != nil {
					// return error somehow
					//ctx.Done()
				}
				// do not call m.queueBurst() here, because only volume fetch append commands to the queue
			}
		}
	}()

	return nil
}

func (m *Mexc) fetchVolumeAndUpdate(ctx context.Context) error {
	stats, err := m.api.Fetch24hTickerStats(ctx)
	if err != nil {
		return err
	}

	for _, stat := range stats {
		// update buffer. From this request we get full symbol (BTCUSDT)
		m.update(stat.Symbol, market.MexcTokenMetaUpdate{Volume: &stat.Volume})
	}

	return nil
}

func (m *Mexc) fetchDepositWithdrawAndUpdate(ctx context.Context) error {
	confs, err := m.api.FetchCurrencyInformation(ctx)
	if err != nil {
		return err
	}

	for _, c := range confs {
		n := c.NetworkList[0]
		changes := market.MexcTokenMetaUpdate{
			Deposit:     &n.DepositEnable,
			Withdraw:    &n.WithdrawEnable,
			WithdrawFee: &n.WithdrawFee,
			Contract:    &n.Contract,
		}

		// Update buffer. From this request we get short symbol (BTC)
		// That means that we need to update all full symbols, where first coin is a received short symbol. Example:
		// Get: <- 'BTC'
		// To update in buffer:
		//  - 'BTCUSDT'
		//  - 'BTCUSDC'
		//  - 'BTCEUR'
		//  - 'BTC...'
		for _, symbolToUpdate := range m.coinToSymbols(c.Coin) {
			m.update(symbolToUpdate, changes)
		}
	}

	return nil
}

// take short symbol ('BTC', 'ETH', ...) and returns a list of symbols with all pairs variations, where base token is a provided one
//
// Example: BTC -> BTCUSDT, BTCUSDC, BTCEUR, BTC...
func (m *Mexc) coinToSymbols(coin string) []string {
	var res []string

	for quote := range m.quotes {
		if _, exists := m.symbols[coin+quote]; exists {
			res = append(res, coin+quote)
		}
	}

	return res
}

func (m *Mexc) update(symbol string, update market.MexcTokenMetaUpdate) {
	// find token in buffer
	tokenMeta, exist := m.bufferFind(symbol)

	if update.Volume != nil {
		// parse volume
		volume, err := strconv.ParseFloat(*update.Volume, 64)
		if err != nil {
			return
		}
		volumeBiggerMin := volume > volumeMin

		if exist {
			if volumeBiggerMin {
				// update
				m.bufferUpdate(tokenMeta, update)
			} else {
				// delete. TODO: Do not forger to push msg to the queue
				m.bufferDelete(symbol)
			}
		} else {
			if volumeBiggerMin {
				// if len(m.buffer) >= 30 { // limit for tests
				// 	return
				// }
				// create. TODO: Do not forger to push msg to the queue
				m.bufferCreate(symbol, update)
			} else {
				// nothing
			}
		}
	}

	if update.Deposit != nil || update.Withdraw != nil {
		if exist {
			m.bufferUpdate(tokenMeta, update)
		}
		return
	}
}

func (m *Mexc) bufferFind(symbol string) (*market.MexcTokenMeta, bool) {
	res, ok := m.buffer[symbol]
	return res, ok
}

func (m *Mexc) bufferUpdate(bufferElem *market.MexcTokenMeta, update market.MexcTokenMetaUpdate) {
	if update.Volume != nil {
		bufferElem.Volume = *update.Volume
	}
	if update.Deposit != nil {
		bufferElem.Deposit = *update.Deposit
	}
	if update.Withdraw != nil {
		bufferElem.Withdraw = *update.Withdraw
	}
	if update.WithdrawFee != nil {
		bufferElem.WithdrawFee = *update.WithdrawFee
	}
	if update.Contract != nil {
		bufferElem.Contract = *update.Contract
	}
}

// Accept only full symbol
func (m *Mexc) bufferCreate(symbol string, update market.MexcTokenMetaUpdate) {
	if symbol == "" {
		return
	}

	m.buffer[symbol] = &market.MexcTokenMeta{Volume: *update.Volume}
}

func (m *Mexc) bufferDelete(symbol string) {
	delete(m.buffer, symbol)
}

// ---- END  BUFFER -----

// ---- START LISTENERS ----

// ListenSpot - blocking
func (m *Mexc) ListenSpot(ctx context.Context, ch chan *market.MexcSpotTick) error {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			// TODO: make with httpclient
			tickers, err := m.api.FetchOrderBookTicker(ctx)
			if err != nil {
				return err
			}
			for _, ticker := range tickers {
				// try to find ticker in the buffer
				tickerBuffer, ok := m.bufferFind(ticker.Symbol)
				if !ok {
					continue
				}

				ch <- m.createSpotTick(&ticker, tickerBuffer)
			}
		}
	}
}
func (m *Mexc) createSpotTick(tick *market.MexcOrderBookTick, bufferData *market.MexcTokenMeta) *market.MexcSpotTick {
	return &market.MexcSpotTick{
		Symbol: tick.Symbol,

		Volume:   bufferData.Volume,
		Deposit:  bufferData.Deposit,
		Withdraw: bufferData.Withdraw,
		Contract: bufferData.Contract,

		BidPrice: tick.BidPrice,
		BidQty:   tick.BidQty,
		AskPrice: tick.AskPrice,
		AskQty:   tick.AskQty,
	}
}

// ----

func (m *Mexc) ListenFutures(ctx context.Context, ch chan *market.MexcFutureTick) error {
	// request limit is 20 req / 2 sec, but it will be okay 1 req / 1 sec
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			tickers, err := m.api.FetchContractTicker(ctx)
			if err != nil {
				return err
			}
			if !tickers.Success {
				return fmt.Errorf("fetch future ticker fail, error code: %d", tickers.Code)
			}

			for _, ticker := range tickers.Data {
				// find in buffer
				// from contract tick got symbol "BTC_USDT". Remove "_" to find in the buffer
				symbolCommon := strings.ReplaceAll(ticker.Symbol, "_", "")
				bufferTicker, ok := m.bufferFind(symbolCommon)
				if !ok {
					continue
				}

				ticker.Symbol = symbolCommon
				ch <- m.createFutureTick(&ticker, bufferTicker)
			}
		}
	}
}

func (m *Mexc) createFutureTick(data *market.MexcContractTick, buffer *market.MexcTokenMeta) *market.MexcFutureTick {
	return &market.MexcFutureTick{
		Symbol:      data.Symbol,
		Bid1:        data.Bid1,
		Ask1:        data.Ask1,
		MaxBidPrice: data.MaxBidPrice,
		MinAskPrice: data.MinAskPrice,

		Volume:      buffer.Volume, // TODO: decide to pass volume from the futures Volume or Buffer Volume
		Deposit:     buffer.Deposit,
		WithdrawFee: buffer.WithdrawFee,
		Withdraw:    buffer.Withdraw,
		Contract:    buffer.Contract, // can be empty
	}
}

// ---- END LISTENERS ----
