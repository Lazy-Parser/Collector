package api_internal

import (
	"context"
	"time"

	"github.com/Lazy-Parser/Collector/config"
	httpclient "github.com/Lazy-Parser/Collector/internal/adapter/http"
	"github.com/Lazy-Parser/Collector/market"
)

type DexscreenerApi struct {
	cfg    *config.Config
	client *httpclient.Client
}

func NewDexscreenerApi(cfg *config.Config) *DexscreenerApi {
	client := httpclient.New(290, 1, time.Duration(time.Second*5))

	return &DexscreenerApi{
		cfg:    cfg,
		client: client,
	}
}

func (api *DexscreenerApi) GetTokenPairs(ctx context.Context, network string, address string) (*market.DexscreenerResponse, error) {
	return nil, nil
}
