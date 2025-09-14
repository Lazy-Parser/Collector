package wsclient

import "fmt"

type Client struct {
	connectionString        string
	subTemplate             string
	unsubTemplate           string
	channel                 string
	connectionMaxChannels   int
	subscriptionMaxChannels int

	conns []*Connection
}

// do not create NewClient(), because already have builder in other filer

func (c *Client) Subscribe(symbols []string) error {
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
			return err
		}
		c.conns[len(c.conns)-1].TrySubscribe(channel)
	}

	for i := range c.conns {
		if err := c.conns[i].FlushSub(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Unsubscribe(symbols []string) error {
symbolLoop:
	for _, symbol := range symbols {
		channel := c.symbolToChannel(symbol)

		for i := range c.conns {
			if ok := c.conns[i].TryUnsubscribe(channel); ok {
				continue symbolLoop
			}
		}

		// do smth here if provided symbol was not subscribed (wasnot found in list)
	}

	for i := range c.conns {
		if err := c.conns[i].FlushUnsub(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Listen() {
	// write some code to combine all listeners from all connections and return just one
}

func (c *Client) Disconnect() error {
	// implement sooner
	return nil
}

// private
func (c *Client) symbolToChannel(symbol string) string {
	return fmt.Sprintf(c.channel, symbol)
}

func (c *Client) addConnection() error {
	newConn, err := NewConnection(c.connectionString, c.connectionMaxChannels, c.subscriptionMaxChannels, c.subTemplate, c.unsubTemplate, )
	if err != nil {
		return err
	}

	go newConn.Run()

	c.conns = append(c.conns, newConn)
	return nil
}
