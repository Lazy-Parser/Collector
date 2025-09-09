package exchange_internal

import (
	"context"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Lazy-Parser/Collector/api"
	wsclient "github.com/Lazy-Parser/Collector/internal/adapter/ws"
	"github.com/Lazy-Parser/Collector/market"
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
	ping      = `{"method": "PING"}`
	volumeMin = 50_000.0
)

type Operation int

const (
	Update Operation = iota
	Create
	Delete
	Pass
)

type Mexc struct {
	mu     sync.RWMutex
	api    api.MexcAPI
	buffer map[string]*market.MexcTokenMeta
	conn   *wsclient.Client
}

func NewMexc(api api.MexcAPI) (*Mexc, error) {
	config := wsclient.NewClientConfig()
	config.SubTemplate = sub
	config.UnsubTamplate = unsub
	config.SubscriptionMaxChannels = 25
	config.UrlConnection = "wss://wbs-api.mexc.com/ws"

	return &Mexc{
		api:    api,
		conn:   wsclient.NewClient(config),
		buffer: make(map[string]*market.MexcTokenMeta),
	}, nil
}

func (m *Mexc) Name() string {
	return "Mexc"
}

// this is a loop, that periodically make requests to mexc api to update info in buffer (like volume, withdraw / deposit)
func (m *Mexc) BufferLoop(ctx context.Context) error {
	// volume - every 5 min
	// deposit / withdraw - every 30 min

	// initial update start
	confs, stats, err := m.fetchForBuffer(ctx)
	if err != nil {
		return err
	}

	// update volume first (because it creates new TokenMeta)
	for _, stat := range *stats {
		m.updateBufferVolume(normalizeSymbol(stat.Symbol), stat.Volume)
	}
	for _, c := range *confs {
		n := c.NetworkList[0]
		m.updateBufferDepositWithdraw(
			c.Coin,
			n.WithdrawFee,
			c.NetworkList[0].DepositEnable,
			c.NetworkList[0].WithdrawEnable,
		)
	}
	// initial update end

	go func() {
		stats, err := m.api.Fetch24hTickerStats(ctx)
		if err != nil {
			// do smth
		}

		for _, stat := range stats {
			// update buffer
			oper := m.updateBufferVolume(normalizeSymbol(stat.Symbol), stat.Volume)
			// add / remove channel in the ws
			if err := m.updateSubscription(stat.Symbol, oper); err != nil {
				log.Printf("Failed to update subscription: %v", err)
			}
		}

		time.Sleep(time.Minute * 5)
	}()

	go func() {
		confs, err := m.api.FetchCurrencyInformation(ctx)
		if err != nil {
			// do smth
		}

		for _, c := range confs {
			n := c.NetworkList[0]
			m.updateBufferDepositWithdraw(
				c.Coin,
				n.WithdrawFee,
				c.NetworkList[0].DepositEnable,
				c.NetworkList[0].WithdrawEnable,
			)
		}

		time.Sleep(time.Minute * 30)
	}()

	return nil
}

func (m *Mexc) BufferToSubscription() {
	
}

// do not normalize symbol
func (m *Mexc) updateSubscription(symbol string, operation Operation) error {
	if symbol == "" {
		return nil
	}

	// make only for futures
	if m.conn.IsRunning() {
		if operation == Delete {
			if err := m.conn.Unsubscribe(symbol); err != nil {
				return err
			}
		}
		if operation == Create {
			if err := m.conn.Subscribe([]string{symbol}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Mexc) GetBuffer() *map[string]*market.MexcTokenMeta {
	return &m.buffer
}

func (m *Mexc) fetchForBuffer(ctx context.Context) (*[]market.MexcAsset, *[]market.MexcTickerStats, error) {
	var wg sync.WaitGroup
	errs := make([]error, 2)
	var tokensConf []market.MexcAsset
	var tokensStats []market.MexcTickerStats

	// make requests in parallel
	wg.Add(2)
	go func() {
		defer wg.Done()

		var err error
		tokensConf, err = m.api.FetchCurrencyInformation(ctx)
		if err != nil {
			errs[0] = err
		}
	}()
	go func() {
		defer wg.Done()

		var err error
		tokensStats, err = m.api.Fetch24hTickerStats(ctx)
		if err != nil {
			errs[1] = err
		}
	}()

	wg.Wait()

	if errs[0] != nil || errs[1] != nil {
		return nil, nil, errors.Join(errs[0], errs[1])
	}

	return &tokensConf, &tokensStats, nil
}

// Do not forget to normalize symbol here!
func (m *Mexc) updateBufferVolume(symbol, volume string) Operation {
	m.mu.Lock()
	defer m.mu.Unlock()

	volumeInt, _ := strconv.ParseFloat(volume, 32)
	if m.existsInBuffer(symbol) {
		if volumeInt < volumeMin {
			// if less then minimum - delete
			delete(m.buffer, symbol)
			return Delete
		} else {
			// if bigger then minimum - update
			m.buffer[symbol].Volume = volume
			return Update
		}
	} else {
		if volumeInt > volumeMin {
			// if does not exists and bigger then minimum - create
			m.buffer[symbol] = &market.MexcTokenMeta{Volume: volume}
			return Create
		} else {
			return Pass
		}
	}
}

// remove quote 'usdt' prefix
func normalizeSymbol(symbol string) string {
	if len(symbol) < 5 {
		return symbol
	}
	return symbol[:len(symbol)-4]
}

func (m *Mexc) updateBufferDepositWithdraw(symbol, withdrawFee string, deposit, withdraw bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// this func do not create new TokenMeta! Only updateBufferVolume can
	if !m.existsInBuffer(symbol) {
		return
	}

	m.buffer[symbol].Deposit = deposit
	m.buffer[symbol].Withdraw = withdraw
	m.buffer[symbol].WithdrawFee = withdrawFee
}

// private
func (m *Mexc) existsInBuffer(symbol string) bool {
	_, ok := m.buffer[symbol]
	return ok
}

func (m *Mexc) ListenSpot(ch chan market.MexcSpotTick) error {
	err := m.conn.Connect()
	if err != nil {
		return err
	}

	m.conn.PingLoop(ping, time.Second*30)
	if err := m.conn.Run(); err != nil {
		return err
	}

	if err := m.conn.Subscribe([]string{}); err != nil {
		return err
	}

	return nil
}

// TODO: not finished yet
func (m *Mexc) ListenFutures(ch chan market.MexcFutureTick) {

}
