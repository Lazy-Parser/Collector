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

	client := wsclient.NewClient(config)
	err := client.Connect()
	if err != nil {
		t.Error(err)
		return
	}

	go client.Run()
	client.PingLoop(`{"method": "PING"}`, 10*time.Second)

	payload := []string{
		channelStr("BTCUSDT"),
		channelStr("ETHUSDT"),
	}
	if err := client.Subscribe(payload); err != nil {
		t.Error(err)
		return
	}
	// control
	t.Log(client.SubsToString())

	go func() {
		time.Sleep(time.Second * 5)
		if err := client.Subscribe([]string{channelStr("FARTCOINUSDT")}); err != nil {
			panic(err)
		}
		// control
		t.Log(client.SubsToString())

		time.Sleep(time.Second * 5)
		if err := client.Unsubscribe(channelStr("FARTCOINUSDT")); err != nil {
			panic(err)
		}
		// control
		t.Log(client.SubsToString())

		time.Sleep(time.Second * 2)
		client.MockDisconnect()
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

func channelStr(symbol string) string {
	return "spot@public.aggre.bookTicker.v3.api.pb@100ms@" + symbol
}
