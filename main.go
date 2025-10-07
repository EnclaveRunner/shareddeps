package shareddeps

import (
	"errors"
	"fmt"
	"path/filepath"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	errLoadConfig    = "failed to load config"
	errDecode        = "unable to decode into struct:"
	errInvalidConfig = "config invalid:"
)

type BaseConfig struct {
	HumanReadableOutput bool   `mapstructure:"human_readable_output" validate:""`
	LogLevel            string `mapstructure:"log_level"             validate:"oneof=debug info warn error"`
	Port                int    `mapstructure:"port"                  validate:"numeric,min=1,max=65535"`
}

type HasBaseConfig interface {
	GetBase() *BaseConfig
}

func (b *BaseConfig) GetBase() *BaseConfig {
	return b
}

func tryLoadFile(filename string, paths ...string) {
	viper.SetConfigFile(filename)
	for _, path := range paths {
		configPath := filepath.Join(path, filename)
		if _, err := os.Stat(configPath); err == nil {
			// File exists, try to load it
			viper.SetConfigFile(configPath)
			if err := viper.MergeInConfig(); err == nil {
				return // Successfully loaded
			} else {
				log.Warn().Err(err).Str("file", configPath).Msg("Failed to load config file, trying next")

				continue
			}
		}
	}
}

var errConfigLoading = errors.New("failed to load configuration")

func configLoadingError(reason string, err error) error {
	return fmt.Errorf("%w: %s: %w", errConfigLoading, reason, err)
}

func LoadAppConfig[T HasBaseConfig](config T, serviceName, version string) error {
	// Set logger fields
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	log.Logger = log.With().
		Str("service", serviceName).
		Str("host", hostname).
		Str("version", version).
		Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	viper.SetDefault("human_readable_output", false)
	viper.SetDefault("log_level", "info")
	//nolint:mnd // Default port for HTTP
	viper.SetDefault("port", 80)

	// Configure enclave config file
	tryLoadFile("config.yaml", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.json", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.toml", "/etc/enclave", "$HOME/.enclave", ".")

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

	// Set log level and human readable output
	switch config.GetBase().LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	if config.GetBase().HumanReadableOutput {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false})
	}

	return nil
}
