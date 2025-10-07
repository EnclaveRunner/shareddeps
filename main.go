package shareddeps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

func LoadAppConfig[T any](config *T) error {
	// Configure enclave config file
	tryLoadFile("config.yaml", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.json", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.toml", "/etc/enclave", "$HOME/.enclave", ".")

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
