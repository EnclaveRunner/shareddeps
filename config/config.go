package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

var Cfg = &BaseConfig{}

type DefaultValue struct {
	Key   string
	Value any
}

type ConfigError struct {
	Msg string
	Err error
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func (e ConfigError) Unwrap() error {
	return e.Err
}

type ValidationError struct {
	Err validator.FieldError
}

func (e ValidationError) Error() string {
	switch e.Err.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required", e.Err.Namespace())
	case "oneof":
		return fmt.Sprintf(
			"Field '%s' must be one of the following: %s",
			e.Err.Namespace(),
			e.Err.Param(),
		)
	case "min":
		return fmt.Sprintf(
			"Field '%s' must be at least %s",
			e.Err.Namespace(),
			e.Err.Param(),
		)
	case "max":
		return fmt.Sprintf(
			"Field '%s' must be at most %s",
			e.Err.Namespace(),
			e.Err.Param(),
		)
	case "numeric":
		return fmt.Sprintf("Field '%s' must be a numeric value", e.Err.Namespace())
	default:
		return fmt.Sprintf("Field '%s' is invalid", e.Err.Namespace())
	}
}

func (e ValidationError) Unwrap() error {
	return e.Err
}

func LoadAppConfig[T HasBaseConfig](
	config T,
	serviceName, version string, defaults ...DefaultValue,
) error {
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

	v := viper.NewWithOptions(
		viper.ExperimentalBindStruct(),
		viper.EnvKeyReplacer(strings.NewReplacer(".", "_")),
	)

	// Set base config defaults
	v.SetDefault("human_readable_output", false)
	v.SetDefault("log_level", "info")
	//nolint:mnd // Default port for HTTP
	v.SetDefault("port", 8080)
	v.SetDefault("production_environment", true)

	// Set passed defaults
	for _, def := range defaults {
		v.SetDefault(def.Key, def.Value)
	}

	// Configure environment variables
	v.SetEnvPrefix("ENCLAVE")
	v.AutomaticEnv()

	// Configure enclave config file
	v.SetConfigName(serviceName + ".yml")
	v.SetConfigType("yaml")

	// Add config paths in order of precedence (last added has lowest
	// precedence)
	// 1. Current directory (highest precedence)
	v.AddConfigPath(".")

	// 2. Home directory
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".enclave"))
	}

	// 3. System-wide config (lowest precedence)
	v.AddConfigPath("/etc/enclave")

	// Read config file (it's okay if it doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Debug().Err(err).Msg("Config file not found")
			// Config file not found; ignore error and use defaults/env vars
		} else {
			return ConfigError{
				Msg: "Failed to load config file",
				Err: err,
			}
		}
	}

	// Validate config
	unmarshalErr := v.Unmarshal(config)
	if unmarshalErr != nil {
		return ConfigError{
			Msg: "Unable to decode into struct",
			Err: unmarshalErr,
		}
	}

	validationErr := validator.New().Struct(config)
	if validationErr != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(validationErr, &validationErrors) {
			formattedErrs := make([]error, 0, len(validationErrors))
			for _, err := range validationErrors {
				formattedErrs = append(formattedErrs, ValidationError{Err: err})
			}

			return ConfigError{
				Msg: "Config is invalid",
				Err: errors.Join(formattedErrs...),
			}
		}

		return ConfigError{
			Msg: "Config cannot be verified",
			Err: validationErr,
		}
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
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false},
		)
	}

	*Cfg = *config.GetBase()

	return nil
}
