// This package converts provided network name from any api (Mexc, Dexscreener, Coingecko) to the global name, that will be used in the whole app ('SOL' -> 'Solana', 'eth' -> 'Ethereum', 'SOLANA' -> 'Solana').
//
// Also it checks if provided network in whitelist.
package chains

// TODO: Add TRON network maybe,  Polygon, TON, but on ton too litle volume

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type Chains struct {
	data []ChainMeta
}

type Selector struct {
	meta   *ChainMeta
	status bool
}

// filepath - is a path to the 'chains.json', default in project/config/configs/ folder
func NewChains(filepath string) (*Chains, error) {
	raw, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file 'chains.json' for Chains service: %v", err)
	}

	var data []ChainMeta
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal []bytes from file 'chains.json' for Chains service: %v", err)
	}

	return &Chains{data: data}, nil
}

// This function returns True if provided network name in any style (Mexc (SOL), global (Solana), Coingecko (solana)) contains in Chains list.
//
// Returns False otherwise
func (ch *Chains) IsWhitelist(network string) bool {
	return ch.Select(network).status
}

// This function returns Chain and status:
//   - Status = True:  Convertion succsess.
//   - Status = False: Provided network not found in the list of chains.
func (ch *Chains) Select(network string) *Selector {
	for _, chain := range ch.data {
		// Check if user provided already global name
		if chain.Id == network {
			return &Selector{
				meta:   &chain,
				status: true,
			}
		}

		v := reflect.ValueOf(chain.Providers)
		for i := range v.NumField() {
			if v.Field(i).String() == network {
				return &Selector{
					meta:   &chain,
					status: true,
				}
			}
		}
	}

	return &Selector{
		meta:   nil,
		status: false,
	}
}

func (sel *Selector) ToBase() (string, bool) {
	if sel.status {
		return sel.meta.Id, true
	} else {
		return "", false
	}
}

func (sel *Selector) ToMexc() (string, bool) {
	if sel.status {
		return sel.meta.Providers.Mexc, true
	} else {
		return "", false
	}
}

func (sel *Selector) ToDexscreener() (string, bool) {
	if sel.status {
		return sel.meta.Providers.Dexscreener, true
	} else {
		return "", false
	}
}

func (sel *Selector) ToCoingecko() (string, bool) {
	if sel.status {
		return sel.meta.Providers.Coingecko, true
	} else {
		return "", false
	}
}
