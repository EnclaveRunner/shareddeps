// environment variables loading
//
//nolint:paralleltest // Config loading code is not thread-safe due to modification of process-wide environment variables during testspackage config
package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig is a test implementation that embeds BaseConfig
type TestConfig struct {
	BaseConfig `mapstructure:",squash"`

	TestField string `mapstructure:"test_field" validate:"required"`
}

func (t *TestConfig) GetBase() *BaseConfig {
	return &t.BaseConfig
}

// MinimalConfig only has BaseConfig
type MinimalConfig struct {
	BaseConfig `mapstructure:",squash"`
}

func (m *MinimalConfig) GetBase() *BaseConfig {
	return &m.BaseConfig
}

// NestedStruct is a nested configuration struct
type NestedStruct struct {
	NestedField string `mapstructure:"nested_field" validate:"required"`
	OptionalInt int    `mapstructure:"optional_int"`
}

// ConfigWithNested has nested configuration
type ConfigWithNested struct {
	BaseConfig `mapstructure:",squash"`

	Database NestedStruct `mapstructure:"database" validate:"required"`
}

func (c *ConfigWithNested) GetBase() *BaseConfig {
	return &c.BaseConfig
}

func TestLoadAppConfig_WithDefaults(t *testing.T) {
	// Reset environment
	clearEnv(t)

	config := &MinimalConfig{}
	err := LoadAppConfig(config, "test-service", "1.0.0")

	require.NoError(t, err)
	assert.False(t, config.HumanReadableOutput)
	assert.Equal(t, "info", config.LogLevel)
	assert.True(t, config.ProductionEnvironment)
	assert.Equal(t, 8080, config.Port)
}

func TestLoadAppConfig_WithEnvironmentVariables(t *testing.T) {
	clearEnv(t)

	// Set environment variables
	t.Setenv("ENCLAVE_LOG_LEVEL", "debug")
	t.Setenv("ENCLAVE_PORT", "9000")
	t.Setenv("ENCLAVE_HUMAN_READABLE_OUTPUT", "true")
	t.Setenv("ENCLAVE_PRODUCTION_ENVIRONMENT", "false")

	config := &MinimalConfig{}
	err := LoadAppConfig(config, "test-service", "1.0.0")

	require.NoError(t, err)
	assert.True(t, config.HumanReadableOutput)
	assert.Equal(t, "debug", config.LogLevel)
	assert.False(t, config.ProductionEnvironment)
	assert.Equal(t, 9000, config.Port)
	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
}

func TestLoadAppConfig_WithCustomDefaults(t *testing.T) {
	clearEnv(t)

	config := &TestConfig{}
	defaults := []DefaultValue{
		{Key: "test_field", Value: "default_value"},
		{Key: "port", Value: "3000"},
	}

	err := LoadAppConfig(config, "test-service", "1.0.0", defaults...)

	require.NoError(t, err)
	assert.Equal(t, "default_value", config.TestField)
	assert.Equal(t, 3000, config.Port)
}

func TestLoadAppConfig_InvalidLogLevel(t *testing.T) {
	clearEnv(t)

	t.Setenv("ENCLAVE_LOG_LEVEL", "invalid")

	config := &MinimalConfig{}
	err := LoadAppConfig(config, "test-service", "1.0.0")

	require.Error(t, err)
	var configErr ConfigError
	assert.True(t, errors.As(err, &configErr))
	assert.Contains(t, err.Error(), "Config is invalid")
	assert.Contains(t, err.Error(), "LogLevel")
}

func TestLoadAppConfig_InvalidPort_TooLow(t *testing.T) {
	clearEnv(t)

	t.Setenv("ENCLAVE_PORT", "0")

	config := &MinimalConfig{}
	err := LoadAppConfig(config, "test-service", "1.0.0")

	require.Error(t, err)
	var configErr ConfigError
	assert.True(t, errors.As(err, &configErr))
	assert.Contains(t, err.Error(), "Config is invalid")
}

func TestLoadAppConfig_InvalidPort_TooHigh(t *testing.T) {
	clearEnv(t)

	t.Setenv("ENCLAVE_PORT", "65536")

	config := &MinimalConfig{}
	err := LoadAppConfig(config, "test-service", "1.0.0")

	require.Error(t, err)
	var configErr ConfigError
	assert.True(t, errors.As(err, &configErr))
	assert.Contains(t, err.Error(), "Config is invalid")
}

func TestLoadAppConfig_MissingRequiredField(t *testing.T) {
	clearEnv(t)

	config := &TestConfig{}
	err := LoadAppConfig(config, "test-service", "1.0.0")

	require.Error(t, err)
	var configErr ConfigError
	assert.True(t, errors.As(err, &configErr))
	assert.Contains(t, err.Error(), "Config is invalid")
	assert.Contains(t, err.Error(), "TestField")
}

