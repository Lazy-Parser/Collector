package wsclient

import "errors"

type ClientBuidler struct {
	connectionString        string
	subTemplate             string
	unsubTemplate           string
	channel                 string
	connectionMaxChannels   int
	subscriptionMaxChannels int
}

func (b *ClientBuidler) SetConnectionString(str string) *ClientBuidler {
	b.connectionString = str
	return b
}

func (b *ClientBuidler) SetSubTemplate(str string) *ClientBuidler {
	b.subTemplate = str
	return b
}

func (b *ClientBuidler) SetUnsubTemplate(str string) *ClientBuidler {
	b.unsubTemplate = str
	return b
}

// a channel tamplate, for example: 'spot@public.aggre.bookTicker.v3.api.pb@100ms@%s'
//
// %s - for symbol, do not forget to pass '%s'!
func (b *ClientBuidler) SetChannel(str string) *ClientBuidler {
	b.channel = str
	return b
}

func (b *ClientBuidler) SetConnectionMaxChannels(max int) *ClientBuidler {
	b.connectionMaxChannels = max
	// make subscription limit the same by default
	b.subscriptionMaxChannels = max
	return b
}

func (b *ClientBuidler) SetSubscriptionMaxChannels(max int) *ClientBuidler {
	b.subscriptionMaxChannels = max
	return b
}

func (b *ClientBuidler) Build() (*Client, error) {
	if b.connectionMaxChannels%b.subscriptionMaxChannels != 0 {
		return nil, errors.New("connectionMaxChannels must be divisible by subscriptionMaxChannels. The number of subscriptions should fit perfectly into the connection")
	}

	conn, err := NewConnection(b.connectionString, b.connectionMaxChannels, b.subscriptionMaxChannels, b.subTemplate, b.unsubTemplate)
	if err != nil {
		return nil, errors.New("failed to create connection: " + err.Error())
	}

	return &Client{
		connectionString:        b.connectionString,
		subTemplate:             b.subTemplate,
		unsubTemplate:           b.unsubTemplate,
		channel:                 b.channel,
		connectionMaxChannels:   b.connectionMaxChannels,
		subscriptionMaxChannels: b.subscriptionMaxChannels,
		conns:                   []*Connection{conn},
	}, nil
}
