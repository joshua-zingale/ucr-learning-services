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
func Parse(flagSet *flag.FlagSet, environmentVariablePrefix string, args []string) error {
	if flagSet.ErrorHandling() != flag.ExitOnError {
		panic("The input flagset must be set to ExitOnError because the other modes are not implemented!")
	}

	flagSet.String("config", "", "the path to a configuration file. The config format is line-separated names, followed by an equals sign, then the value; each name in the config file is the parameter name with dashes (-) replaced with underscores (_)")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	var flags set[*flag.Flag]
	flagSet.VisitAll(func(f *flag.Flag) {
		flags.Add(f)
	})

	var setFlags set[*flag.Flag]
	flagSet.Visit(func(f *flag.Flag) {
		setFlags.Add(f)
	})

	unsetFlags := flags.Minus(setFlags)

	configContent := ""

	if configFilePath := flagSet.Lookup("config").Value.String(); configFilePath != "" {
		var err error
		fileContent, err := os.ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		configContent = string(fileContent)
	}

	config, err := parseConfigContent(configContent)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	for _, flg := range unsetFlags {

		configName := paramNameToConfigName(flg.Name)
		if value, ok := config[configName]; ok {
			if err := flg.Value.Set(value); err != nil {
				return fmt.Errorf("setting value from config name '%e': %w", configName, err)
			}
		}

		envVarName := paramNameToEnvironName(environmentVariablePrefix, flg.Name)
		if envVarValue := os.Getenv(envVarName); envVarValue != "" {
			if err := flg.Value.Set(envVarValue); err != nil {
				return fmt.Errorf("setting value from environment variable '%s': %w", envVarName, err)
			}
		}

	}

	return nil
}

type set[T comparable] []T

func (s *set[T]) Add(e T) {
	*s = append(*s, e)
}

func (s set[T]) Minus(subtrahend set[T]) set[T] {
	var diff set[T]

outer:
	for _, min := range s {
		for _, sub := range subtrahend {
			if sub == min {
				continue outer
			}
		}
		diff.Add(min)
	}

	return diff
}

func (s set[T]) Contains(e T) bool {
	for _, e2 := range s {
		if e2 == e {
			return true
		}
	}
	return false
}

func paramNameToConfigName(s string) string {
	return strings.Replace(s, "-", "_", -1)
}

func paramNameToEnvironName(prefix string, s string) string {
	s = strings.Replace(s, "-", "_", -1)
	s = strings.ToUpper(s)
	return prefix + s
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
