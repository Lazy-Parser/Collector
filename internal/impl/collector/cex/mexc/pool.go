package mexc

import (
	"context"
	"flag"
	"fmt"
	"github.com/Lazy-Parser/Collector/internal/utils"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type subscription struct {
	Method string `json:"method"`
	Param  param  `json:"param"`
}

type param struct {
	Symbol string `json:"symbol"`
}

type Connection struct {
	conn     *websocket.Conn
	subsSize int // how many subscriptions in this connection
}

type Pool struct {
	connections []*Connection
	dataFlow    chan []byte
}

var (
	subsMax     = flag.Int("subsMax", 25, "Maximum number of subscriptions per connection")
	pingTimeout = flag.Duration("pingTimeout", time.Second*10, "How long should this service wait to ping MEXC")
)

func CreatePool() *Pool {
	return &Pool{
		connections: []*Connection{},
		dataFlow:    make(chan []byte, 1000),
	}
}

func (p *Pool) Subscribe(ctx context.Context, token string) error {
	for _, connection := range p.connections {
		if connection.subsSize < *subsMax {
			if err := tryToSubscribe(connection.conn, token); err != nil {
				return err
			}

			// if subscription success, increment inner counter and return
			connection.subsSize++
			return nil
		}
	}

	// if no free connection available
	newConn, err := createConnection()
	if err != nil {
		return err
	}

	p.connections = append(p.connections, &Connection{newConn, 0}) // create a new connection
	if err := tryToSubscribe(newConn, token); err != nil {
		return err
	}

	go p.listenWS(ctx, newConn)
	go p.ping(ctx, newConn)

	return nil
}

func (p *Pool) Listen() <-chan []byte {
	return p.dataFlow
}

func (p *Pool) UnsubscribeAll() error {
	return nil // TODO: implement
}

// private methods
func createConnection() (*websocket.Conn, error) {
	connectionString, err := utils.LoadEnv("MEXC_FUTURES_WS")
	if err != nil {
		return nil, err
	}

	conn, _, err := websocket.DefaultDialer.Dial(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection to Mexc Futures: %v", err)
	}

	return conn, nil
}

func tryToSubscribe(conn *websocket.Conn, token string) error {
	payload := generateSubscriptionPayload(token) // generate subscribe payload
	fmt.Printf("%+v\n", payload)

	if err := conn.WriteJSON(payload); err != nil { // try to subscribe, if failed, return
		return fmt.Errorf("failed to subscribe token '%s' to Mexc Futures: %v", token, err)
	}

	return nil
}

func (p *Pool) listenWS(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			// unsubscribe(conn)
			return

		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return
			}
			p.dataFlow <- msg
		}
	}
}

func (p *Pool) ping(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(*pingTimeout)
	defer ticker.Stop()

	payload := map[string]interface{}{"method": "ping"}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := conn.WriteJSON(payload)
			if err != nil {
				fmt.Errorf("ping error: %w", err)
			}
		}
	}
}

func generateSubscriptionPayload(token string) interface{} {
	payload := subscription{
		Method: "sub.depth",
		Param: param{
			Symbol: token + "_USDT",
		},
	}

	return payload
}

func unsubscribe(conn *websocket.Conn) {
	// TODO: implement
}
