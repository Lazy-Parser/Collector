package generator

import (
	config "Cleopatra/config/service"
	"Cleopatra/internal/adapter/out/webapi/chains"
	market "Cleopatra/internal/market/entity"
	"context"
)

type Exchange interface {
	GetFutures(ctx context.Context, cfg *config.Config, chainsService *chains.Chains) ([]market.Token, error)
}

type Database interface {
	SaveToken(token market.Token) error
	// SavePort(token market.Pair) error
}

type DexscreenerRepo interface {
	Fetch(ctx context.Context, cfg *config.Config, token market.Token, chainsService *chains.Chains) (*[]market.PairCandidat, error)
}

type CoingeckoRepo interface {
	FetchChunk(ctx context.Context, cfg *config.Config, network string, tokens []market.Token) ([]market.Token, error)
}

// TODO: maybe move this struct to global port folder
const ChunkMaxSize = 30

type Chunk struct {
	network string
	tokens  []market.Token
}

func (chunk Chunk) Push(token market.Token) {
	chunk.tokens = append(chunk.tokens, token)
}
