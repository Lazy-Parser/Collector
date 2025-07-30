package generator_test

import (
	config "Cleopatra/config/service"
	logger "Cleopatra/internal/adapter/out/log/zerolog"
	database "Cleopatra/internal/adapter/out/persistent/sqlite"
	"Cleopatra/internal/adapter/out/webapi/chains"
	"Cleopatra/internal/adapter/out/webapi/coingecko"
	"Cleopatra/internal/adapter/out/webapi/dexscreener"
	"Cleopatra/internal/adapter/out/webapi/mexc"
	generator "Cleopatra/internal/generator/usecase"
	market "Cleopatra/internal/market/entity"
	server_test "Cleopatra/tests/server"
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
	generatorService *generator.Generator
	futures          []market.Token
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
	dbPath := filepath.Join(wd, "..", "database", "collector_fake.db")
	db, err := database.NewConnection(dbPath)
	s.Require().NoError(err, "Creating fake database")

	// chains
	chainsPaths := filepath.Join(wd, "..", "..", "config", "configs", "chains.json")
	chainsService, err := chains.NewChains(chainsPaths)
	s.Require().NoError(err, "failed to create new instanse of Chains Service")

	// exchange
	exchange := mexc.NewMexc(chainsService)

	// logger
	l := logger.New(os.Stdout)

	// dexscreener
	dex := dexscreener.NewDexscreener()

	// coingecko
	cg := coingecko.NewCoingecko()

	generatorService := generator.NewGenerator(l, db, exchange, dex, cg, chainsService)
	s.generatorService = generatorService

	log.Println("Generator created!")
}

func (s *IntegrationTestSuite) Test_01_Mexc() {
	shouldContain := []string{"CELR", "SCR", "TRUMP", "FARTCOIN"}

	ctx := context.Background()
	tokens, err := s.generatorService.FetchFuturesMexc(ctx, s.config)
	s.Require().NoError(err, "Fetching and validating tokens from mexc")

	for _, token := range tokens {
		log.Printf("Token: %+v", token)
		s.Require().True(slices.Contains(shouldContain, token.Name), "MexcService response doesn't contains expected tokens")
	}

	s.futures = tokens
}

func (s *IntegrationTestSuite) Test_02_Dexscreener() {
	shouldContain := []string{"trump", "fartcoin "}
	emptyPair := market.Pair{}

	ctx := context.Background()
	pairCh := make(chan market.Pair, 2048)

	s.generatorService.FetchPairs(ctx, s.config, pairCh, s.futures)

	for pair := range pairCh {
		if pair == emptyPair {
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