func TestLoadAppConfig_AllLogLevels(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      string
		expectedLevel zerolog.Level
	}{
		{"debug", "debug", zerolog.DebugLevel},
		{"info", "info", zerolog.InfoLevel},
		{"warn", "warn", zerolog.WarnLevel},
		{"error", "error", zerolog.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)

			t.Setenv("ENCLAVE_LOG_LEVEL", tt.logLevel)

			config := &MinimalConfig{}
			err := LoadAppConfig(config, "test-service", "1.0.0")

			require.NoError(t, err)
			assert.Equal(t, tt.logLevel, config.LogLevel)
			assert.Equal(t, tt.expectedLevel, zerolog.GlobalLevel())
		})
	}
}

func TestLoadAppConfig_WithConfigFile(t *testing.T) {
	clearEnv(t)

	// Create a temporary directory and config file
	tmpDir := t.TempDir()
	configContent := `
log_level: warn
port: 5000
human_readable_output: true
production_environment: false
test_field: from_file
`
	configPath := filepath.Join(tmpDir, "test-service.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // defer in test
	defer os.Chdir(originalDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	config := &TestConfig{}
	err = LoadAppConfig(config, "test-service", "1.0.0")

	require.NoError(t, err)
	assert.Equal(t, "warn", config.LogLevel)
	assert.Equal(t, 5000, config.Port)
	assert.True(t, config.HumanReadableOutput)
	assert.False(t, config.ProductionEnvironment)
	assert.Equal(t, "from_file", config.TestField)
}

func TestLoadAppConfig_EnvOverridesFile(t *testing.T) {
	clearEnv(t)

	// Create a temporary directory and config file
	tmpDir := t.TempDir()
	configContent := `
log_level: warn
port: 5000
test_field: from_file
`
	configPath := filepath.Join(tmpDir, "test-service.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Set environment variable to override
	t.Setenv("ENCLAVE_PORT", "7000")

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // defer in test
	defer os.Chdir(originalDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	config := &TestConfig{}
	err = LoadAppConfig(config, "test-service", "1.0.0")

	require.NoError(t, err)
	assert.Equal(t, "warn", config.LogLevel)
	assert.Equal(
		t,
		7000,
		config.Port,
	) // Environment variable should override file
	assert.Equal(t, "from_file", config.TestField)
}

func TestLoadAppConfig_SetsGlobalConfig(t *testing.T) {
	clearEnv(t)

	t.Setenv("ENCLAVE_PORT", "4000")

	config := &MinimalConfig{}
	err := LoadAppConfig(config, "test-service", "1.0.0")

	require.NoError(t, err)
	assert.Equal(t, 4000, Cfg.Port)
	assert.Equal(t, config.Port, Cfg.Port)
}

func TestGetBase(t *testing.T) {
	base := &BaseConfig{
		LogLevel: "debug",
		Port:     9000,
	}

	assert.Equal(t, base, base.GetBase())
}

func TestLoadAppConfig_WithNestedConfig(t *testing.T) {
	clearEnv(t)

	// Create a temporary directory and config file
	tmpDir := t.TempDir()
	configContent := `
log_level: info
port: 8080
database:
  nested_field: test_value
  optional_int: 42
`
	configPath := filepath.Join(tmpDir, "test-service.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // defer in test
	defer os.Chdir(originalDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	config := &ConfigWithNested{}
	err = LoadAppConfig(config, "test-service", "1.0.0")

	require.NoError(t, err)
	assert.Equal(t, "test_value", config.Database.NestedField)
	assert.Equal(t, 42, config.Database.OptionalInt)
}

func TestLoadAppConfig_NestedConfigFromEnv(t *testing.T) {
	clearEnv(t)

	// Set nested fields via environment variables
	t.Setenv("ENCLAVE_DATABASE_NESTED_FIELD", "env_value")
	t.Setenv("ENCLAVE_DATABASE_OPTIONAL_INT", "100")

	config := &ConfigWithNested{}
	err := LoadAppConfig(config, "test-service", "1.0.0")

	require.NoError(t, err)
	assert.Equal(t, "env_value", config.Database.NestedField)
	assert.Equal(t, 100, config.Database.OptionalInt)
}

func TestLoadAppConfig_MissingNestedRequiredField(t *testing.T) {
	clearEnv(t)

	config := &ConfigWithNested{}
	err := LoadAppConfig(config, "test-service", "1.0.0")

	require.Error(t, err)
	var configErr ConfigError
	assert.True(t, errors.As(err, &configErr))
	assert.Contains(t, err.Error(), "Config is invalid")
	assert.Contains(t, err.Error(), "NestedField")
}

// Helper function to clear relevant environment variables
func clearEnv(t *testing.T) {
	t.Helper()
	_ = os.Unsetenv("ENCLAVE_LOG_LEVEL")
	_ = os.Unsetenv("ENCLAVE_PORT")
	_ = os.Unsetenv("ENCLAVE_HUMAN_READABLE_OUTPUT")
	_ = os.Unsetenv("ENCLAVE_PRODUCTION_ENVIRONMENT")
	_ = os.Unsetenv("ENCLAVE_TEST_FIELD")
	_ = os.Unsetenv("ENCLAVE_DATABASE_NESTED_FIELD")
	_ = os.Unsetenv("ENCLAVE_DATABASE_OPTIONAL_INT")

	// Reset global config
	Cfg = &BaseConfig{}

	// Reset log level to info for consistency
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
