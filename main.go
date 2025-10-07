package shareddeps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var (
	errLoadConfig    = "failed to load configuration"
	errDecode        = "unable to decode into struct:"
	errInvalidConfig = "config invalid:"
)

type BaseConfig struct {
	HumanReadableOutput *bool  `mapstructure:"human_readable_output" validate:"required"`
	LogLevel            string `mapstructure:"log_level"             validate:"required,oneof=debug info warn error"`
	Port                string `mapstructure:"port"                  validate:"required,numeric,min=1,max=65535"`
}

func tryLoadFile(filename string, paths ...string) {
	// Try to load from each path
	for _, path := range paths {
		configPath := filepath.Join(path, filename)
		if _, err := os.Stat(configPath); err == nil {
			// File exists, try to load it
			viper.SetConfigFile(configPath)
			if err := viper.MergeInConfig(); err == nil {
				return // Successfully loaded
			}
		}
	}
	// Try current directory if not in paths
	if _, err := os.Stat(filename); err == nil {
		viper.SetConfigFile(filename)
		_ = viper.MergeInConfig() // Ignore error
	}
}

var errConfigLoading = errors.New(errLoadConfig)

func configLoadingError(reason string, err error) error {
	return fmt.Errorf("%w: %s: %w", errConfigLoading, reason, err)
}

// bindEnvironmentVariables automatically binds environment variables based on mapstructure tags
func bindEnvironmentVariables[T any](config *T, envPrefix string) {
	configType := reflect.TypeOf(config).Elem()
	bindStructFields(configType, envPrefix)
}

// bindStructFields recursively binds struct fields to environment variables
func bindStructFields(structType reflect.Type, envPrefix string) {
	for i := range structType.NumField() {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			bindStructFields(field.Type, envPrefix)

			continue
		}

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct {
			bindStructFields(field.Type, envPrefix)

			continue
		}

		// Get mapstructure tag
		mapstructureTag := field.Tag.Get("mapstructure")
		if mapstructureTag == "" || mapstructureTag == "-" {
			continue
		}

		// Create environment variable name
		envVarName := strings.ToUpper(envPrefix + "_" + mapstructureTag)

		// Bind the environment variable
		_ = viper.BindEnv(mapstructureTag, envVarName)
	}
}

func LoadAppConfig[T any](config *T) error {
	// Configure enclave config file
	tryLoadFile("config.yaml", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.json", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.toml", "/etc/enclave", "$HOME/.enclave", ".")

	// Configure environment variables
	viper.SetEnvPrefix("ENCLAVE")
	viper.AutomaticEnv()

	// Automatically bind environment variables based on struct fields
	bindEnvironmentVariables(config, "ENCLAVE")

	tryLoadFile(".env", ".")

	// Validate config
	unmarshalErr := viper.Unmarshal(config)
	if unmarshalErr != nil {
		return configLoadingError(errDecode, unmarshalErr)
	}

	validationErr := validator.New().Struct(config)
	if validationErr != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(validationErr, &validationErrs) {
			formattedErrs := make([]error, 0, len(validationErrs))
			for _, err := range validationErrs {
				formattedErrs = append(formattedErrs, err)
			}

			return configLoadingError(errInvalidConfig, errors.Join(formattedErrs...))
		}

		return configLoadingError(errInvalidConfig, validationErr)
	}

	return nil
}
