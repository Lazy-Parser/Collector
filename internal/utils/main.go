package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	d "github.com/Lazy-Parser/Collector/internal/core"
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

func LoadWhitelistFile() ([]d.Whitelist, error) {
	workDir, err := GetWorkDirPath()
	if err != nil {
		return []d.Whitelist{}, err
	}

	path := filepath.Join(workDir, "config", "network_pool_whitelist.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return []d.Whitelist{}, fmt.Errorf("loading 'config/network_pool_whitelist.json' file: %v", err)
	}

	var res []d.Whitelist
	err = json.Unmarshal(data, &res)
	if err != nil {
		return []d.Whitelist{}, fmt.Errorf("unmarshal data from 'config/network_pool_whitelist.json' file: %v", err)
	}

	return res, nil
}
