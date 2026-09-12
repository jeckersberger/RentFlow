package config

import (
	"os"
	"strconv"
)

// FeatureEnabled returns true only when the named environment variable
// contains a valid boolean true value. Missing and invalid values fail closed.
func FeatureEnabled(name string) bool {
	value, err := strconv.ParseBool(os.Getenv(name))
	return err == nil && value
}
