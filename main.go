package shareddeps

import (
	"fmt"

	"github.com/EnclaveRunner/shareddeps/config"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config any
	router *gin.Engine
}

// GetConfig is the main function for consumers to load and get their configuration.
// It takes a pointer to any struct type that defines the configuration schema.
// The struct should have appropriate mapstructure and validate tags.
//
// Example usage:
//
//	cfg, err := shareddeps.GetConfig(&MyConfig{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Port: %s\n", cfg.Port)
func InitConfig[T any](cfg *T) (*Server, error) {
	if err := config.LoadAppConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	router := gin.Default()

	server := &Server{
		config: cfg,
		router: router,
	}

	return server, nil
}

// NewRoute registers a new route with the gin server.
// It takes an HTTP method, endpoint path, and a gin.HandlerFunc callback.
//
// Example usage:
//
//	server.NewRoute("GET", "/health", func(c *gin.Context) {
//	    c.JSON(200, gin.H{"status": "ok"})
//	})
func (s *Server) NewRoute(method, endpoint string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		s.router.GET(endpoint, handler)
	case "POST":
		s.router.POST(endpoint, handler)
	case "PUT":
		s.router.PUT(endpoint, handler)
	case "DELETE":
		s.router.DELETE(endpoint, handler)
	case "PATCH":
		s.router.PATCH(endpoint, handler)
	default:
		// Default to GET if method is not recognized
		s.router.GET(endpoint, handler)
	}
}

// Start starts the gin server on the specified port.
// If no port is provided, it defaults to ":8080".
func (s *Server) Start(port ...string) error {
	addr := ":8080"
	if len(port) > 0 && port[0] != "" {
		addr = port[0]
	}

	if err := s.router.Run(addr); err != nil {
		return fmt.Errorf("failed to start server on %s: %w", addr, err)
	}

	return nil
}
