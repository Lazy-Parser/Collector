package wsclient

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type SubUnsubMsg struct {
	Sub   string
	Unsub string
}

type Client struct {
	conn *websocket.Conn
	msgs []SubUnsubMsg

	doneCh chan struct{}
}

func NewClient(urlString string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(urlString, nil)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn, doneCh: make(chan struct{})}, nil
}

func (c *Client) Close() error {
	close(c.doneCh)
	if err := c.unsubscribeAll(); err != nil {
		return err
	}

	return c.conn.Close()
}

func (c *Client) Subscribe(msg SubUnsubMsg) error {
	c.msgs = append(c.msgs, msg)

	payload := []byte(msg.Sub)
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

func (c *Client) Unsubscribe(msg SubUnsubMsg) error {
	payload := []byte(msg.Unsub)
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

// msg - a message to send as "ping". Interval - how often to send pings. {"method": "PING"} for Mexc.
//
// Importnant! Do not start this func in goroutine
func (c *Client) PingLoop(msg string, interval time.Duration) error {
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
				if err := c.conn.WriteMessage(websocket.PingMessage, payload); err != nil {
					log.Println(err)
				}

			case <-c.doneCh:
				return
			}
		}
	}()

	return nil
}

func (c *Client) unsubscribeAll() error {
	payload := []byte(`{"method": "UNSUBSCRIBE"}`) // some str to unsubscribe
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

// TODO: make reconnection
// make pings in 30 secs
