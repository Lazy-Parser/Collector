package generator

import (
	config "Cleopatra/config/service"
	"Cleopatra/internal/adapter/out/webapi/chains"
	market "Cleopatra/internal/market/entity"
	"Cleopatra/internal/port"
	"context"

	"golang.org/x/sync/errgroup"
)

var (
	minVolume     = 70_000.0
	maxGoroutines = 4
)

type Generator struct {
	logger             port.Logger
	database           Database
	exchange           Exchange // TODO: make a list of exchanges in future fot supporting multiple exchanges
	dexscreenerService DexscreenerRepo
	coingecko          CoingeckoRepo
	chainsService      *chains.Chains
	validated          []market.Pair // TODO: remove
}

func NewGenerator(
	logger port.Logger,
	db Database,
	exchange Exchange,
	dexscreenerService DexscreenerRepo,
	coingeckoService CoingeckoRepo,
	chainsService *chains.Chains,
) *Generator {
	return &Generator{
		logger:             logger,
		exchange:           exchange,
		database:           db,
		dexscreenerService: dexscreenerService,
		coingecko:          coingeckoService,
		chainsService:      chainsService,
		validated:          make([]market.Pair, 0, 2000),
	}
}

// FetchFuturesMexc method fetср futures from mexc. Return all tokens
// FetchAndValidate (rename to FetchPairsByTokens), it should stream fetched and validated data to out channel. Progress maximum can be maid from FetchFuturesMexc res len
// FetchDecimals also streams out it results

func (generator *Generator) FetchFuturesMexc(ctx context.Context, cfg *config.Config) ([]market.Token, error) {
	// load filtered mexc futures tokens
	return generator.exchange.GetFutures(ctx, cfg, generator.chainsService)
}

// Function GenerateMexc is a function, that fetchs tokens from Mexc api, validates data. 'progress' is a channel that shows how many pairs "validated" for progrss visualization in ui
func (generator *Generator) FetchPairs(
	ctx context.Context,
	cfg *config.Config,
	pairCh chan market.Pair,
	mexcTokens []market.Token,
) error {
	defer close(pairCh)

	grp, childCtx := errgroup.WithContext(ctx)
	grp.SetLimit(4) // becasue api limitation is 300req / min (but we will take 290), 290 req / 60 sec = +-4.8 per sec. TODO: do not hardcode this value, but calculate it.

	// next we need to load decimals for each token and load pair info for each token
	// pair info from dexscreener (limit 300 req/min)
	// token decimal from coingecko (30 addresses per request, 100req/min)
	// 1. Pair info (dexscreener)
	for _, t := range mexcTokens {
		token := t

		grp.Go(func() error {
			dexResp, err := generator.dexscreenerService.Fetch(ctx, cfg, token, generator.chainsService)
			if err != nil {
				generator.logger.Error("Dexscreener try to fetch pair from contract (%s), that caused error: %v", token.Address, err)
				pairCh <- market.Pair{}
				return nil
			}

			// if token is bad, dexscreener will return empty arr
			if len(*dexResp) == 0 {
				pairCh <- market.Pair{}
				return nil
			}

			bestPair, ok := market.SelectBest(dexResp, minVolume)
			if ok {
				// all validation passed
				select {
				case pairCh <- bestPair:
				case <-childCtx.Done():
					return childCtx.Err()
				}
			}

			pairCh <- market.Pair{}
			return nil
		})
	}

	return grp.Wait()
}

// Function CreateChunks groups all provided tokens to []Chunk by network, in one chunk can be maximum 30 tokens. Used to fetch decimals from CoinGecko API
func (generator *Generator) CreateChunks(tokens []market.Token) []Chunk {
	chunks := make([]Chunk, 0, 64)
	for _, token := range tokens {
		for i := range chunks {
			// try to find existing chunk
			if chunks[i].network == token.Network && len(chunks[i].tokens) < ChunkMaxSize {
				chunks[i].Push(token)
				continue
			}
		}

		// none found, create one
		arr := make([]market.Token, 0, ChunkMaxSize)
		arr = append(arr, token)
		chunks = append(chunks, Chunk{
			network: token.Network,
			tokens:  arr,
		})
	}

	return chunks
}

// 2. Decimals (coingecko)
func (generator *Generator) FetchDecimals(ctx context.Context, cfg *config.Config, chunk Chunk) ([]market.Token, error) {
	return generator.coingecko.FetchChunk(ctx, cfg, chunk.network, chunk.tokens)
}

// Function Save saves all generated data to Database
func (generator *Generator) SaveTokens(tokens []market.Token) {

}
