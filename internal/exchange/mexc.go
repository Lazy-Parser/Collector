package exchange_internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Lazy-Parser/Collector/api"
	wsclient "github.com/Lazy-Parser/Collector/internal/adapter/ws"
	"github.com/Lazy-Parser/Collector/market"
	"github.com/Lazy-Parser/Collector/pb"
	"google.golang.org/protobuf/proto"
)

var (
	sub = `{
    "method": "SUBSCRIPTION",
    "params": [%s]
}`
	unsub = `{
    "method": "UNSUBSCRIPTION",
    "params": [%s]
}`
	pingMsg   = `{"method": "PING"}`
	volumeMin = 500_000.0
)

type operation int

const (
	Subscribe operation = iota
	Unsubscribe
)

// m.BufferLoop()
// m.ListenSpot()
type Mexc struct {
	clientSpot    *wsclient.Client
	clientFutures *wsclient.Client
	api           api.MexcAPI

	symbols map[string]struct{}
	quotes  map[string]struct{}

	// xIndex are used for the work with both short and full symbols efficiently
	buffer  map[string]*market.MexcTokenMeta // key - full symbol ('BTCUSDT')
	queue   map[string]operation             // key - full symbol
	queueMu sync.RWMutex
}

func NewMexc(api api.MexcAPI) (*Mexc, error) {
	clientSpot, err := wsclient.NewClientBuilder().
		SetConnectionString("wss://wbs-api.mexc.com/ws").
		SetSubTemplate(sub).
		SetUnsubTemplate(unsub).
		SetConnectionMaxChannels(30).
		SetSubscriptionMaxChannels(15).
		SetPintMsgString(pingMsg).
		SetChannel("spot@public.aggre.bookTicker.v3.api.pb@100ms@%s").
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create a mexc spot lient: %v", err)
	}

	clientFutures, err := wsclient.NewClientBuilder().
		SetConnectionString("wss://contract.mexc.com/edge").
		SetSubTemplate(futures_sub).
		SetUnsubTemplate(futures_unsub).
		SetChannel("%s").
		SetPintMsgString(futuresPingMsg).
		SetConnectionMaxChannels(30).
		SetSubscriptionMaxChannels(1).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create a mexc futures lient: %v", err)
	}

	// fetch all symbols here or not here (maybe in BufferLoop)
	symbols, quotes, err := fetchAllSymbols(api)
	if err != nil {
		return nil, err
	}

	return &Mexc{
		api:           api,
		clientSpot:    clientSpot,
		clientFutures: clientFutures,
		symbols:       symbols,
		quotes:        quotes,
		buffer:        make(map[string]*market.MexcTokenMeta),
		queue:         make(map[string]operation),
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

func (m *Mexc) Name() string {
	return "Mexc"
}

// general
// symbol must be normalized for Volume!
// Done
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
				m.queuePush(symbol, Unsubscribe)
			}
		} else {
			if volumeBiggerMin {
				if len(m.buffer) >= 300 { // limit for tests
					return
				}
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

// close all internal processes and call ctx.Done()
func (m *Mexc) StopAll(ctx context.Context) {
	ctx.Done()
	m.clientSpot.Disconnect()
	m.clientFutures.Disconnect()
}

// general

// buffer. TODO: make init buffer fetch
// Start in the main goroutine
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

	log.Println("Fetched first requests for buffer")
	m.queueBurst()
	log.Println("Trying to burst all queuec")

	go func() {
		ticker := time.NewTicker(time.Minute * 5)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.fetchVolumeAndUpdate(ctx)
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
				m.fetchDepositWithdrawAndUpdate(ctx)
				// do not call m.queueBurst() here, because only volume fetch addes commands to the queue
			}
		}
	}()

	return nil
}

// TODO: Make buffer Fetch for provided symbol (token) name
func (m *Mexc) BufferUpdate(ctx context.Context) error {
	// init requests
	err := m.fetchVolumeAndUpdate(ctx)
	if err != nil {
		return fmt.Errorf("buffer loop error: %v", err)
	}
	err = m.fetchDepositWithdrawAndUpdate(ctx)
	if err != nil {
		return fmt.Errorf("buffer loop error: %v", err)
	}
	m.queueBurst()

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

	return nil
}

// accept only short symbol
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

// Accept full symbol
func (m *Mexc) bufferDelete(symbol string) {
	delete(m.buffer, symbol)
}

// Accept only full symbol
func (m *Mexc) bufferCreate(symbol string, update market.MexcTokenMetaUpdate) {
	if symbol == "" {
		return
	}

	m.buffer[symbol] = &market.MexcTokenMeta{Volume: *update.Volume}
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

// BTC|USDT -> if symbol.Contains(quotes.ForEach()) -> return (symbol - quote[i]) + "_" + quote[i]

// buffer

// queue
// accept onlu full symbol
func (m *Mexc) queuePush(symbol string, oper operation) {
	m.queueMu.Lock()
	defer m.queueMu.Unlock()

	m.queue[symbol] = oper
}

// Subscribe / unsubscribe symbols based on the queue
func (m *Mexc) queueBurst() {
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

	// if !m.client. {
	// 	log.Println("failed to burst all tasks in mexc queue, because the connection is not active")
	// 	return
	// }

	m.clientSpot.Unsubscribe(unsubList)
	m.clientSpot.Subscribe(subList)

	// empty queue
	for key := range m.queue {
		delete(m.queue, key)
	}
}

// queue

// listeners

// blocking
func (m *Mexc) ListenSpot(ctx context.Context, ch chan *market.MexcSpotTick) error {
	// non-blocking
	for {
		select {
		case <-ctx.Done():
			return nil
		case raw, ok := <-m.clientSpot.Listen():
			if !ok {
				return nil
			}
			if raw == nil {
				continue
			}

			// unmarshall
			wrapper := &pb.PushDataV3ApiWrapper{}
			if err := proto.Unmarshal(*raw, wrapper); err != nil {
				log.Println(string(*raw))
				continue
			}

			bufferData, ok := m.bufferFind(wrapper.Channel)
			if !ok {
				continue
			}

			ch <- m.createSpotTick(wrapper, bufferData)
		}
	}
}

func (m *Mexc) createSpotTick(wsData *pb.PushDataV3ApiWrapper, bufferData *market.MexcTokenMeta) *market.MexcSpotTick {
	wsTick := wsData.GetPublicAggreBookTicker()

	return &market.MexcSpotTick{
		Symbol: *wsData.Symbol,

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

func (m *Mexc) ListenFutures(ctx context.Context, ch chan *market.MexcFutureTick) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case raw, ok := <-m.clientFutures.Listen():
			if !ok {
				return nil
			}
			if raw == nil {
				continue
			}

			// unmarshall
			res := &market.MexcFutureTickWS{}
			if err := json.Unmarshal(*raw, res); err != nil {
				log.Println(string(*raw))
				continue
			}

			bufferData, ok := m.bufferFind(wrapper.Channel)
			if !ok {
				continue
			}

			ch <- m.createFutureTick(res)
		}
	}
}

func (m *Mexc) createFutureTick(data *market.MexcFutureTickWS) *market.MexcFutureTick {
	res := 
}

// listeners
