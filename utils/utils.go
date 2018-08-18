package utils

import (
	"os"
	"strconv"
)

// GetEnv retrieves value of environment variable with given key.
// If the key does not exists it returns the fallback string.
func GetEnv(key string, fallback string) string {
	value, success := os.LookupEnv(key)
	if success {
		return value
	} else {
		return fallback
	}
}

// GetEnvAsInt converts return value of GetEnv as int.
func GetEnvAsInt(key string, fallback string) int {
	value, _ := strconv.Atoi(GetEnv(key, fallback))
	return value
}
