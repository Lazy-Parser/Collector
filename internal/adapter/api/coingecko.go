package api_internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Lazy-Parser/Collector/config"
	"github.com/Lazy-Parser/Collector/market"
	httpclient "github.com/Lazy-Parser/Collector/internal/adapter/http"
)

type CoingeckoApi struct {
	cfg    *config.Config
	client *httpclient.Client
}

func NewCoingeckoApi(cfg *config.Config) *CoingeckoApi {
	return &CoingeckoApi{
		cfg:    cfg,
		client: httpclient.New(20, 1, time.Duration(time.Second*5)),
	}
}

func (api *CoingeckoApi) GetTokenData(ctx context.Context, network string, addresses []string) (market.CGResponse, error) {
	urlStr := api.cfg.Coingecko.API.TOKENS_INFO                                           // for example "https://coingecko.com/get-token/{network}/{addresses}"
	urlStr = strings.ReplaceAll(urlStr, "{network}", network)                             // set network
	urlStr = strings.ReplaceAll(urlStr, "{addresses}", joinAddressesWithComma(addresses)) // set addresses

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return market.CGResponse{}, fmt.Errorf("failed to create coingecko request: %v", err)
	}
	req.Header.Add("x-cg-demo-api-key", api.cfg.Coingecko.API.KEY)

	resp, err := api.client.Do(req)
	if err != nil {
		return market.CGResponse{}, fmt.Errorf("coingecko request: %v", err)
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return market.CGResponse{}, fmt.Errorf("coingecko read body while error (%d): %v", resp.StatusCode, err)
		}
		return market.CGResponse{}, fmt.Errorf("coingecko request status code not OK: %d - %s", resp.StatusCode, string(raw))
	}

	var res market.CGResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return market.CGResponse{}, fmt.Errorf("coingecko decode body: %w", err)
	}

	return res, nil

}

func joinAddressesWithComma(addresses []string) string {
	var builder strings.Builder
	for i, address := range addresses {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(address)
	}
	return builder.String()
}
