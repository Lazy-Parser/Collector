package utils

import (
	"os"
	"fmt"
	"os/exec"
	"strings"
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