// package wsclient

// import (
// 	"errors"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"

// 	"github.com/Lazy-Parser/Collector/pb"
// 	"github.com/gorilla/websocket"
// 	"google.golang.org/protobuf/proto"
// )

// type SubUnsubMsg struct {
// 	Sub   string
// 	Unsub string
// }

// type Config struct {
// 	UrlConnection           string
// 	SubTemplate             string
// 	UnsubTamplate           string
// 	ChannelTemplate         string
// 	SubscriptionMaxChannels int
// 	ReconnectAttempts       int
// 	ReconnectionBackoff     time.Duration // how long to wait between disconnection and recconection
// }

// func NewClientConfig() Config {
// 	return Config{
// 		UrlConnection:           "wss://wbs-api.mexc.com/ws",
// 		SubTemplate:             `{"method": "SUBSCRIBE", "params": ["%s"]}`,
// 		UnsubTamplate:           `{"method": "UNSUBSCRIBE", "params": ["%s"]}`,
// 		ChannelTemplate:         "spot@public.aggre.bookTicker.v3.api.pb@100ms%s",
// 		SubscriptionMaxChannels: 25,
// 		ReconnectAttempts:       5,
// 		ReconnectionBackoff:     time.Second * 5,
// 	}
// }

// type State int

// const (
// 	Running State = iota
// 	Disconnected
// 	Reconnection
// 	None
// )

// // TODO: there is a limit 30 subs per connection. So i need to implement a dinamic list of connections
// type Client struct {
// 	mu     sync.Mutex
// 	conn   *websocket.Conn
// 	config Config
// 	subs   []Subscription

// 	msgCh  chan *pb.PushDataV3ApiWrapper
// 	doneCh chan struct{}

// 	// state
// 	state State

// 	reconnectionCounter int
// }

// func NewClient(config Config) *Client {
// 	return &Client{
// 		conn:                nil,
// 		config:              config,
// 		subs:                []Subscription{NewSubscription(config.SubscriptionMaxChannels)},
// 		doneCh:              make(chan struct{}),
// 		msgCh:               make(chan *pb.PushDataV3ApiWrapper, 2048),
// 		state:               None,
// 		reconnectionCounter: 0,
// 	}
// }

// func (c *Client) Connect() error {
// 	conn, _, err := websocket.DefaultDialer.Dial(c.config.UrlConnection, nil)
// 	if err != nil {
// 		return err
// 	}
// 	c.conn = conn
// 	c.state = Running

// 	return nil
// }

// func (c *Client) Close() error {
// 	close(c.doneCh)
// 	if err := c.unsubscribeAll(); err != nil {
// 		return err
// 	}

// 	return c.conn.Close()
// }

// func (c *Client) Subscribe(channels []string) error {
// 	updatedSubs := make(map[int]struct{})

// channLoop:
// 	for _, channel := range channels {
// 		// TODO: symbol 'BTCUSDT' -> channel 'ticker@100ms@BTCUSDT' should not be here
// 		channel = fmt.Sprintf(c.config.ChannelTemplate, channel)

// 		// check if already contains
// 		for _, sub := range c.subs {
// 			if sub.Contains(channel) {
// 				continue channLoop
// 			}
// 		}

// 		idx := -1 // the idx of target subscription
// 		for i := range c.subs {
// 			if !c.subs[i].IsFull() {
// 				idx = i
// 				break
// 			}
// 		}

// 		// if not found, create a new sub
// 		if idx == -1 {
// 			c.subs = append(c.subs, NewSubscription(c.config.SubscriptionMaxChannels))
// 			idx = len(c.subs) - 1
// 		}

// 		// add to the list
// 		if c.subs[idx].Push(channel) {
// 			updatedSubs[idx] = struct{}{}
// 		}
// 	}

// 	// now refresh only updated subscriptions
// 	for idx := range updatedSubs {
// 		// unsubscribe first. It also will add new channels to the payload, but it's ok, there should not be error
// 		sub := c.subs[idx]
// 		payload := []byte(c.getUnsubPayload(sub.GetChannelsList()))
// 		if err := c.saveWriteMessage(websocket.TextMessage, payload); err != nil {
// 			return err
// 		}

// 		// resubscribe with new channels
// 		payload = []byte(c.getSubPayload(sub.GetChannelsList()))
// 		if err := c.saveWriteMessage(websocket.TextMessage, payload); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (c *Client) Unsubscribe(channels []string) error {
// 	var toRemove []string
// 	for _, channel := range channels {
// 		channel = fmt.Sprintf(c.config.ChannelTemplate, channel)
// 		if exists := c.channelExist(channel); !exists {
// 			return errors.New("failed to unsubscribe: channel not found in list: " + channel)
// 		}

// 		toRemove = append(toRemove, channel)
// 	}

// 	// remove from the local list
// 	for _, channel := range toRemove {
// 		c.channelRemove(channel)
// 	}

// 	payload := []byte(c.getUnsubPayload(toRemove))
// 	return c.saveWriteMessage(websocket.TextMessage, payload)
// }

// func (c *Client) Run() error {
// 	defer close(c.msgCh)
// 	for {
// 		select {
// 		case <-c.doneCh:
// 			return nil

