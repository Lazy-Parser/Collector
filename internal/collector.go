package collector

import (
	"fmt"
	"log"
	"os"
	"context"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func Run(ctx context.Context) error {
	mexcWS, dexWS, err := loadDotenv();
	if err != nil {
		return fmt.Errorf("Error retreiving dotenv vars: %w", err);
	}
	log.Printf("Dotenv vars success ✅");

	nc, err := connect(ctx, mexcWS, dexWS);
	if err != nil {
		fmt.Errorf("error connection: %w", err);
	}
	defer nc.Close();

	return nil;
}

func loadDotenv() (string, string, error) {
	if err := godotenv.Load(); err != nil {
		return "", "", err;
	}

	mexcWS := os.Getenv("MEXC_WS");
	dexWS := os.Getenv("DEX_WS");
	return mexcWS, dexWS, nil;
}

func connect(ctx context.Context, mexcWS string, dexWS string) (*nats.Conn, error) {
	// connect nats
	nc, err := nats.Connect(mexcWS);
	if err != nil {
		return nil, fmt.Errorf("connect to NATS Mexc: %w", err);
	}
	log.Println("Connected to Mexc ✅");

	// TODO: connect to dex. Remove this string
	fmt.Println("%s | %s", mexcWS, dexWS);



	return nc, nil;
}