package wsclient_test

import (
	"fmt"
	"testing"
	"time"

	wsclient "github.com/Lazy-Parser/Collector/internal/adapter/ws"
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
)

func TestClientSubUnsub(t *testing.T) {
	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	config := wsclient.NewClientConfig()
	config.SubTemplate = sub
	config.UnsubTamplate = unsub
	config.SubscriptionMaxChannels = 1
	config.UrlConnection = "wss://wbs-api.mexc.com/ws"

	client, err := wsclient.NewClient(config)
	if err != nil {
		t.Error(err)
		return
	}
	go client.Run()
	client.PingLoop(`{"method": "PING"}`, 10*time.Second)

	payload := []string{
		"spot@public.aggre.bookTicker.v3.api.pb@100ms@BTCUSDT",
		"spot@public.aggre.bookTicker.v3.api.pb@100ms@ETHUSDT",
		"spot@public.aggre.bookTicker.v3.api.pb@100ms@FARTCOINUSDT",
	}
	if err := client.Subscribe(payload); err != nil {
		t.Error(err)
		return
	}

	go func() {
		// waitring 5 sec for sub
		time.Sleep(5 * time.Second)
		if client.GetSubs() != 3 {
			t.Errorf("expected 3 subs, got %d", client.GetSubs())
		} else {
			t.Log("Subscribed to 3 channels successfully")
			client.Close()
		}
	}()

	// blocking
	for msg := range client.ListenTicks() {
		body := msg.GetPublicAggreBookTicker()
		str := fmt.Sprintf("Symbol: %s | BID: %s | ASK: %s", *msg.Symbol, body.BidPrice, body.AskPrice)
		t.Log(str)
	}

	t.Log("Closing webscocket client...")
	if err := client.Close(); err != nil {
		t.Error(err)
		return
	}
}
