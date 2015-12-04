package utils

import (
	"os"
)

// GetOpt returns the specified environment variable's value or a default value if that
// environment variable's value is the empty string.
func GetOpt(name string, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		value = dfault
	}
	return value
}
