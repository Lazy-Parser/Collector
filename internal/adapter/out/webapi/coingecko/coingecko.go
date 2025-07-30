package coingecko

import (
	config "Cleopatra/config/service"
	market "Cleopatra/internal/market/entity"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

var (
	requestGroupSize = 30
)

type Coingecko struct {
	limitter *rate.Limiter // 90 requests per minute
	client   *http.Client
	mu       sync.Mutex
}

func NewCoingecko() *Coingecko {
	return &Coingecko{
		limitter: rate.NewLimiter(rate.Limit(90.0/60.0), 1),
		client:   &http.Client{},
	}
}

// TODO: make requests and return
func (cg *Coingecko) FetchChunk(ctx context.Context, cfg *config.Config, network string, tokens []market.Token) ([]market.Token, error) {
	urlStr := cfg.Coingecko.API.TOKENS_INFO                                            // for example "https://coingecko.com/get-token/{network}/{addresses}"
	urlStr = strings.ReplaceAll(urlStr, "{network}", network)                          // set network
	urlStr = strings.ReplaceAll(urlStr, "{addresses}", joinAddressesWithComma(tokens)) // set addresses

	if err := cg.limitter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create coingecko request: %v", err)
	}
	req.Header.Add("x-cg-demo-api-key", cfg.Coingecko.API.KEY)

	resp, err := cg.makeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create coingecko request: %v", err)
	}

	tokensList := make([]market.Token, len(tokens))
	for _, tokenFromApi := range resp.Data {
		// find corresponding token
		t, ok := market.FindTokenByAddress(&tokens, tokenFromApi.Attributes.Address)
		if ok {
			// set decimal from api resp and push to channel
			t.Decimal = uint8(tokenFromApi.Attributes.Decimals)
			tokensList = append(tokensList, t)
		}
	}

	return tokensList, nil
}

func (cg *Coingecko) makeRequest(req *http.Request) (Response, error) {
	resp, err := cg.client.Do(req)
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
