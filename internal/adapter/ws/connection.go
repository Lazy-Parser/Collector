package wsclient

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Lazy-Parser/Collector/pb"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type State int

const (
	Inited State = iota
	Connected
	Running
	Disconnected
)

type Connection struct {
	conn     *websocket.Conn
	subs     []*Subscription
	listenCh chan *pb.PushDataV3ApiWrapper
	closeCh  chan struct{}
	state    State
	mu       sync.RWMutex

	toUpdateIdx map[int]struct{} // indexes of subscriptions to update
	toRemove    []string

	connectionString        string
	connectionMaxChannels   int
	subscriptionMaxChannels int
	subTemplate             string
	unsubTemplate           string
}

// create new websocket connection
func NewConnection(connectionString string, connectionMaxChannels int, subscriptionMaxChannels int, subTemplate, unsubTemplate string) (*Connection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial connection: %v", err)
	}

	return &Connection{
		conn:                    conn,
		listenCh:                make(chan *pb.PushDataV3ApiWrapper, 1024),
		subs:                    []*Subscription{NewSubscription(subscriptionMaxChannels)},
		closeCh:                 make(chan struct{}),
		state:                   Connected,
		connectionString:        connectionString,
		connectionMaxChannels:   connectionMaxChannels,
		subscriptionMaxChannels: subscriptionMaxChannels,
		subTemplate:             subTemplate,
		unsubTemplate:           unsubTemplate,
		toUpdateIdx:             map[int]struct{}{},
	}, nil
}

// a ping loop
func (c *Connection) HeartBeat(msg string, interval time.Duration) error {
	if c.conn == nil {
		return nil
	}

	payload := []byte(msg)
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-c.closeCh:
				return
			case <-ticker.C:
				if err := c.saveWriteMessage(websocket.TextMessage, payload); err != nil {
					panic(err)
					// return
				}
			}
		}
	}()

	return nil
}

func (c *Connection) Run() error {
	c.state = Running

	defer close(c.listenCh)
	for {
		select {
		case <-c.closeCh:
			return nil

		default:
			msgType, msg, err := c.conn.ReadMessage()
			if err != nil {
				// try to reconnect
				c.state = Disconnected
				if err := c.reconnect(); err != nil {
					// failed to reconnect
					return err
				}

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

			c.listenCh <- wrapper
		}
	}
}

func (c *Connection) Listen() <-chan *pb.PushDataV3ApiWrapper {
	return c.listenCh
}

// write provided channel to the internal queue and list. Call 'Flush()' to Subscribe/Unsubscribe all channels in the queue
func (c *Connection) TrySubscribe(channel string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, sub := range c.subs {
		if sub.Exists(channel) {
			return true
		}

		if sub.IsFull() {
			continue
		}

		sub.Push(channel)
		c.toUpdateIdx[i] = struct{}{}

		return true
	}

	// check if we can add a new subscription
	if len(c.subs)*c.subscriptionMaxChannels < c.connectionMaxChannels {
		sub := NewSubscription(c.subscriptionMaxChannels)
		sub.Push(channel)

		c.subs = append(c.subs, sub)
		c.toUpdateIdx[len(c.subs)-1] = struct{}{}
		return true
	}

	return false
}

func (c *Connection) TryUnsubscribe(channel string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.subs {
		if c.subs[i].TryRemove(channel) {
			c.toRemove = append(c.toRemove, channel)

			if c.subs[i].Size() == 0 {
				c.deleteSubscription(i)
			}

			return true
		}
	}

	return false
}

func (c *Connection) FlushSub() error {
	log.Println("Trying to flush")
	log.Printf("Lenght: %d, toUpdateLen: %+v, toUpdate: %+v", len(c.subs), len(c.toUpdateIdx), c.toUpdateIdx)

	for idx := range c.toUpdateIdx {
		// unsubscribe first. It also will add new channels to the payload, but it's ok, there should not be error
		sub := c.subs[idx]
		payload := []byte(c.getUnsubPayload(sub.GetChannelsList()))
		if err := c.saveWriteMessage(websocket.TextMessage, payload); err != nil {
			c.toUpdateIdx = map[int]struct{}{}
			return err
		}

		// time.Sleep(time.Millisecond * 1000)

		// resubscribe with new channels
		payload = []byte(c.getSubPayload(sub.GetChannelsList()))
		if err := c.saveWriteMessage(websocket.TextMessage, payload); err != nil {
			c.toUpdateIdx = map[int]struct{}{}
			return err
		}
	}

	// do not forger to clear arr
	c.toUpdateIdx = map[int]struct{}{}
	return nil
}

func (c *Connection) FlushUnsub() error {
	payload := []byte(c.getUnsubPayload(c.toRemove))
	err := c.saveWriteMessage(websocket.TextMessage, payload)
	c.toRemove = make([]string, 0)
	return err
}

func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.unsubscribeAll(); err != nil {
		return err
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	c.state = Disconnected

	c.closeCh <- struct{}{}

	return nil
}

func (c *Connection) GetChannelsAmount() int {
	counter := 0
	for i := range c.subs {
		counter += c.subs[i].Size()
	}
	return counter
}

// private

// Write message with [mu.Lock]. Reconnection on write msg error, if reconnection error - return error.
//
// If reconnection success - try to resend message one more time, if error - return error
func (c *Connection) saveWriteMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == Disconnected {
		return nil
	}

	log.Println(string(data))
	if err := c.conn.WriteMessage(messageType, data); err != nil {
		// try to reconnect
		log.Println("Reconnection...")

		if err := c.reconnect(); err != nil {
			// failed  to reconnect
			c.state = Disconnected
			return err
		}

		// try to resend message
		if err := c.conn.WriteMessage(messageType, data); err != nil {
			c.state = Disconnected
			return err
		}

		c.state = Running
	}

	return nil
}

func (c *Connection) reconnect() error {
	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(c.connectionString, nil)
	if err != nil {
		return err
	}

	c.state = Running

	for _, sub := range c.subs {
		payload := c.getSubPayload(sub.GetChannelsList())
		if err := c.conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Connection) unsubscribeAll() error {
	for i := range c.subs {
		if c.subs[i].Size() == 0 {
			continue
		}

		payload := []byte(c.getUnsubPayload(c.subs[i].GetChannelsList()))
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			return err
		}
	}

	return nil
}

func (c *Connection) deleteSubscription(i int) {
	if i < 0 || i >= len(c.subs) {
		return
	}

	c.subs[i] = nil
	copy(c.subs[i:], c.subs[i+1:])
	c.subs[len(c.subs)-1] = nil // avoid memory leak
	c.subs = c.subs[:len(c.subs)-1]
}

func (c *Connection) getSubPayload(channels []string) string {
	return fmt.Sprintf(c.subTemplate, c.channelsToString(channels))
}

func (c *Connection) getUnsubPayload(channels []string) string {
	return fmt.Sprintf(c.unsubTemplate, c.channelsToString(channels))
}

func (c *Connection) channelsToString(channels []string) string {
	var str string

	for i, channel := range channels {
		str += "\"" + channel + "\""
		if i != len(channels)-1 {
			str += ", "
		}
	}

	return str
}
