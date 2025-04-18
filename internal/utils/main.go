package utils

import "os"

func GetDotenv(names ...string) ([]string, error) {
	var vars []string
	for idx, name := range names {
		vars[idx] = os.Getenv(name)
	}

	return vars, nil
}
