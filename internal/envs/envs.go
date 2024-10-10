package envs

import (
	"os"
)

// getEnvOrDefault retrieves the environment variable value or returns a default value if it's not set.
func GetEnvOrDefault(envVar, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}
	return value
}
