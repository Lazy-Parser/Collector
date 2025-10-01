package chains_test

import (
	"fmt"
	"path/filepath"

	"github.com/Lazy-Parser/Collector/chains"
)

// ExampleNewChains demonstrates how to create a new Chains instance.
func ExampleNewChains() {
	path := filepath.Join("path", "to", "the", "chains.json")
	chains, err := chains.NewChains(path)
	if err != nil {
		// do not panic in real app
		panic(err)
	}

	// do smth
	chains.Select("eth")
}

func CreateChains() *chains.Chains {
	path := filepath.Join("path", "to", "the", "chains.json")
	chains, err := chains.NewChains(path)
	if err != nil {
		// do not panic in real app
		panic(err)
	}

	return chains
}
func ExampleChains_Select() {
	// create instance
	// Chains - is a service that just store different netowork names. It used to change one network name to another one.
	chains := CreateChains()

	// First, we need to select what network we whant to change.
	// This method accept all network names possible (Default (which i called 'based'), Mexc, Dexscreener, Coingecko)
	// This method returns 'Selector'.
	//  - It has a 'status' bool. It indicates if the network was successfully selected / found.
	//  - And 'meta'. Its a struct that contains all possible network name variations for selected network.
	chains.Select("SOL")

	// Let's try to select 'eth' and change it to the Base Name ('Ethereum')
	baseName, ok := chains.Select("eth").ToBase()

	// check if 'eth' found
	if !ok {
		panic("network not found")
	}

	fmt.Println(baseName) // will print "Ethereum"

	// ----- Or let's try to select 'SOL' and change it to the Coingecko type ('solana')
	baseName, ok = chains.Select("sol").ToCoingecko()

	// check if 'sol' found
	if !ok {
		panic("network not found")
	}

	fmt.Println(baseName) // will print "solana"
}

func ExampleChains_IsWhitelist() {
	// Another one useful chains's abilitiy is to check if some token or pair is 'allowed'. 
	// Thus we can filter tokens or pairs by needed networks and not use some unpopular ones.
	
	chains := CreateChains()
	
	// This method works like Selector, but return only status. So we can pass any types of network names
	
	chains.IsWhitelist("eth") // -> True
	chains.IsWhitelist("Ethereum") // -> True
	chains.IsWhitelist("ETH") // -> True
	
	chains.IsWhitelist("Solana") // -> True
	chains.IsWhitelist("SOL") // -> True
	chains.IsWhitelist("Solanaaaa") // -> False. Wrong spelling
	
	chains.IsWhitelist("TRON") // -> False. There are not TRON network in Chains
	chains.IsWhitelist("tron") // -> False. There are not tron network in Chains
}