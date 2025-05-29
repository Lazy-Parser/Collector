package publisher

import "github.com/nats-io/nats.go"

type Publisher struct {
	nc *nats.Conn
}
