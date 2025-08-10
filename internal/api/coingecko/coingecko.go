package api_coingecko

import (
	config "github.com/Lazy-Parser/Collector/config/service"
	logger "github.com/Lazy-Parser/Collector/internal/common/zerolog"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/time/rate"
)

type CoingeckoAPI struct {
	cfg     *config.Config
	limiter *rate.Limiter
	client  *http.Client
}

func NewCoingeckoAPI(cfg *config.Config) *CoingeckoAPI {
	return &CoingeckoAPI{
		cfg:     cfg,
		limiter: rate.NewLimiter(rate.Limit(20.0/60.0), 1), // TODO: do not hardcode
		client:  &http.Client{},
	}
}

func (api *CoingeckoAPI) GetTokenData(ctx context.Context, network string, addresses []string) (Response, error) {
	if err := api.limiter.Wait(ctx); err != nil {
		return Response{}, err
	}

	urlStr := api.cfg.Coingecko.API.TOKENS_INFO                                           // for example "https://coingecko.com/get-token/{network}/{addresses}"
	urlStr = strings.ReplaceAll(urlStr, "{network}", network)                             // set network
	urlStr = strings.ReplaceAll(urlStr, "{addresses}", joinAddressesWithComma(addresses)) // set addresses

	logger.Get().Z.Info().Msgf("Coingecko request: %s", urlStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create coingecko request: %v", err)
	}
	req.Header.Add("x-cg-demo-api-key", api.cfg.Coingecko.API.KEY)

	resp, err := api.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("coingecko request: %v", err)
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return Response{}, fmt.Errorf("coingecko read body while error (%d): %v", resp.StatusCode, err)
		}
		return Response{}, fmt.Errorf("coingecko request status code not OK: %d - %s", resp.StatusCode, string(raw))
	}

	var res Response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return Response{}, fmt.Errorf("coingecko decode body: %w", err)
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
