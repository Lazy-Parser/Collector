package api_mexc

import (
	config "github.com/Lazy-Parser/Collector/config/service"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

type MexcAPI struct {
}

func NewMexcAPI() *MexcAPI {
	return &MexcAPI{}
}

// All tokens
func (api *MexcAPI) FetchCurrencyInformation(ctx context.Context, cfg *config.Config) ([]Asset, error) {
	qs := url.Values{"timestamp": {fmt.Sprint(time.Now().UnixMilli())}}
	mac := hmac.New(sha256.New, []byte(cfg.Mexc.PRIVATE_TOKEN))
	mac.Write([]byte(qs.Encode()))
	qs.Set("signature", hex.EncodeToString(mac.Sum(nil)))

	var res []Asset
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetQueryString(qs.Encode()).
		SetResult(&res).
		SetHeader("X-MEXC-APIKEY", cfg.Mexc.ACCESS_TOKEN).
		Get(cfg.Mexc.API.CONFIG_GETALL)
	return res, err
}

// Futures. url - is a string from config.Mexc.API.CONTRACT_DETAIL
func (api *MexcAPI) FetchContractInformation(ctx context.Context, url string) ([]ContractDetail, error) {
	if len(url) == 0 {
		return nil, errors.New("provided url in 'FetchContractInformation' is empty!")
	}

	var res Contracts
	_, err := resty.New().
		R().
		SetContext(ctx).
		SetResult(&res).
		Get(url)
	return res.Data, err
}
