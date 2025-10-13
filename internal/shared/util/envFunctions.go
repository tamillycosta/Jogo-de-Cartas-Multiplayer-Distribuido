// Adicione ao seu arquivo utils

package util

import (
	"os"
	"strconv"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetPortFromEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if port, err := strconv.Atoi(value); err == nil {
			return port
		}
	}
	return defaultValue
}

func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}