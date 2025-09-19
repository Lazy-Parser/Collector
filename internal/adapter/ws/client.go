package wsclient

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type Client struct {
	connectionString        string
	subTemplate             string
	unsubTemplate           string
	channel                 string
	pingMsg                 string
	connectionMaxChannels   int
	subscriptionMaxChannels int

	conns    []*Connection
	listenCh chan *[]byte
	mu       sync.RWMutex
	errorCh  chan error
	wg       sync.WaitGroup
}

// do not create NewClient(), because already have builder in other filer

func (c *Client) Subscribe(symbols []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

symbolLoop:
	for _, symbol := range symbols {
		channel := c.symbolToChannel(symbol)

		for i := range c.conns {
			if ok := c.conns[i].TrySubscribe(channel); ok {
				continue symbolLoop
			}
		}

		// if no one from existing connection couldnot subscribe provided channel, then create new connection
		if err := c.addConnection(); err != nil {
			return fmt.Errorf("failed to create new connection: %v", err)
		}
		if ok := c.conns[len(c.conns)-1].TrySubscribe(channel); !ok {
			return errors.New(fmt.Sprintf("failed to subscribe to %s in a new connection", channel))
		}
	}

	for i := range c.conns {
		if err := c.conns[i].FlushSub(); err != nil {
			return fmt.Errorf("failed to flush subscriptions on connection %d: %w", i, err)
		}
	}

	return nil
}

func (c *Client) Unsubscribe(symbols []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, symbol := range symbols {
		channel := c.symbolToChannel(symbol)

		for i := range c.conns {
			if ok := c.conns[i].TryUnsubscribe(channel); ok {
				break
			}
		}
	}

	for i := range c.conns {
		if err := c.conns[i].FlushUnsub(); err != nil {
			return err
		}
	}

	// for i := len(c.conns) - 1; i >= 0; i-- {
	// 	if c.conns[i].GetChannelsAmount() == 0 {
	// 		if err := c.deleteConnection(i); err != nil {
	// 			return errors.New(fmt.Sprintf("Failed to delete empty connection %d: %v", i, err))
	// 		}
	// 	}
	// }

	return nil
}

func (c *Client) Listen() <-chan *[]byte {
	return c.listenCh
}

func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.conns {
		if err := c.conns[i].Close(); err != nil {
			return errors.New(fmt.Sprintf("Failed to close connection %d: %v", i, err))
		}
	}

	c.wg.Wait()

	if c.listenCh != nil {
		close(c.listenCh)
	}

	return nil
}

// private
func (c *Client) deleteConnection(i int) error {
	if i < 0 || i >= len(c.conns) {
		return nil
	}

	if err := c.conns[i].Close(); err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}

	copy(c.conns[i:], c.conns[i+1:])
	c.conns[len(c.conns)-1] = nil // avoid memory leak
	c.conns = c.conns[:len(c.conns)-1]

	return nil
}

func (c *Client) symbolToChannel(symbol string) string {
	return fmt.Sprintf(c.channel, symbol)
}

func (c *Client) addConnection() error {
	newConn, err := NewConnection(c.connectionString, c.connectionMaxChannels, c.subscriptionMaxChannels, c.subTemplate, c.unsubTemplate, c.pingMsg)
	if err != nil {
		return err
	}

	c.startConnection(newConn)

	c.conns = append(c.conns, newConn)
	return nil
}

func (c *Client) startConnection(conn *Connection) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := conn.Run(); err != nil {
			log.Println(err)
			return
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case msg, ok := <-conn.Listen():
				if !ok {
					log.Println("connection channel closed")
					return
				}

				c.listenCh <- msg
			}
		}
	}()

	if err := conn.HeartBeat(time.Second * 30); err != nil {
		log.Println(err)
		return
	}
}
