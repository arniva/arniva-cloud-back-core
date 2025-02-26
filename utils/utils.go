package utils

import "os"

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Ptr[T any](v T) *T {
	return &v
}
