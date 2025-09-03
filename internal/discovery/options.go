package discovery_internal

import "github.com/Lazy-Parser/Collector/market"

type Option func(*Discovery)

// Option that checkes in method [Meta] if discovered pools are unique in database
func (d *Discovery) WithOnlyNew(enable bool, tokenRepo market.TokenRepo, pairRepo market.PairRepo) Option {
	return func(d *Discovery) {
		d.onlyNewMode = enable
		d.tokenRepo = tokenRepo
		d.pairRepo = pairRepo
	}
}

// Option that fetch decimal for token in [Meta] method.
// This is an option, because in coingecko there are few requests per month. So to save some requests, it is better to group tokens by network and send just one request with up to 30 tokens
func (d *Discovery) WithDecimals(enable bool) Option {
	return func(d *Discovery) {
		d.decimalsMode = enable
	}
}

func (d *Discovery) ApplyOptions(opts ...Option) {
	for _, opt := range opts {
		opt(d)
	}
}
