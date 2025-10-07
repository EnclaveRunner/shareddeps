package shareddeps

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EnclaveRunner/shareddeps/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) { //nolint:paralleltest // Simple test function
	// Test that InitConfig function exists and can be called
	// Note: This will likely fail without actual config files, but that's expected
	cfg := &config.BaseConfig{} //nolint:exhaustruct // Test struct initialization
	server, err := InitConfig(cfg)
	// We expect an error because there are no config files set up
	assert.Error(t, err)
	assert.Nil(t, server)
}

func TestConfigPackageIntegration(t *testing.T) { //nolint:paralleltest // Simple test function
	// Test that we can create and use config types from the config package
	cfg := &config.BaseConfig{} //nolint:exhaustruct // Test struct initialization
	assert.NotNil(t, cfg)

	// Test that the BaseConfig has the expected fields
	assert.Nil(t, cfg.HumanReadableOutput) // Should be nil initially
	assert.Empty(t, cfg.LogLevel)          // Should be empty initially
	assert.Empty(t, cfg.Port)              // Should be empty initially
}

func TestNewRoute(t *testing.T) { //nolint:paralleltest // Simple test function
	// Set gin to test mode for testing
	gin.SetMode(gin.TestMode)

	// Create a mock config that should work without files
	cfg := &config.BaseConfig{} //nolint:exhaustruct // Test struct initialization

	// Create server manually for testing (bypassing config loading)
	server := &Server{
		config: cfg,
		router: gin.New(), // Use gin.New() for testing to avoid default middleware
	}

	// Test adding a GET route
	server.NewRoute("GET", "/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test the route works
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
}

func TestNewRouteWithDifferentMethods(t *testing.T) { //nolint:paralleltest // Simple test function
	// Set gin to test mode for testing
	gin.SetMode(gin.TestMode)

	cfg := &config.BaseConfig{} //nolint:exhaustruct // Test struct initialization
	server := &Server{
		config: cfg,
		router: gin.New(),
	}

	// Test different HTTP methods
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		endpoint := "/" + method
		server.NewRoute(method, endpoint, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"method": method})
		})

		// Test the route
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), method, endpoint, http.NoBody)
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Failed for method: %s", method)
	}
}
