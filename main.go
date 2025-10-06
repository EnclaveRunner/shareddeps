package shareddeps

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type BaseConfig struct {
	HumanReadableOutput bool   `mapstructure:"human_readable_output" validate:"required"`
	LogLevel            string `mapstructure:"log_level"             validate:"required,oneof=debug info warn error"`
	Port                string `mapstructure:"port"                  validate:"required,port"`
}

func tryLoadFile(filename string, paths ...string) {
	viper.SetConfigFile(filename)
	for _, path := range paths {
		viper.AddConfigPath(path)
	}

	err := viper.MergeInConfig()
	var fileNotFoundError *viper.ConfigFileNotFoundError
	if err != nil && errors.As(err, fileNotFoundError) {
		fmt.Println(
			fmt.Errorf("failed to load config file: %w", err),
		)
	}
}

var errConfigLoading = errors.New("failed to load configuration")

func configLoadingError(reason string, err error) error {
	return fmt.Errorf("%w: %s: %w", errConfigLoading, reason, err)
}

func LoadAppConfig(config BaseConfig) error {
	// Configure enclave config file
	tryLoadFile("config.yaml", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.json", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.toml", "/etc/enclave", "$HOME/.enclave", ".")

	// Configure environment variables
	viper.SetEnvPrefix("ENCLAVE")
	viper.AutomaticEnv()
	tryLoadFile(".env", ".")

	// Validate config
	unmarshalErr := viper.Unmarshal(config)
	if unmarshalErr != nil {
		return configLoadingError("Unable to decode into struct", unmarshalErr)
	}

	var validationErr *validator.ValidationErrors
	errors.As(validator.New().Struct(config), &validationErr)

	formattedErrs := make([]error, 0, len(*validationErr))
	for _, err := range *validationErr {
		formattedErrs = append(formattedErrs, err)
	}

	if len(*validationErr) > 0 {
		return configLoadingError("Config invalid", errors.Join(formattedErrs...))
	}

	return nil
}
