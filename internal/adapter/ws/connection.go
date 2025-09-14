package wsclient

import (
	"fmt"
	"log"
	"time"

	"github.com/Lazy-Parser/Collector/pb"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// TODO: PING LOOP
type Connection struct {
	conn     *websocket.Conn
	subs     []*Subscription
	listenCh chan *pb.PushDataV3ApiWrapper
	closeCh  chan struct{}

	toUpdateIdx []int // indexes of subscriptions to update
	toRemove    []string

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
		connectionMaxChannels:   connectionMaxChannels,
		subscriptionMaxChannels: subscriptionMaxChannels,
		subTemplate:             subTemplate,
		unsubTemplate:           unsubTemplate,
		toUpdateIdx:             make([]int, 0),
	}, nil
}

func (c *Connection) PingCycle(msg string, interval time.Duration) error {
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
				if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}()

	return nil
}

func (c *Connection) Run() error {
	defer close(c.listenCh)
	for {
		select {
		case <-c.closeCh:
			return nil

		default:
			msgType, msg, err := c.conn.ReadMessage()
			if err != nil {
				return err
			}
			if msgType == websocket.PongMessage {
				continue
			}

			wrapper := &pb.PushDataV3ApiWrapper{}
			if err := proto.Unmarshal(msg, wrapper); err != nil {
				return err
			}

			c.listenCh <- wrapper
		}
	}
}

func (c Connection) Listen() <-chan *pb.PushDataV3ApiWrapper {
	return c.listenCh
}

// write provided channel to the internal queue and list. Call 'Flush()' to Subscribe/Unsubscribe all channels in the queue
func (c *Connection) TrySubscribe(channel string) bool {
	for i, sub := range c.subs {
		if sub.Exists(channel) {
			return true
		}

		if sub.IsFull() {
			continue
		}

		sub.Push(channel)
		c.toUpdateIdx = append(c.toUpdateIdx, i)

		return true
	}

	// check if we can add a new subscription
	if len(c.subs)*c.subscriptionMaxChannels < c.connectionMaxChannels {
		sub := NewSubscription(c.subscriptionMaxChannels)
		sub.Push(channel)

		c.subs = append(c.subs, sub)
		c.toUpdateIdx = append(c.toUpdateIdx, len(c.subs) - 1)
		return true
	}

	return false
}

func (c *Connection) TryUnsubscribe(channel string) bool {
	for i := range c.subs {
		if c.subs[i].TryRemove(channel) {
			c.toRemove = append(c.toRemove, channel)
			return true
		}
	}

	return false
}

func (c *Connection) FlushSub() error {
	for idx := range c.toUpdateIdx {
		// unsubscribe first. It also will add new channels to the payload, but it's ok, there should not be error
		sub := c.subs[idx]
		payload := []byte(c.getUnsubPayload(sub.GetChannelsList()))
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			c.toUpdateIdx = make([]int, 0)
			return err
		}

		// resubscribe with new channels
		payload = []byte(c.getSubPayload(sub.GetChannelsList()))
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			c.toUpdateIdx = make([]int, 0)
			return err
		}
	}

	// do not forger to clear arr
	c.toUpdateIdx = make([]int, 0)
	return nil
}

func (c *Connection) FlushUnsub() error {
	payload := []byte(c.getUnsubPayload(c.toRemove))
	if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		c.toRemove = make([]string, 0)
		return err
	}

	c.toRemove = make([]string, 0)
	return nil
}

func (c *Connection) Close() error {
	if err := c.unsubscribeAll(); err != nil {
		return err
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	c.closeCh <- struct{}{}

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

// private
func (c *Connection) getSubPayload(channels []string) string {
	return fmt.Sprintf(c.subTemplate, c.channelsToString(channels))
}

func (c *Connection) getUnsubPayload(channels []string) string {
	return fmt.Sprintf(c.unsubTemplate, c.channelsToString(channels))
}

func (c *Connection) channelsToString(channels []string) string {
	var str string

	for i, channel := range channels {
		str += wrapInQuotes(channel)
		if i != len(channels)-1 {
			str += ", "
		}
	}

	return str
}
