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
	urlStr := api.cfg.Dexscreener.API.TOKEN_PAIRS
	urlStr = strings.ReplaceAll(urlStr, "{chainId}", network)
	urlStr = strings.ReplaceAll(urlStr, "{tokenAddress}", address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("dexscreener create request: %v", err)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dexscreener fetch data: %v", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("dexscreener read body while error (%d): %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("dexscreener request status code not OK: %d \nRequest: %s  \nError:%s", resp.StatusCode, urlStr, string(raw))
	}

	var resFromApi market.DexscreenerResponse
	if err := json.NewDecoder(resp.Body).Decode(&resFromApi); err != nil {
		return nil, fmt.Errorf("dexscreener decode body: %w", err)
	}

	return &resFromApi, nil
}
