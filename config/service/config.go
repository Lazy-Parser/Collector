// Package config extract .env vars and sets to custom types
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type (
	Config struct {
		ENVIRONMENT string `env:"ENVIRONMENT" envDefault:"nats://localhost:4222"`
		Nats        NATS

		Mexc        MEXC
		Dexscreener DEXSCREENER
		Coingecko   COINGECKO
	}

	MEXC struct {
		SPOT_WS       string `env:"MEXC_SPOT_WS,required"`
		FUTURES_WS    string `env:"MEXC_FUTURES_WS,required"`
		ACCESS_TOKEN  string `env:"MEXC_ACCESS_TOKEN,required"`
		PRIVATE_TOKEN string `env:"MEXC_PRIVATE_TOKEN,required"`
		API           MEXC_API
	}
	MEXC_API struct {
		CONFIG_GETALL    string `env:"MEXC_API_CONFIG_GETALL,required"`
		CONTRACTS_DETAIL string `env:"MEXC_API_CONTRACTS_DETAIL,required"`
	}

	DEXSCREENER struct {
		API DEXSCREENER_API
	}
	DEXSCREENER_API struct {
		TOKEN_PAIRS string `env:"DEXSCREENER_GET_TOKEN_PAIRS,required"`
	}

	COINGECKO struct {
		API COINGECKO_API
	}
	COINGECKO_API struct {
		TOKENS_INFO string `env:"COINGECKO_API_TOKENS_INFO,required"`
		KEY         string `env:"COINGECKO_API_KEY,required"`
	}

	NATS struct {
		URL string `env:"NATS_URL" envDefault:"nats://localhost:4222"`
	}
)

func NewConfig(filename string) (*Config, error) {
	cfg := &Config{}

	// Remove in prod
	err := godotenv.Load(filename)
	if err != nil {
		return nil, fmt.Errorf("loading .env error: %v", err)
	}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
	}

	return cfg, nil
}
