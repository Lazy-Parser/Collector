package generator

import (
	worker_coingecko "github.com/Lazy-Parser/Collector/internal/domain/api/coingecko"
	worker_dexscreener "github.com/Lazy-Parser/Collector/internal/domain/api/dexscreener"
	worker_mexc "github.com/Lazy-Parser/Collector/internal/domain/api/mexc"
	market "github.com/Lazy-Parser/Collector/internal/domain/market"
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type GeneratorService struct {
	mexcWorker        *worker_mexc.MexcWorker
	dexscreenerWorker *worker_dexscreener.DexscreenerWorker
	coingeckoWorker   *worker_coingecko.CoingeckoWorker
}

func NewGeneratorService(
	mexcWorker *worker_mexc.MexcWorker,
	dexscreenerWorker *worker_dexscreener.DexscreenerWorker,
	coingeckoWorker *worker_coingecko.CoingeckoWorker,
) *GeneratorService {
	return &GeneratorService{
		mexcWorker:        mexcWorker,
		dexscreenerWorker: dexscreenerWorker,
		coingeckoWorker:   coingeckoWorker,
	}
}

// Get futures from mexc
func (service *GeneratorService) GetFutures(ctx context.Context) ([]market.Token, error) {
	tokens, err := service.mexcWorker.GetAllTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mexc futures in generator: %v", err)
	}

	futures, err := service.mexcWorker.GetAllFutures(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mexc futures in generator: %v", err)
	}

	// tokens are already validated in mexcWorker, so we just need to combine all info both from tokens and futures
	var res []market.Token
	for _, token := range tokens {
		contract, ok := service.mexcWorker.FindContractBySymbol(&futures, token.Coin)
		if !ok {
			// This token does not traides on futures
			continue
		}

		res = append(res, market.Token{
			Name:        token.Coin,
			Decimal:     0,                             // unknown, will find it later
			Address:     token.NetworkList[0].Contract, // TODO: Network list can contains > 1 elems
			Network:     token.NetworkList[0].Network,
			WithdrawFee: token.NetworkList[0].WithdrawFee,
			Image_url:   contract.ImageUrl,
			CreateTime:  contract.CreateTime,
		})
	}

	return res, nil
}

// Get pairs info from dexscreener by provided tokens
func (service *GeneratorService) GetPairs(ctx context.Context, tokens []market.Token) ([]market.Pair, error) {
	out := make(chan market.Pair, len(tokens))
	done := make(chan struct{})
	pairList := make([]market.Pair, 0, len(tokens))

	go func() {
		for pair := range out {
			pairList = append(pairList, pair)
		}
		close(done)
	}()

	grp, childCtx := errgroup.WithContext(ctx)
	grp.SetLimit(4) // becasue api limitation is 300req / min (but we will take 290), 290 req / 60 sec = +-4.8 per sec. TODO: do not hardcode this value, but calculate it.

	emptyPair := market.Pair{}
	for _, t := range tokens {
		token := t

		grp.Go(func() error {
			pair, err := service.dexscreenerWorker.FetchPairByToken(childCtx, token)
			if err != nil {
				return err
			}

			if pair == emptyPair {
				return nil
			}

			// add custom token to pair. They are the same tokens, but the custom token has more info.
			pair.BaseToken = token
			select {
			case out <- pair:
			case <-childCtx.Done():
				return childCtx.Err()
			}

			return nil
		})
	}

	err := grp.Wait()
	close(out)
	<-done
	return pairList, err
}

func (service *GeneratorService) GetDecimals(ctx context.Context, tokens []market.Token) ([]market.Token, error) {
	// create chunks
	chunks := service.coingeckoWorker.CreateChunks(tokens)

	var tokenList []market.Token
	type upd struct {
		address string
		decimal uint8
	}
	updates := make(chan upd, len(tokens))
	done := make(chan struct{}, 1)

	grp, childCtx := errgroup.WithContext(ctx)
	grp.SetLimit(1)

	go func() {
		for update := range updates {
			if update.address == "" {
				continue
			}

			t, ok := market.FindTokenByAddress(&tokens, update.address)
			if ok {
				t.Decimal = update.decimal
				tokenList = append(tokenList, t)
			}
		}
		close(done)
	}()

	// fetch decimals
	for _, ch := range chunks {
		chunk := ch
		grp.Go(func() error {
			resp, err := service.coingeckoWorker.FetchDecimals(childCtx, chunk)
			if err != nil {
				return err
			}

			for _, elem := range resp.Data {
				u := upd{address: elem.Attributes.Address, decimal: uint8(elem.Attributes.Decimals)}
				select {
				case updates <- u:
				case <-childCtx.Done():
					return childCtx.Err()
				}
			}

			return nil
		})
	}

	err := grp.Wait()
	close(updates)
	<-done

	return tokenList, err
}
