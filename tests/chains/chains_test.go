package chains_test

import (
	"Cleopatra/internal/adapter/out/webapi/chains"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	chainsService *chains.Chains
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// SetupSuite runs only once before all tests.
func (s *IntegrationTestSuite) SetupSuite() {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "store", "chains.json")
	
	chainsService, err := chains.NewChains(path)
	s.Require().NoError(err, "err while creation new ChainsService")

	s.chainsService = chainsService
}

func (s *IntegrationTestSuite) Test_Ethereum() {
	// Ethereum
	res, ok := s.chainsService.Select("Ethereum").ToMexc()
	s.Require().Equal("ETH", res, "Ethereum -> ETH (ToMexc) failed")
	s.Require().True(ok, "Ethereum -> ETH (ToMexc) status failed")

	res, ok = s.chainsService.Select("eth").ToBase()
	s.Require().Equal("Ethereum", res, "eth -> ethereum (ToBase) failed")
	s.Require().True(ok, "eth -> ethereum (ToBase) status failed")

	res, ok = s.chainsService.Select("ETH").ToBase()
	s.Require().Equal("Ethereum", res, "eth -> Ethereum (ToBase) failed")
	s.Require().True(ok, "eth -> Ethereum (ToBase) status failed")

	// Should be error
	res, ok = s.chainsService.Select("Eth").ToBase()
	s.Require().Equal("", res, "Eth -> '' (ToBase) failed")
	s.Require().False(ok, "Eth -> '' (ToBase) status failed")
}

func (s *IntegrationTestSuite) Test_Solana() {
	// Solana
	res, ok := s.chainsService.Select("Solana").ToBase()
	s.Require().Equal("Solana", res, "Solana -> Solana (ToBase) failed")
	s.Require().True(ok, "Solana -> Solana (ToBase) status failed")

	res, ok = s.chainsService.Select("SOL").ToBase()
	s.Require().Equal("Solana", res, "SOL -> Solana (ToBase) failed")
	s.Require().True(ok, "SOL -> Solana (ToBase) status failed")

	res, ok = s.chainsService.Select("SOL").ToCoingecko()
	s.Require().Equal("solana", res, "SOL -> solana (ToCoingecko) failed")
	s.Require().True(ok, "SOL -> solana (ToCoingeckp) status failed")
}

func (s *IntegrationTestSuite) Test_Whitelist() {
	// in any style
	whitelist := []string{"SOL", "eth", "BSC", "TRON"} // TRON in not whitelisted

	for i, elem := range whitelist {
		ok := s.chainsService.IsWhitelist(elem)

		// if last. Becasue TRON is the last elem, that should give false
		if i == len(whitelist)-1 {
			s.Require().False(ok, fmt.Sprintf("Network %s should not be whitelisted, but service gived True", elem))
		} else {
			s.Require().True(ok, fmt.Sprintf("Network %s should be whitelisted, but service gived False", elem))
		}
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	log.Println("END!")
}
