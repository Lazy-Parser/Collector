package wsclient

import "fmt"

func wrapInQuotes(str string) string {
	return fmt.Sprintf("\"%s\"", str)
}
