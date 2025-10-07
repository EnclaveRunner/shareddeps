//nolint:exhaustruct,gosec,paralleltest,varnamelen // Test configuration file
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExtendedConfig extends BaseConfig with additional attributes for testing
type ExtendedConfig struct {
	BaseConfig `mapstructure:",squash"`

	// Additional string fields
	DatabaseURL string `mapstructure:"database_url" validate:"required,url"`
	ServiceName string `mapstructure:"service_name" validate:"required,min=3,max=50"`
	Environment string `mapstructure:"environment"  validate:"required,oneof=development staging production"`

	// Additional numeric fields
	MaxConnections int `mapstructure:"max_connections" validate:"required,min=1,max=1000"`
	Timeout        int `mapstructure:"timeout"         validate:"required,min=1,max=300"`
	RetryAttempts  int `mapstructure:"retry_attempts"  validate:"min=0,max=10"`

	// Additional boolean fields
	EnableMetrics bool `mapstructure:"enable_metrics"`
	EnableTracing bool `mapstructure:"enable_tracing"`
	Debug         bool `mapstructure:"debug"`

	// Array/slice fields
	AllowedHosts []string `mapstructure:"allowed_hosts" validate:"required,dive,hostname_rfc1123"`
	Features     []string `mapstructure:"features"      validate:"dive,oneof=auth logging metrics tracing"`

	// Nested object
	Database struct {
		Host     string `mapstructure:"host"     validate:"required,hostname"`
		Port     int    `mapstructure:"port"     validate:"required,min=1,max=65535"`
		Username string `mapstructure:"username" validate:"required"`
		Password string `mapstructure:"password" validate:"required,min=8"`
		Name     string `mapstructure:"name"     validate:"required"`
		SSL      bool   `mapstructure:"ssl"`
	} `mapstructure:"database"`
}

func TestLoadAppConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupConfig   func(t *testing.T) string // returns temp dir path
		config        *BaseConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "valid config from YAML file",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: "8080"
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   false,
			errorContains: "",
		},
		{
			name: "valid config from JSON file",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `{
  "human_readable_output": false,
  "log_level": "debug",
  "port": "3000"
}`
				configPath := filepath.Join(tmpDir, "config.json")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   false,
			errorContains: "",
		},
		{
			name: "valid config from TOML file",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `human_readable_output = false
log_level = "warn"
port = "3000"`
				configPath := filepath.Join(tmpDir, "config.toml")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   false,
			errorContains: "",
		},
		{
			name: "valid config from .env file",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envContent := `human_readable_output=false
log_level=info
port=4000`
				envPath := filepath.Join(tmpDir, ".env")
				err := os.WriteFile(envPath, []byte(envContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   false,
			errorContains: "",
		},
		{
			name: "invalid log level",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: invalid
port: "8080"
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   true,
			errorContains: "LogLevel",
		},
		{
			name: "invalid port",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: "invalid"
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   true,
			errorContains: "Port",
		},
		{
			name: "missing required fields",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
log_level: info
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   true,
			errorContains: "HumanReadableOutput",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper instance for each test
			viper.Reset()

			// Setup config files
			tmpDir := tt.setupConfig(t)

			// Change to temp directory to test file loading
			originalWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(originalWd)
				require.NoError(t, err)
			}()
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			// Test LoadAppConfig
			err = LoadAppConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				// Only verify config if no error occurred
				if err == nil && tt.config != nil {
					assert.NotNil(t, tt.config.HumanReadableOutput)
					assert.NotEmpty(t, tt.config.LogLevel)
					assert.NotEmpty(t, tt.config.Port)
				}
			}
		})
	}
}

func TestTryLoadFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "load existing file",
			filename: "test-config.yaml",
		},
		{
			name:     "load non-existing file - should not panic",
			filename: "non-existent-config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Only create file for the first test
			if tt.name == "load existing file" {
				configContent := "test: value"
				configPath := filepath.Join(tmpDir, tt.filename)
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)
			}

			// Should not panic
			assert.NotPanics(t, func() {
				tryLoadFile(tt.filename, tmpDir)
			})
		})
	}
}

func TestConfigLoadingError(t *testing.T) {
	baseErr := configLoadingError("test reason", assert.AnError)
	assert.Error(t, baseErr)
	assert.Contains(t, baseErr.Error(), "failed to load configuration")
	assert.Contains(t, baseErr.Error(), "test reason")
}

