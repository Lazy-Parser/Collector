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

func (api *MexcApi) FetchContractInformation(ctx context.Context) ([]market.MexcContractDetail, error) {
	var res market.MexcContracts
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetResult(&res).
		Get(api.cfg.Mexc.API.CONTRACTS_DETAIL)
	return res.Data, err
}
