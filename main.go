package shareddeps

import (
	"fmt"

	"github.com/EnclaveRunner/shareddeps/config"
	"github.com/EnclaveRunner/shareddeps/middleware"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var Server *gin.Engine

// GetConfig is the main function for consumers to load and get their configuration.
// It takes a pointer to any struct type that defines the configuration schema.
// The struct should have appropriate mapstructure and validate tags.
func Init[T config.HasBaseConfig](cfg T, serviceName, version string, policiyAdapter persist.Adapter) {
	err := config.LoadAppConfig(cfg, serviceName, version)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	if config.Cfg.ProductionEnvironment {
		gin.SetMode(gin.ReleaseMode)
	}

	Server = gin.New()

	// Add recovery middleware
	Server.Use(gin.Recovery())

	// Add our custom zerolog middleware
	Server.Use(middleware.Zerolog())
	Server.Use(middleware.Authz(policiyAdapter))

	log.Info().
		Str("service", serviceName).
		Str("version", version).
		Msg("Server initialized with middleware")
}

// Start starts the gin server on the specified port.
// If no port is provided, it defaults to ":8080".
func Start() {
	addr := fmt.Sprintf(":%d", config.Cfg.Port)
	if err := Server.Run(addr); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start server on %s", addr)
	}
}
