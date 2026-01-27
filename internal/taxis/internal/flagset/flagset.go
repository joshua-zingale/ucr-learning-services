package flagset

import (
	"flag"
	"os"
	"strings"
)

type EnvironmentDefaultFlagSet struct {
	flagSet      *flag.FlagSet
	envVarPrefix string
}

func NewEnvironmentDefaultFlagSet(flagSet *flag.FlagSet, envVarPrefix string) *EnvironmentDefaultFlagSet {
	return &EnvironmentDefaultFlagSet{
		flagSet:      flagSet,
		envVarPrefix: envVarPrefix,
	}
}

func (edf *EnvironmentDefaultFlagSet) String(name string, value string, usage string) *string {
	envVarName := edf.paramNameToEnvironName(name)
	return edf.flagSet.String(name, envOr(envVarName, value, func(s string) string { return s }), usage)
}

func (edf *EnvironmentDefaultFlagSet) Bool(name string, value bool, usage string) *bool {
	envVarName := edf.paramNameToEnvironName(name)
	return edf.flagSet.Bool(name, envOr(envVarName, value, func(s string) bool {
		switch strings.ToLower(s) {
		case "true", "t", "yes", "y":
			return true
		default:
			return false
		}
	}), usage)
}

func (edf *EnvironmentDefaultFlagSet) paramNameToEnvironName(s string) string {
	s = strings.Replace(s, "-", "_", -1)
	s = strings.ToUpper(s)

	return strings.Join([]string{edf.envVarPrefix, s}, "")
}

func envOrString(key string, fallback string) string {
	if envVarValue := os.Getenv(key); envVarValue != "" {
		return envVarValue
	}
	return fallback
}

func envOr[T any](key string, fallback T, conversion func(string) T) T {
	if envVarValue := os.Getenv(key); envVarValue != "" {
		return conversion(envVarValue)
	}
	return fallback
}
