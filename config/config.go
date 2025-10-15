package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type BaseConfig struct {
	HumanReadableOutput   bool   `mapstructure:"human_readable_output"  validate:""`
	LogLevel              string `mapstructure:"log_level"              validate:"oneof=debug info warn error"`
	ProductionEnvironment bool   `mapstructure:"production_environment" validate:""`
	Port                  int    `mapstructure:"port"                   validate:"numeric,min=1,max=65535"`
}

type HasBaseConfig interface {
	GetBase() *BaseConfig
}

func (b *BaseConfig) GetBase() *BaseConfig {
	return b
}

func tryLoadFile(filename string, paths ...string) {
	for _, path := range paths {
		// Expand environment variables
		expandedPath := os.ExpandEnv(path)
		configPath := filepath.Join(expandedPath, filename)
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

var Cfg = &BaseConfig{}

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
	viper.SetDefault("production_environment", true)

	// Configure enclave config file
	tryLoadFile("config.yaml", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.json", "/etc/enclave", "$HOME/.enclave", ".")
	tryLoadFile("config.toml", "/etc/enclave", "$HOME/.enclave", ".")

	// Validate config
	unmarshalErr := viper.Unmarshal(config)
	if unmarshalErr != nil {
		return configLoadingError("Unable to decode into struct", unmarshalErr)
	}

	validationErr := validator.New().Struct(config)
	if validationErr != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(validationErr, &validationErrors) {
			formattedErrs := make([]error, 0, len(validationErrors))
			for _, err := range validationErrors {
				formattedErrs = append(formattedErrs, err)
			}

			return configLoadingError("config invalid", errors.Join(formattedErrs...))
		}

		return configLoadingError("config invalid", validationErr)
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

	*Cfg = *config.GetBase()

	return nil
}