func TestLoadAppConfigWithAdditionalAttributes(t *testing.T) {
	tests := []struct {
		name         string
		setupConfig  func(t *testing.T) string
		expectError  bool
		validateFunc func(t *testing.T, config *ExtendedConfig)
	}{
		{
			name: "complete YAML config with all additional attributes",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
# BaseConfig fields
human_readable_output: true
log_level: info
port: "8080"

# Extended fields
database_url: "postgres://user:pass@localhost:5432/db"
service_name: "test-service"
environment: "development"
max_connections: 100
timeout: 30
retry_attempts: 3
enable_metrics: true
enable_tracing: false
debug: true

# Arrays
allowed_hosts:
  - "localhost"
  - "example.com"
features:
  - "auth"
  - "logging"
  - "metrics"

# Nested object
database:
  host: "localhost"
  port: 5432
  username: "testuser"
  password: "password123"
  name: "testdb"
  ssl: true
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *ExtendedConfig) {
				// BaseConfig fields
				assert.NotNil(t, config.HumanReadableOutput)
				assert.True(t, *config.HumanReadableOutput)
				assert.Equal(t, "info", config.LogLevel)
				assert.Equal(t, "8080", config.Port)

				// Extended fields
				assert.Equal(t, "postgres://user:pass@localhost:5432/db", config.DatabaseURL)
				assert.Equal(t, "test-service", config.ServiceName)
				assert.Equal(t, "development", config.Environment)
				assert.Equal(t, 100, config.MaxConnections)
				assert.Equal(t, 30, config.Timeout)
				assert.Equal(t, 3, config.RetryAttempts)
				assert.True(t, config.EnableMetrics)
				assert.False(t, config.EnableTracing)
				assert.True(t, config.Debug)

				// Arrays
				assert.Equal(t, []string{"localhost", "example.com"}, config.AllowedHosts)
				assert.Equal(t, []string{"auth", "logging", "metrics"}, config.Features)

				// Nested object
				assert.Equal(t, "localhost", config.Database.Host)
				assert.Equal(t, 5432, config.Database.Port)
				assert.Equal(t, "testuser", config.Database.Username)
				assert.Equal(t, "password123", config.Database.Password)
				assert.Equal(t, "testdb", config.Database.Name)
				assert.True(t, config.Database.SSL)
			},
		},
		{
			name: "JSON config with additional attributes",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `{
  "human_readable_output": false,
  "log_level": "warn",
  "port": "9000",
  "database_url": "redis://localhost:6379",
  "service_name": "json-service",
  "environment": "staging",
  "max_connections": 50,
  "timeout": 60,
  "retry_attempts": 5,
  "enable_metrics": false,
  "enable_tracing": true,
  "debug": false,
  "allowed_hosts": ["api.example.com"],
  "features": ["auth", "tracing"],
  "database": {
    "host": "db.example.com",
    "port": 3306,
    "username": "jsonuser",
    "password": "jsonpass123",
    "name": "jsondb",
    "ssl": false
  }
}`
				configPath := filepath.Join(tmpDir, "config.json")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *ExtendedConfig) {
				assert.NotNil(t, config.HumanReadableOutput)
				assert.False(t, *config.HumanReadableOutput)
				assert.Equal(t, "warn", config.LogLevel)
				assert.Equal(t, "redis://localhost:6379", config.DatabaseURL)
				assert.Equal(t, "json-service", config.ServiceName)
				assert.Equal(t, "staging", config.Environment)
			},
		},
		{
			name: "invalid database URL validation",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: "8080"
database_url: "invalid-url"
service_name: "test-service"
environment: "development"
max_connections: 100
timeout: 30
allowed_hosts: ["localhost"]
features: ["auth"]
database:
  host: "localhost"
  port: 5432
  username: "testuser"
  password: "password123"
  name: "testdb"
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			expectError: true,
		},
		{
			name: "invalid environment value",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: "8080"
database_url: "postgres://user:pass@localhost:5432/db"
service_name: "test-service"
environment: "invalid-env"
max_connections: 100
timeout: 30
allowed_hosts: ["localhost"]
features: ["auth"]
database:
  host: "localhost"
  port: 5432
  username: "testuser"
  password: "password123"
  name: "testdb"
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			expectError: true,
		},
		{
			name: "invalid feature values",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: "8080"
database_url: "postgres://user:pass@localhost:5432/db"
service_name: "test-service"
environment: "development"
max_connections: 100
timeout: 30
allowed_hosts: ["localhost"]
features: ["auth", "invalid-feature"]
database:
  host: "localhost"
  port: 5432
  username: "testuser"
  password: "password123"
  name: "testdb"
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper instance for each test
			viper.Reset()

			// Setup config files
			tmpDir := tt.setupConfig(t)

			// Change to temp directory to test file loading
			originalWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(originalWd)
				require.NoError(t, err)
			}()
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			// Test LoadAppConfig with ExtendedConfig
			config := &ExtendedConfig{}
			err = LoadAppConfig(config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, config)
				}
			}
		})
	}
}

func TestMandatoryBaseConfigFields(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		expectError   bool
		errorContains string
	}{
		{
			name: "missing human_readable_output should fail",
			configContent: `
log_level: info
port: "8080"
`,
			expectError:   true,
			errorContains: "HumanReadableOutput",
		},
		{
			name: "missing log_level should fail",
			configContent: `
human_readable_output: true
port: "8080"
`,
			expectError:   true,
			errorContains: "LogLevel",
		},
		{
			name: "missing port should fail",
			configContent: `
human_readable_output: true
log_level: info
`,
			expectError:   true,
			errorContains: "Port",
		},
		{
			name: "all fields present should pass",
			configContent: `
human_readable_output: false
log_level: debug
port: "3000"
`,
			expectError:   false,
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper instance for each test
			viper.Reset()

			// Setup config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.configContent), 0o644)
			require.NoError(t, err)

			// Change to temp directory
			originalWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(originalWd)
				require.NoError(t, err)
			}()
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			// Test LoadAppConfig
			config := &BaseConfig{}
			err = LoadAppConfig(config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				// Verify required fields are set
				assert.NotNil(t, config.HumanReadableOutput)
				assert.NotEmpty(t, config.LogLevel)
				assert.NotEmpty(t, config.Port)
			}
		})
	}
}
