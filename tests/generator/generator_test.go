package generator_test

import (
	config "github.com/Lazy-Parser/Collector/config/service"
	api_coingecko "github.com/Lazy-Parser/Collector/internal/api/coingecko"
	api_dexscreener "github.com/Lazy-Parser/Collector/internal/api/dexscreener"
	api_mexc "github.com/Lazy-Parser/Collector/internal/api/mexc"
	"github.com/Lazy-Parser/Collector/internal/application/generator"
	"github.com/Lazy-Parser/Collector/internal/common/chains"
	logger "github.com/Lazy-Parser/Collector/internal/common/zerolog"
	worker_coingecko "github.com/Lazy-Parser/Collector/internal/domain/api/coingecko"
	worker_dexscreener "github.com/Lazy-Parser/Collector/internal/domain/api/dexscreener"
	worker_mexc "github.com/Lazy-Parser/Collector/internal/domain/api/mexc"
	"github.com/Lazy-Parser/Collector/internal/domain/market"
	server_test "github.com/Lazy-Parser/Collector/tests/server"
	"context"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	config           *config.Config
	server           *httptest.Server
	generatorService *generator.GeneratorService
	futures          []market.Token
	cg               *worker_coingecko.CoingeckoWorker
	tokenWorker      *market.TokenWorker
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// SetupSuite runs only once before all tests.
func (s *IntegrationTestSuite) SetupSuite() {
	wd, _ := os.Getwd()

	// server
	server := server_test.SetupServer()
	s.server = server
	log.Printf("Fake server: %s", server.URL)

	// set env values endpoints
	dotenvPath := filepath.Join(wd, "..", ".env.test")
	err := updateDotenv(dotenvPath, "MEXC_API_CONFIG_GETALL", server.URL+"/mexc/config/getall")
	err = updateDotenv(dotenvPath, "MEXC_API_CONTRACTS_DETAIL", server.URL+"/mexc/contract/detail")
	s.Require().NoError(err, "Failed to update test dotenv")

	cfg, err := config.NewConfig(dotenvPath)
	s.Require().NoError(err, "Load config")
	s.config = cfg

	// database
	// dbPath := filepath.Join(wd, "..", "database", "collector_fake.db")
	// db, err := database.NewConnection(dbPath)
	// s.Require().NoError(err, "Creating fake database")

	// chains
	chainsPaths := filepath.Join(wd, "..", "..", "config", "configs", "chains.json")
	chainsService, err := chains.NewChains(chainsPaths)
	s.Require().NoError(err, "failed to create new instanse of Chains Service")

	// exchange
	mexcApi := api_mexc.NewMexcAPI()
	exchange := worker_mexc.NewWorker(mexcApi, cfg, chainsService)

	// logger
	logger.New(os.Stdout)

	// dexscreener
	dsApi := api_dexscreener.NewDexscreenerAPI(cfg)
	dex := worker_dexscreener.NewWorker(dsApi, chainsService)

	// coingecko
	cgApi := api_coingecko.NewCoingeckoAPI(cfg)
	cg := worker_coingecko.NewWorker(cfg, cgApi)
	s.cg = cg

	generatorService := generator.NewGeneratorService(exchange, dex, cg)
	s.generatorService = generatorService

	tokenWorker := market.NewTokenWorker()
	s.tokenWorker = tokenWorker
	
	log.Println("Generator created!")
}

func (s *IntegrationTestSuite) Test_01_Mexc() {
	shouldContain := []string{"CELR", "SCR", "TRUMP", "FARTCOIN"}

	ctx := context.Background()
	tokens, err := s.generatorService.GetFutures(ctx)
	s.Require().NoError(err, "Fetching and validating tokens from mexc")

	for _, token := range tokens {
		s.Require().True(slices.Contains(shouldContain, token.Name), "MexcService response doesn't contains expected tokens")
	}

	s.tokenWorker.PushMany(tokens)
}

func (s *IntegrationTestSuite) Test_02_Dexscreener() {
	ctx := context.Background()
	shouldContain := []string{"trump", "fartcoin", "celr"}

	// here we got Pair info + base token info. Quote token has not all info we need. So we need to fetch it separately
	pairs, err := s.generatorService.GetPairs(ctx, s.tokenWorker.GetAllTokens())
	s.Require().NoError(err, "Fetching and validating pairs from dexscreener")

	
	
	log.Println("--------- PAIRS ---------")
	var payload []market.Token
	for _, pair := range pairs {
		log.Printf("Pair: %+v", pair)
		payload = append(payload, pair.QuoteToken)
		payload = append(payload, pair.BaseToken)
	}
	log.Println("--------- PAIRS ---------")
	tokens, err := s.generatorService.GetDecimals(ctx, payload)
	s.Require().NoError(err, "Fetching and validating quote tokens from dexscreener")
	for i := range pairs {
		for _, token := range tokens {
			if token.Address == pairs[i].QuoteToken.Address {
				pairs[i].QuoteToken = token
				break
			} else if token.Address == pairs[i].BaseToken.Address {
				pairs[i].BaseToken = token
				break
			}
		}
	}

	log.Println("----------------------")
	log.Println("RESULT")
	log.Println("----------------------")

	for _, pair := range pairs {
		if pair.Address == "" {
			continue
		}
		log.Printf("Pair: %+v", pair)
		s.Require().True(slices.Contains(shouldContain, strings.ToLower(pair.BaseToken.Name)), "DexscreenerService response doesn't contains expected pairs")
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.server.Close()
}

func updateDotenv(path, key, value string) error {
	// Read .env into a map
	env, err := godotenv.Read(path)
	if err != nil {
		return err
	}

	// Update or insert
	env[key] = value

	// Write back to the same file (comments/order won’t be preserved)
	return godotenv.Write(env, path)
}
