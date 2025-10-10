package api_internal

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"github.com/Lazy-Parser/Collector/config"
	"github.com/Lazy-Parser/Collector/market"
	"github.com/go-resty/resty/v2"
)

type MexcApi struct {
	cfg *config.Config
}

func NewMexcApi(cfg *config.Config) *MexcApi {
	return &MexcApi{cfg: cfg}
}

// Returns a list of all spot tokens with their info
func (api *MexcApi) FetchCurrencyInformation(ctx context.Context) ([]market.MexcAsset, error) {
	qs := url.Values{"timestamp": {fmt.Sprint(time.Now().UnixMilli())}}
	mac := hmac.New(sha256.New, []byte(api.cfg.Mexc.PRIVATE_TOKEN))
	mac.Write([]byte(qs.Encode()))
	qs.Set("signature", hex.EncodeToString(mac.Sum(nil)))

	var res []market.MexcAsset
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetQueryString(qs.Encode()).
		SetResult(&res).
		SetHeader("X-MEXC-APIKEY", api.cfg.Mexc.ACCESS_TOKEN).
		Get(api.cfg.Mexc.API.CONFIG_GETALL)
	return res, err
}

// Returns a list of futures contracts
func (api *MexcApi) FetchContractInformation(ctx context.Context) ([]market.MexcContractDetail, error) {
	var res market.MexcContracts
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetResult(&res).
		Get(api.cfg.Mexc.API.CONTRACTS_DETAIL)
	return res.Data, err
}

func (api *MexcApi) Fetch24hTickerStats(ctx context.Context) ([]market.MexcTickerStats, error) {
	var res []market.MexcTickerStats
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetResult(&res).
		Get(api.cfg.Mexc.API.TICKER_24HR)
	return res, err
}

func (api *MexcApi) FetchOrderBookTicker(ctx context.Context) ([]market.MexcOrderBookTick, error) {
	var res []market.MexcOrderBookTick
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetResult(&res).
		Get(api.cfg.Mexc.API.TICKER_24HR)
	return res, err
}

func (api *MexcApi) FetchExchangeInfo(ctx context.Context) ([]market.MexcExchangeInfo, error) {
	var res market.MexcExchangeInfoRes
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetResult(&res).
		Get(api.cfg.Mexc.API.EXCHANGE_INFO)
	return res.Symbols, err
}
