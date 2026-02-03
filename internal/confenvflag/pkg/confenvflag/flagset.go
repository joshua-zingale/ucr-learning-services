package confenvflag

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Parses arguments, in order, from arguments, environment variables, then a config file.
// Thus arguments overwrite environment variables, which overwrite the config file.
type ConfEnvFlagSet struct {
	FlagSet      *flag.FlagSet
	envVarPrefix string

	envVars []EnvironmentVariable
	config  map[string]string
}

type EnvironmentVariable struct {
	name  string
	type_ string
	usage string
}

func NewConfEnvFlagSet(flagSet *flag.FlagSet, envVarPrefix string, configContent string) (*ConfEnvFlagSet, error) {
	config, err := parseConfigContent(configContent)
	if err != nil {
		return nil, err
	}
	return &ConfEnvFlagSet{
		FlagSet:      flagSet,
		envVarPrefix: envVarPrefix,
		config:       config,
	}, nil
}

func ParseConfigArgument(args []string) (string, []string, error) {
	const CONFIG_NOT_FOUND = -1
	configIdx := CONFIG_NOT_FOUND
	for i := range args {
		if args[i] == "-config" || args[i] == "--config" {
			if configIdx != CONFIG_NOT_FOUND {
				return "", nil, fmt.Errorf("found argument '%s', but '%s' was already specified", args[i], args[configIdx])
			}
			configIdx = i
		}
	}
	if configIdx == CONFIG_NOT_FOUND {
		return "", args, nil
	}

	if configIdx == len(args)-1 {
		return "", nil, fmt.Errorf("expected path to config file after %s", args[configIdx])
	}

	var newArgs []string
	for i, arg := range args {
		if i == configIdx || i == configIdx+1 {
			continue
		}
		newArgs = append(newArgs, arg)
	}

	return args[configIdx+1], newArgs, nil
}

func (edf *ConfEnvFlagSet) Parse(args []string) error {
	flagNames := make(map[string]struct{})
	edf.FlagSet.VisitAll(func(f *flag.Flag) {
		flagNames[f.Name] = struct{}{}
	})
	for configName := range edf.config {
		if _, ok := flagNames[configNameToParamName(configName)]; !ok {
			return fmt.Errorf("unknown key in config: %s", configName)
		}
	}

	return edf.FlagSet.Parse(args)
}

func (edf *ConfEnvFlagSet) String(name string, value string, usage string) *string {
	edf.envVars = append(edf.envVars, EnvironmentVariable{name: edf.paramNameToEnvironName(name), type_: "str", usage: usage})
	return edf.FlagSet.String(name, resolveDefault(edf, name, value, func(s string) string { return s }), usage)
}

func (edf *ConfEnvFlagSet) Bool(name string, value bool, usage string) *bool {

	edf.envVars = append(edf.envVars, EnvironmentVariable{name: edf.paramNameToEnvironName(name), type_: "str", usage: fmt.Sprintf("'true', 't', 'yes', and 'y' set and all else unsets. %s", usage)})
	return edf.FlagSet.Bool(name, resolveDefault(edf, name, value, func(s string) bool {
		switch strings.ToLower(s) {
		case "true", "t", "yes", "y":
			return true
		default:
			return false
		}
	}), usage)
}

func (edf *ConfEnvFlagSet) PrintSupportedEnvironmentVariables() {
	for _, envVar := range edf.envVars {
		fmt.Printf("%s|%s|%s\n", envVar.name, envVar.type_, envVar.usage)
	}
}

func (edf *ConfEnvFlagSet) paramNameToEnvironName(s string) string {
	s = strings.Replace(s, "-", "_", -1)
	s = strings.ToUpper(s)

	return strings.Join([]string{edf.envVarPrefix, s}, "")
}

func configNameToParamName(s string) string {
	return strings.Replace(s, "_", "-", -1)
}

func paramNameToConfigName(s string) string {
	return strings.Replace(s, "-", "_", -1)
}

func parseConfigContent(content string) (map[string]string, error) {
	config := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if strings.Contains(key, " ") {
				return nil, fmt.Errorf("invalid key '%s': must not contain white space", key)
			}
			config[key] = val
		}
	}
	return config, scanner.Err()
}

func resolveDefault[T any](edf *ConfEnvFlagSet, name string, fallback T, conversion func(string) T) T {

	if envVarValue := os.Getenv(edf.paramNameToEnvironName(name)); envVarValue != "" {
		return conversion(envVarValue)
	}
	if value, ok := edf.config[paramNameToConfigName(name)]; ok {
		return conversion(value)
	}
	return fallback
}
