package wsclient

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Lazy-Parser/Collector/pb"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type SubUnsubMsg struct {
	Sub   string
	Unsub string
}

type Config struct {
	UrlConnection           string
	SubTemplate             string
	UnsubTamplate           string
	SubscriptionMaxChannels int
}

func NewClientConfig() Config {
	return Config{
		UrlConnection:           "wss://wbs-api.mexc.com/ws",
		SubTemplate:             `{"method": "SUBSCRIBE", "params": ["%s"]}`,
		UnsubTamplate:           `{"method": "UNSUBSCRIBE", "params": ["%s"]}`,
		SubscriptionMaxChannels: 25,
	}
}

type Client struct {
	mu     sync.Mutex
	conn   *websocket.Conn
	config Config
	subs   []Subscription

	msgCh  chan *pb.PushDataV3ApiWrapper
	doneCh chan struct{}
}

func NewClient(config Config) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(config.UrlConnection, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		config: config,
		subs:   []Subscription{NewSubscription(config.SubscriptionMaxChannels)},
		doneCh: make(chan struct{}),
		msgCh:  make(chan *pb.PushDataV3ApiWrapper, 1024),
	}, nil
}

func (c *Client) Close() error {
	close(c.doneCh)
	if err := c.unsubscribeAll(); err != nil {
		return err
	}

	return c.conn.Close()
}

func (c *Client) Subscribe(channels []string) error {
	updatedSubs := make(map[int]struct{})

channLoop:
	for _, channel := range channels {
		// check if already contains
		for _, sub := range c.subs {
			if sub.Contains(channel) {
				continue channLoop
			}
		}

		idx := -1 // the idx of target subscription
		for i := range c.subs {
			if !c.subs[i].IsFull() {
				idx = i
				break
			}
		}

		// if not found, create a new sub
		if idx == -1 {
			c.subs = append(c.subs, NewSubscription(c.config.SubscriptionMaxChannels))
			idx = len(c.subs) - 1
		}

		// add to the list
		if c.subs[idx].Push(channel) {
			updatedSubs[idx] = struct{}{}
		}
	}

	// now refresh only updated subscriptions
	for idx := range updatedSubs {
		// unsubscribe first. It also will add new channels to the payload, but it's ok, there should not be error
		sub := c.subs[idx]
		payload := []byte(c.getUnsubPayload(sub.ToString()))
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			return err
		}

		// resubscribe with new channels
		payload = []byte(c.getSubPayload(sub.ToString()))
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Unsubscribe(channel string) error {
	for _, sub := range c.subs {
		// try to remove from the local list
		if ok := sub.TryRemove(channel); ok {
			// if found in local list, unsubscribe from the connection
			payload := []byte(c.getUnsubPayload(channel))
			return c.conn.WriteMessage(websocket.TextMessage, payload)
		}
	}

	// if not found
	return errors.New("failed to unsubscribe: channel not found in list: " + channel)
}

func (c *Client) Run() {
	defer close(c.msgCh)
	for {
		select {
		case <-c.doneCh:
			return

		default:
			msgType, msg, err := c.conn.ReadMessage()
			if err != nil {
				log.Println(err)
				continue
			}

			if msgType == websocket.PongMessage {
				continue
			}

			wrapper := &pb.PushDataV3ApiWrapper{}
			if err := proto.Unmarshal(msg, wrapper); err != nil {
				log.Println(string(msg))
				continue
			}

			c.msgCh <- wrapper
		}
	}
}

func (c *Client) ListenTicks() <-chan *pb.PushDataV3ApiWrapper {
	return c.msgCh
}

// msg - a message to send as "ping". Interval - how often to send pings. {"method": "PING"} for Mexc.
//
// Importnant! Do not start this func in goroutine
func (c *Client) PingLoop(msg string, interval time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if interval <= 0 {
		return fmt.Errorf("provided interval is <= 0")
	}

	payload := []byte(msg)
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// todo: check if connection is closed
				if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
					log.Println(err)
				}

			case <-c.doneCh:
				return
			}
		}
	}()

	return nil
}

func (c *Client) GetSubs() int {
	return len(c.subs)
}

// private
func (c *Client) getSubPayload(channels string) string {
	return fmt.Sprintf(c.config.SubTemplate, channels)
}

func (c *Client) getUnsubPayload(channels string) string {
	return fmt.Sprintf(c.config.UnsubTamplate, channels)
}

func (c *Client) unsubscribeAll() error {
	for _, sub := range c.subs {
		payload := []byte(c.getUnsubPayload(sub.ToString()))
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			return err
		}
	}

	return nil
}

// TODO: make reconnection
// make pings in 30 secs
