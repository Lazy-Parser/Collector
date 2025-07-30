package dexscreener

import (
	config "Cleopatra/config/service"
	"Cleopatra/internal/adapter/out/webapi/chains"
	market "Cleopatra/internal/market/entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

type Dexscreener struct {
	limiter *rate.Limiter
	client  *http.Client
}

func NewDexscreener() *Dexscreener {
	return &Dexscreener{
		limiter: rate.NewLimiter(rate.Limit(290.0/60.0), 1),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// Function Fetch gets a list of pairs by provided token and return the best one (with the biggest volume). Pairs with volume < 70k$ do not used
func (dexscreener *Dexscreener) Fetch(ctx context.Context, cfg *config.Config, token market.Token, chainsService *chains.Chains) (*[]market.PairCandidat, error) {
	if err := dexscreener.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	network, ok := chainsService.Select(token.Network).ToDexscreener()
	if !ok {
		return nil, errors.New(fmt.Sprintf("failed to change global network name to dexscreener one. Global: %s", token.Network))
	}

	urlStr := cfg.Dexscreener.API.TOKEN_PAIRS
	urlStr = strings.ReplaceAll(urlStr, "{chainId}", network)
	urlStr = strings.ReplaceAll(urlStr, "{tokenAddress}", token.Address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("dexscreener create request: %v", err)
	}

	resp, err := dexscreener.client.Do(req)
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
		return nil, fmt.Errorf("dexscreener request status code not OK: %d - %s", resp.StatusCode, string(raw))
	}

	var resFromApi DexScreenerResponse
	if err := json.NewDecoder(resp.Body).Decode(&resFromApi); err != nil {
		return nil, fmt.Errorf("dexscreener decode body: %w", err)
	}

	// mapping. PairDS -> market.PairCandidat
	// if len(resFromApi) == 0, that means that provided token address is not traded, we need to skip it.
	// so this loop will "skip" empty arr and will return empty arr too. Its very important, because we need to check on empty res in usecase
	var res []market.PairCandidat
	for _, pairds := range resFromApi {
		newPair, err := normalizePair(pairds, chainsService)
		if err != nil {
			fmt.Printf("failed to normalize pair from dexscreener api response: %v", err)
			continue
		}
		
		res = append(res, newPair)
	}

	return &res, nil
}
