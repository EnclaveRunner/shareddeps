package shareddeps

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EnclaveRunner/shareddeps/config"
	"github.com/EnclaveRunner/shareddeps/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) { //nolint:paralleltest // Simple test function
	// Test that Init function works with a BaseConfig that has defaults
	cfg := &config.BaseConfig{
		HumanReadableOutput:   false,
		LogLevel:              "info",
		Port:                  8080,
		ProductionEnvironment: false,
	}

	// This should not panic and should initialize the server
	assert.NotPanics(t, func() {
		Init(cfg, "testservice", "1.0.0")
	})

	// Server should be initialized
	assert.NotNil(t, Server)
}

func TestConfigPackageIntegration(t *testing.T) { //nolint:paralleltest // Simple test function
	// Test that we can create and use config types from the config package
	cfg := &config.BaseConfig{
		HumanReadableOutput:   false,
		LogLevel:              "info",
		Port:                  8080,
		ProductionEnvironment: false,
	}
	assert.NotNil(t, cfg)

	// Test that the BaseConfig has the expected fields set
	assert.False(t, cfg.HumanReadableOutput)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 8080, cfg.Port)
}

func TestZerologMiddleware(t *testing.T) { //nolint:paralleltest // Simple test function
	// Set gin to test mode for testing
	gin.SetMode(gin.TestMode)

	// Create a test router with our middleware
	router := gin.New()
	router.Use(middleware.Zerolog())

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test the route works with middleware
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
}

func TestGlobalServerInitialization(t *testing.T) { //nolint:paralleltest // Simple test function
	// Set gin to test mode for testing
	gin.SetMode(gin.TestMode)

	// Initialize with a valid config
	cfg := &config.BaseConfig{
		HumanReadableOutput:   false,
		LogLevel:              "info",
		Port:                  8080,
		ProductionEnvironment: false,
	}

	// Initialize should set up the global Server
	Init(cfg, "testservice", "1.0.0")

	// Global Server should be initialized
	assert.NotNil(t, Server)

	// Should be able to add routes to the global server
	Server.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Test the route works
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/health", http.NoBody)
	Server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}
