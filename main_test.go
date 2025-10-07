package shareddeps

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

func (e *ExtendedConfig) GetBase() *BaseConfig {
	return &e.BaseConfig
}

//nolint:paralleltest // viper is not thread-safe
func TestTryLoadFile(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		setupFile   func(t *testing.T) string // returns file path
		expectPanic bool
	}{
		{
			name:     "load existing file",
			filename: "test-config.yaml",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `test_key: test_value`
				configPath := filepath.Join(tmpDir, "test-config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
		},
		{
			name:     "load non-existing file - should not panic",
			filename: "non-existent.yaml",
			setupFile: func(t *testing.T) string {
				return t.TempDir() // empty directory
			},
		},
	}

	//nolint:paralleltest // viper is not thread-safe
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for each test
			viper.Reset()

			tmpDir := tt.setupFile(t)

			// This should not panic even if file doesn't exist
			assert.NotPanics(t, func() {
				tryLoadFile(tt.filename, tmpDir)
			})
		})
	}
}

//nolint:paralleltest // viper is not thread-safe
func TestConfigLoadingError(t *testing.T) {
	baseErr := assert.AnError
	reason := "test reason"

	err := configLoadingError(reason, baseErr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), reason)
	assert.ErrorIs(t, err, errConfigLoading)
}

//nolint:paralleltest // viper is not thread-safe
func TestLoadAppConfigWithAdditionalAttributes(t *testing.T) {
	// Not parallel because viper uses global state
	tests := []struct {
		name          string
		setupConfig   func(t *testing.T) string
		expectError   bool
		errorContains string
		validateFunc  func(t *testing.T, config *ExtendedConfig)
	}{
		{
			name: "complete YAML config with all additional attributes",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: 8080
database_url: "postgres://user:pass@localhost:5432/mydb"
service_name: "test-service"
environment: "development"
max_connections: 100
timeout: 30
retry_attempts: 3
enable_metrics: true
enable_tracing: false
debug: true
allowed_hosts:
  - "localhost"
  - "127.0.0.1"
  - "test.example.com"
features:
  - "auth"
  - "logging"
  - "metrics"
database:
  host: "localhost"
  port: 5432
  username: "testuser"
  password: "password123"
  name: "testdb"
  ssl: true
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *ExtendedConfig) {
				assert.NotNil(t, config.HumanReadableOutput)
				assert.True(t, config.HumanReadableOutput)
				assert.Equal(t, "info", config.LogLevel)
				assert.Equal(t, 8080, config.Port)
				assert.Equal(t, "postgres://user:pass@localhost:5432/mydb", config.DatabaseURL)
				assert.Equal(t, "test-service", config.ServiceName)
				assert.Equal(t, "development", config.Environment)
				assert.Equal(t, 100, config.MaxConnections)
				assert.Equal(t, 30, config.Timeout)
				assert.Equal(t, 3, config.RetryAttempts)
				assert.True(t, config.EnableMetrics)
				assert.False(t, config.EnableTracing)
				assert.True(t, config.Debug)
				assert.Equal(t, []string{"localhost", "127.0.0.1", "test.example.com"}, config.AllowedHosts)
				assert.Equal(t, []string{"auth", "logging", "metrics"}, config.Features)
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
  "log_level": "debug",
  "port": 9090,
  "database_url": "mysql://user:pass@localhost:3306/mydb",
  "service_name": "json-service",
  "environment": "staging",
  "max_connections": 50,
  "timeout": 60,
  "retry_attempts": 5,
  "enable_metrics": false,
  "enable_tracing": true,
  "debug": false,
  "allowed_hosts": ["api.example.com", "staging.example.com"],
  "features": ["tracing", "logging"],
  "database": {
    "host": "db.staging.com",
    "port": 3306,
    "username": "stageuser",
    "password": "stagepass123",
    "name": "stagedb",
    "ssl": false
  }
}`
				configPath := filepath.Join(tmpDir, "config.json")
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *ExtendedConfig) {
				assert.NotNil(t, config.HumanReadableOutput)
				assert.False(t, config.HumanReadableOutput)
				assert.Equal(t, "debug", config.LogLevel)
				assert.Equal(t, 9090, config.Port)
				assert.Equal(t, "mysql://user:pass@localhost:3306/mydb", config.DatabaseURL)
				assert.Equal(t, "json-service", config.ServiceName)
				assert.Equal(t, "staging", config.Environment)
				assert.Equal(t, 50, config.MaxConnections)
				assert.Equal(t, 60, config.Timeout)
				assert.Equal(t, 5, config.RetryAttempts)
				assert.False(t, config.EnableMetrics)
				assert.True(t, config.EnableTracing)
				assert.False(t, config.Debug)
				assert.Equal(t, []string{"api.example.com", "staging.example.com"}, config.AllowedHosts)
				assert.Equal(t, []string{"tracing", "logging"}, config.Features)
				assert.Equal(t, "db.staging.com", config.Database.Host)
				assert.Equal(t, 3306, config.Database.Port)
				assert.Equal(t, "stageuser", config.Database.Username)
				assert.Equal(t, "stagepass123", config.Database.Password)
				assert.Equal(t, "stagedb", config.Database.Name)
				assert.False(t, config.Database.SSL)
			},
		},

		{
			name: "invalid database URL validation",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: 8080
database_url: "not-a-valid-url"
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
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			expectError:   true,
			errorContains: "config invalid",
		},
		{
			name: "invalid environment value",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: 8080
database_url: "postgres://user:pass@localhost:5432/mydb"
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
			expectError:   true,
			errorContains: "config invalid",
		},
		{
			name: "invalid feature values",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: 8080
database_url: "postgres://user:pass@localhost:5432/mydb"
service_name: "test-service"
environment: "development"
max_connections: 100
timeout: 30
allowed_hosts: ["localhost"]
features: ["auth", "invalid-feature", "logging"]
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
			expectError:   true,
			errorContains: "config invalid",
		},
	}

	//nolint:paralleltest // viper is not thread-safe
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

			// Test with ExtendedConfig
			config := &ExtendedConfig{}
			err = LoadAppConfig(config, "testing", "v0.0.1")

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, config)
				}
			}
		})
	}
}
