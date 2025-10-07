// nolint:errcheck,exhaustruct,gosec,intrange,paralleltest,varnamelen,golines // Test configuration
// with various struct patterns and environment variables
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
	HumanReadableOutput *bool  `mapstructure:"human_readable_output" validate:"required"`
	LogLevel            string `mapstructure:"log_level"             validate:"required,oneof=debug info warn error"`
	Port                string `mapstructure:"port"                  validate:"required,numeric,min=1,max=65535"`

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
		setupEnv      func()
		cleanupEnv    func()
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
			setupEnv:      nil,
			cleanupEnv:    nil,
			expectError:   false,
			errorContains: "",
		},
		{
			name: "valid config from JSON file",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `{
	"human_readable_output": true,
	"log_level": "debug",
	"port": "9090"
}`
				configPath := filepath.Join(tmpDir, "config.json")
				err := os.WriteFile(configPath, []byte(configContent), 0o600)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			setupEnv:      nil,
			cleanupEnv:    nil,
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
			setupEnv:      nil,
			cleanupEnv:    nil,
			expectError:   false,
			errorContains: "",
		},
		{
			name: "valid config from environment variables",
			setupConfig: func(t *testing.T) string {
				return t.TempDir() // empty dir, no config files
			},
			config: &BaseConfig{},
			setupEnv: func() {
				_ = os.Setenv("ENCLAVE_HUMAN_READABLE_OUTPUT", "true")
				_ = os.Setenv("ENCLAVE_LOG_LEVEL", "error")
				_ = os.Setenv("ENCLAVE_PORT", "5000")
			},
			cleanupEnv: func() {
				_ = os.Unsetenv("ENCLAVE_HUMAN_READABLE_OUTPUT")
				_ = os.Unsetenv("ENCLAVE_LOG_LEVEL")
				_ = os.Unsetenv("ENCLAVE_PORT")
			},
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
				err := os.WriteFile(envPath, []byte(envContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			config:      &BaseConfig{},
			expectError: false,
		},
		{
			name: "invalid log level",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: invalid_level
port: "8080"
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   true,
			errorContains: "config invalid",
		},
		{
			name: "invalid port",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
human_readable_output: true
log_level: info
port: "invalid_port"
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   true,
			errorContains: "config invalid",
		},
		{
			name: "missing required fields",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configContent := `
log_level: info
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			config:        &BaseConfig{},
			expectError:   true,
			errorContains: "config invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper instance for each test
			viper.Reset()

			// Setup environment if needed
			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			// Cleanup environment if needed
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

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

			// Test the function
			err = LoadAppConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify the config was loaded correctly
				assert.NotEmpty(t, tt.config.LogLevel)
				assert.NotEmpty(t, tt.config.Port)
				assert.Contains(t, []string{"debug", "info", "warn", "error"}, tt.config.LogLevel)
			}
		})
	}
}

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

func TestConfigLoadingError(t *testing.T) {
	baseErr := assert.AnError
	reason := "test reason"

	err := configLoadingError(reason, baseErr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), errLoadConfig)
	assert.Contains(t, err.Error(), reason)
	assert.ErrorIs(t, err, errConfigLoading)
}

// Benchmark the LoadAppConfig function
func BenchmarkLoadAppConfig(b *testing.B) {
	// Setup a temporary config file
	tmpDir := b.TempDir()
	configContent := `
human_readable_output: true
log_level: info
port: "8080"
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(b, err)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(b, err)
	defer func() {
		err := os.Chdir(originalWd)
		require.NoError(b, err)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		viper.Reset() // Reset for each iteration
		config := &BaseConfig{}
		err := LoadAppConfig(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestLoadAppConfigWithAdditionalAttributes tests loading extended configuration with additional
// attributes
func TestLoadAppConfigWithAdditionalAttributes(t *testing.T) {
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
port: "8080"
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
				assert.True(t, *config.HumanReadableOutput)
				assert.Equal(t, "info", config.LogLevel)
				assert.Equal(t, "8080", config.Port)
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
  "port": "9090",
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
				assert.False(t, *config.HumanReadableOutput)
				assert.Equal(t, "debug", config.LogLevel)
				assert.Equal(t, "9090", config.Port)
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
			name: "environment variables with additional attributes",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create a minimal config file to satisfy required nested/array fields
				configContent := `
allowed_hosts: ["localhost"]
features: ["auth"]
database:
  host: "localhost"
  port: 5432
  username: "envuser"
  password: "envpass123"
  name: "envdb"
`
				configPath := filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(configPath, []byte(configContent), 0o644)
				require.NoError(t, err)

				return tmpDir
			},
			expectError: false,
			validateFunc: func(t *testing.T, config *ExtendedConfig) {
				// Environment variables should override file values
				assert.NotNil(t, config.HumanReadableOutput)
				assert.True(t, *config.HumanReadableOutput)
				assert.Equal(t, "error", config.LogLevel)
				assert.Equal(t, "3000", config.Port)
				assert.Equal(t, "redis://localhost:6379", config.DatabaseURL)
				assert.Equal(t, "env-service", config.ServiceName)
				assert.Equal(t, "production", config.Environment)
				assert.Equal(t, 200, config.MaxConnections)
				assert.Equal(t, 120, config.Timeout)
				assert.Equal(t, 0, config.RetryAttempts)
				assert.True(t, config.EnableMetrics)
				assert.True(t, config.EnableTracing)
				assert.False(t, config.Debug)
				// File values for complex structures
				assert.Equal(t, []string{"localhost"}, config.AllowedHosts)
				assert.Equal(t, []string{"auth"}, config.Features)
				assert.Equal(t, "localhost", config.Database.Host)
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
port: "8080"
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
port: "8080"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper instance for each test
			viper.Reset()

			// Setup environment variables for the env test
			if tt.name == "environment variables with additional attributes" {
				os.Setenv("ENCLAVE_HUMAN_READABLE_OUTPUT", "true")
				os.Setenv("ENCLAVE_LOG_LEVEL", "error")
				os.Setenv("ENCLAVE_PORT", "3000")
				os.Setenv("ENCLAVE_DATABASE_URL", "redis://localhost:6379")
				os.Setenv("ENCLAVE_SERVICE_NAME", "env-service")
				os.Setenv("ENCLAVE_ENVIRONMENT", "production")
				os.Setenv("ENCLAVE_MAX_CONNECTIONS", "200")
				os.Setenv("ENCLAVE_TIMEOUT", "120")
				os.Setenv("ENCLAVE_RETRY_ATTEMPTS", "0")
				os.Setenv("ENCLAVE_ENABLE_METRICS", "true")
				os.Setenv("ENCLAVE_ENABLE_TRACING", "true")
				os.Setenv("ENCLAVE_DEBUG", "false")

				defer func() {
					os.Unsetenv("ENCLAVE_HUMAN_READABLE_OUTPUT")
					os.Unsetenv("ENCLAVE_LOG_LEVEL")
					os.Unsetenv("ENCLAVE_PORT")
					os.Unsetenv("ENCLAVE_DATABASE_URL")
					os.Unsetenv("ENCLAVE_SERVICE_NAME")
					os.Unsetenv("ENCLAVE_ENVIRONMENT")
					os.Unsetenv("ENCLAVE_MAX_CONNECTIONS")
					os.Unsetenv("ENCLAVE_TIMEOUT")
					os.Unsetenv("ENCLAVE_RETRY_ATTEMPTS")
					os.Unsetenv("ENCLAVE_ENABLE_METRICS")
					os.Unsetenv("ENCLAVE_ENABLE_TRACING")
					os.Unsetenv("ENCLAVE_DEBUG")
				}()
			}

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
			err = LoadAppConfig(config)

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

// TestGenericEnvironmentBinding tests that the generic environment binding works for any struct
func TestGenericEnvironmentBinding(t *testing.T) {
	// Define a custom config struct
	type CustomConfig struct {
		AppName     string `mapstructure:"app_name"`
		MaxRetries  int    `mapstructure:"max_retries"`
		EnableDebug bool   `mapstructure:"enable_debug"`
		ServerPort  string `mapstructure:"server_port"  validate:"required,numeric"`
	}

	// Reset viper
	viper.Reset()

	// Set custom environment variables
	os.Setenv("ENCLAVE_APP_NAME", "test-app")
	os.Setenv("ENCLAVE_MAX_RETRIES", "5")
	os.Setenv("ENCLAVE_ENABLE_DEBUG", "true")
	os.Setenv("ENCLAVE_SERVER_PORT", "9999")

	defer func() {
		os.Unsetenv("ENCLAVE_APP_NAME")
		os.Unsetenv("ENCLAVE_MAX_RETRIES")
		os.Unsetenv("ENCLAVE_ENABLE_DEBUG")
		os.Unsetenv("ENCLAVE_SERVER_PORT")
	}()

	// Test the generic LoadAppConfig function
	config := &CustomConfig{}
	err := LoadAppConfig(config)

	assert.NoError(t, err)
	assert.Equal(t, "test-app", config.AppName)
	assert.Equal(t, 5, config.MaxRetries)
	assert.True(t, config.EnableDebug)
	assert.Equal(t, "9999", config.ServerPort)
}

// TestMandatoryBaseConfigFields tests that BaseConfig fields are mandatory
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
port: "9090"
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

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

			// Test with BaseConfig
			config := &BaseConfig{}
			err = LoadAppConfig(config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config.HumanReadableOutput)
				assert.Equal(t, "debug", config.LogLevel)
				assert.Equal(t, "9090", config.Port)
			}
		})
	}
}