// 		default:
// 			msgType, msg, err := c.conn.ReadMessage()
// 			if err != nil {
// 				log.Println(err)
// 				// try to reconnect
// 				if err := c.tryReconnect(); err != nil {
// 					return fmt.Errorf("reconnection failed: %v", err)
// 				}
// 				continue
// 			}

// 			if msgType == websocket.PongMessage {
// 				continue
// 			}

// 			wrapper := &pb.PushDataV3ApiWrapper{}
// 			if err := proto.Unmarshal(msg, wrapper); err != nil {
// 				log.Println(string(msg))
// 				continue
// 			}

// 			c.msgCh <- wrapper
// 		}
// 	}
// }

// func (c *Client) ListenTicks() <-chan *pb.PushDataV3ApiWrapper {
// 	return c.msgCh
// }

// // msg - a message to send as "ping". Interval - how often to send pings. {"method": "PING"} for Mexc.
// //
// // Importnant! Do not start this func in goroutine
// func (c *Client) PingLoop(msg string, interval time.Duration) error {
// 	if interval <= 0 {
// 		return fmt.Errorf("provided interval is <= 0")
// 	}

// 	payload := []byte(msg)
// 	ticker := time.NewTicker(interval)
// 	go func() {
// 		defer ticker.Stop()
// 		for {
// 			select {
// 			case <-ticker.C:
// 				// todo: check if connection is closed
// 				if err := c.saveWriteMessage(websocket.TextMessage, payload); err != nil {
// 					log.Println(err)
// 				}

// 			case <-c.doneCh:
// 				return
// 			}
// 		}
// 	}()

// 	return nil
// }

// func (c *Client) IsRunning() bool {
// 	return c.state == Running
// }

// func (c *Client) GetSubs() int {
// 	return len(c.subs)
// }

// func (c *Client) SubsToString() string {
// 	var str string

// 	for i, sub := range c.subs {
// 		str += fmt.Sprintf("SUBSCRIPTION #%d. Channels: %s\n", i+1, sub.ToString())
// 	}

// 	return str
// }

// // private

// // on error - try to reconnect
// func (c *Client) saveWriteMessage(messageType int, data []byte) error {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()

// 	if c.state == Disconnected || c.conn == nil {
// 		return errors.New("failed to write msg to conn when disconnected or nil")
// 	}
// 	if c.state == Reconnection {
// 		return nil
// 	}

// 	if err := c.conn.WriteMessage(messageType, data); err != nil {
// 		// try to reconnect
// 		if err := c.tryReconnect(); err != nil {
// 			return fmt.Errorf("failed to reconnect: %v", err)
// 		}
// 	}

// 	return nil
// }

// func (c *Client) tryReconnect() error {
// 	c.mu.Lock()
// 	if c.reconnectionCounter >= c.config.ReconnectAttempts {
// 		return errors.New("failed to reconnect: all attempts are spend (" + fmt.Sprintf("%d", c.config.ReconnectAttempts) + ")")
// 	}

// 	c.reconnectionCounter++
// 	c.state = Reconnection
// 	log.Println("Try to reconnect!")

// 	if c.conn != nil {
// 		c.conn.Close()
// 	}

// 	var err error
// 	c.conn, _, err = websocket.DefaultDialer.Dial(c.config.UrlConnection, nil)
// 	c.mu.Unlock()
// 	if err != nil {
// 		c.state = Disconnected
// 		return err
// 	}

// 	// resubscribe
// 	time.Sleep(c.config.ReconnectionBackoff)
// 	for i := range c.subs {
// 		channelsCopy := c.subs[i].GetChannelsList()
// 		c.subs[i].ClearChannels() // do not forget to clear subscription's channels
// 		if err := c.Subscribe(channelsCopy); err != nil {
// 			c.state = Disconnected
// 			return err
// 		}
// 	}

// 	c.state = Running
// 	return nil
// }

// func (c *Client) getSubPayload(channels []string) string {
// 	// create

// 	return fmt.Sprintf(c.config.SubTemplate, c.channelsToString(channels))
// }

// func (c *Client) getUnsubPayload(channels []string) string {
// 	return fmt.Sprintf(c.config.UnsubTamplate, c.channelsToString(channels))
// }

// func (c *Client) channelsToString(channels []string) string {
// 	var str string

// 	for i, channel := range channels {
// 		str += wrapInQuotes(channel)
// 		if i != len(channels)-1 {
// 			str += ", "
// 		}
// 	}

// 	return str
// }

// func (c *Client) unsubscribeAll() error {
// 	for _, sub := range c.subs {
// 		payload := []byte(c.getUnsubPayload(sub.GetChannelsList()))
// 		if err := c.saveWriteMessage(websocket.TextMessage, payload); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (c *Client) channelExist(channel string) bool {
// 	for _, sub := range c.subs {
// 		if sub.Exists(channel) {
// 			return true
// 		}
// 	}

// 	return false
// }

// func (c *Client) channelRemove(channel string) {
// 	for _, sub := range c.subs {
// 		sub.TryRemove(channel)
// 	}
// }

// // TODO: remove
// func (c *Client) MockDisconnect() {
// 	c.conn.Close()
// 	c.conn = nil
// }
