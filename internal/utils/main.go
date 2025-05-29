package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	core "github.com/Lazy-Parser/Collector/internal/core"
)

// converts symbol string with underscores or without
// a separator to the "BTC/USDT" format by replacing "_" with "/" or
// appending "/" before "USDT" if no separators are found.
func NormalizeSymbol(symbol string) string {
	newString := strings.ReplaceAll(symbol, "_", "/")

	if newString == symbol {
		newString = strings.ReplaceAll(symbol, "USDT", "/USDT")
	}

	return newString
}

func ClearConsole() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func GetWorkDirPath() (string, error) {
	workDirPath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get work directory: %v", err)
	}

	return workDirPath, nil
}

func LoadWhitelistFile() ([]core.Whitelist, error) {
	workDir, err := GetWorkDirPath()
	if err != nil {
		return []core.Whitelist{}, err
	}

	path := filepath.Join(workDir, "config", "network_pool_whitelist.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return []core.Whitelist{}, fmt.Errorf("loading 'config/network_pool_whitelist.json' file: %v", err)
	}

	var res []core.Whitelist
	err = json.Unmarshal(data, &res)
	if err != nil {
		return []core.Whitelist{}, fmt.Errorf("unmarshal data from 'config/network_pool_whitelist.json' file: %v", err)
	}

	return res, nil
}

func TernaryIf[T any](cond bool, argtrue T, argfalse T) T {
	if cond {
		return argtrue
	}

	return argfalse
}

func LoadEnv(envName string) (string, error) {
	godotenv.Load(".env")
	res := os.Getenv(envName)
	if res == "" {
		return "", errors.New("failed to load .env var: " + envName)
	}

	return res, nil
}

func IsErrorReturn(err error, message string, messageArgs ...string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(message, messageArgs)
}

func IsErrorLog(err error, message string, messageArgs ...string) error {
	if err != nil {
		log.Println(message, messageArgs, err)
		return err
	}

	return nil
}
