package exchange_internal

import (
	"context"
	"strconv"
	"sync"
	"time"
	"fmt"

	"github.com/Lazy-Parser/Collector/internal/adapter/ws"
	"github.com/Lazy-Parser/Collector/market"
	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/pb"
)

type operation int

const (
	Subscribe operation = iota
	Unsubscribe
)


// m.BufferLoop()
// m.ListenSpot()
type Mexc2 struct {
	conn *wsclient.Client
	api  api.MexcAPI

	symbols map[string]struct{}
	quotes  map[string]struct{}

	// xIndex are used for the work with both short and full symbols efficiently
	buffer  map[string]*market.MexcTokenMeta // key - full symbol ('BTCUSDT')
	queue   map[string]operation             // key - full symbol
	queueMu sync.RWMutex
}

func NewMexc2(api api.MexcAPI) (*Mexc2, error) {
	config := wsclient.NewClientConfig()
	config.SubTemplate = sub
	config.UnsubTamplate = unsub
	config.SubscriptionMaxChannels = 25
	config.UrlConnection = "wss://wbs-api.mexc.com/ws"

	// fetch all symbols here or not here (maybe in BufferLoop)
	symbols, quotes, err := fetchAllSymbols(api)
	if err != nil {
		return nil, err
	}

	return &Mexc2{
		api:     api,
		conn:    wsclient.NewClient(config),
		symbols: symbols,
		quotes:  quotes,
		buffer:  make(map[string]*market.MexcTokenMeta),
		queue:   make(map[string]operation),
	}, nil
}

// Fetches all symbols and quote coins from mexc.
//
// Returns two maps: symbols and quotes
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

func (m *Mexc2) Name() string {
	return "Mexc2"
}

// general
// symbol must be normalized for Volume!
// Done
func (m *Mexc2) update(symbol string, update market.MexcTokenMetaUpdate) {
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
				m.queuePush(symbol, Unsubscribe)
			}
		} else {
			if volumeBiggerMin {
				// create. TODO: Do not forger to push msg to the queue
				m.bufferCreate(symbol, update)
				m.queuePush(symbol, Subscribe)
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

// general

// buffer. TODO: make init buffer fetch
// Start in the main goroutine
func (m *Mexc2) BufferLoop(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Minute * 5)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats, err := m.api.Fetch24hTickerStats(ctx)
				if err != nil {
					// do smth
				}

				for _, stat := range stats {
					// update buffer. From this request we get full symbol (BTCUSDT)
					m.update(stat.Symbol, market.MexcTokenMetaUpdate{Volume: &stat.Volume})
				}

				// don't forget to run the burst method for the changes (update/delete/create) to take effect.
				m.queueBurst()
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
				confs, err := m.api.FetchCurrencyInformation(ctx)
				if err != nil {
					// do smth
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
					// That means, that we need to update all full symbols, where first coin is a received short symbol. Example:
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

				m.queueBurst()
			}
		}
	}()
}

// TODO: fetch here
func (m *Mexc2) BufferFetch(ctx context.Context, symbol string) error {
	return nil
}

// accept only short symbol
func (m *Mexc2) bufferFind(symbol string) (*market.MexcTokenMeta, bool) {
	res, ok := m.buffer[symbol]
	return res, ok
}

func (m *Mexc2) bufferUpdate(bufferElem *market.MexcTokenMeta, update market.MexcTokenMetaUpdate) {
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

// Accept full symbol
func (m *Mexc2) bufferDelete(symbol string) {
	delete(m.buffer, symbol)
}

// Accept only full symbol
func (m *Mexc2) bufferCreate(symbol string, update market.MexcTokenMetaUpdate) {
	if symbol == "" {
		return
	}

	m.buffer[symbol] = &market.MexcTokenMeta{Volume: *update.Volume}
}

// take short symbol ('BTC', 'ETH', ...) and returns symbols, where provided coin contains as base token
//
// Example: BTC -> BTCUSDT, BTCUSDC, BTCEUR, BTC...
func (m *Mexc2) coinToSymbols(coin string) []string {
	var res []string

	for quote := range m.quotes {
		if _, exists := m.symbols[coin+quote]; exists {
			res = append(res, coin+quote)
		}
	}

	return res
}

// buffer

// queue
// accept onlu full symbol
func (m *Mexc2) queuePush(symbol string, oper operation) {
	m.queueMu.Lock()
	defer m.queueMu.Unlock()

	m.queue[symbol] = oper
}

// Subscribe / unsubscribe symbols based on the queue
func (m *Mexc2) queueBurst() {
	m.queueMu.Lock()
	defer m.queueMu.Unlock()

	var subList []string
	var unsubList []string
	for symbol, oper := range m.queue {
		switch oper {
		case Subscribe:
			// TODO: decide for spot or futures connection
			subList = append(subList, symbol)
		case Unsubscribe:
			unsubList = append(unsubList, symbol)
		}
	}

	m.conn.Unsubscribe(unsubList)
	m.conn.Subscribe(subList)

	// empty queue
	for key := range m.queue {
		delete(m.queue, key)                              
	}
	// OR
	// m.queue = map[string]operation{}
}

// queue

// listeners
func (m *Mexc2) ListenSpot(ctx context.Context, ch chan *market.MexcSpotTick) error {
	// start all staff

	err := m.conn.Connect()
	if err != nil {
		return err
	}

	m.conn.PingLoop(ping, time.Second*30)
	if err := m.conn.Run(); err != nil {
		return err
	}

	// listener
	go func() {
		listenCh := m.conn.ListenTicks()
		for {
			select {
			case <-ctx.Done():
				return
			case wsData := <-listenCh:
				if wsData == nil {
					continue
				}

				bufferData, ok := m.bufferFind(*wsData.Symbol)
				if !ok {
					continue
				}

				ch <- m.createSpotTick(wsData, bufferData)
			}
		}
	}()

	return nil
}

func (m *Mexc2) createSpotTick(wsData *pb.PushDataV3ApiWrapper, bufferData *market.MexcTokenMeta) *market.MexcSpotTick {
	wsTick := wsData.GetPublicAggreBookTicker()

	return &market.MexcSpotTick{
		Symbol:   *wsData.Symbol,

		Volume:   bufferData.Volume,
		Deposit:  bufferData.Deposit,
		Withdraw: bufferData.Withdraw,
		Contract: bufferData.Contract,

		BidPrice: wsTick.BidPrice,
		BidQty:   wsTick.BidQuantity,
		AskPrice: wsTick.AskPrice,
		AskQty:   wsTick.AskQuantity,
	}
}
// listeners
