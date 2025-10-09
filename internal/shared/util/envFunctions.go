package util

import(
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
		port, err := strconv.Atoi(value)
		if err == nil {
			return port
		}
	}
	return defaultValue
}