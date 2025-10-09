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

type Authentication struct {
	BasicAuthenticator middleware.BasicAuthenticator
}

// GetConfig is the main function for consumers to load and get their configuration.
// It takes a pointer to any struct type that defines the configuration schema.
// The struct should have appropriate mapstructure and validate tags.
func Init[T config.HasBaseConfig](
	cfg T,
	serviceName, version string,
	policyAdapter persist.Adapter,
	authentication *Authentication,
) {
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

	if authentication != nil && policyAdapter != nil {
		Server.Use(middleware.Authentication(authentication.BasicAuthenticator))
		Server.Use(middleware.Authz(policyAdapter))
	}

	if (authentication != nil && policyAdapter == nil) ||
		(authentication == nil && policyAdapter != nil) {
		log.Fatal().Msg("Both authentication and policyAdapter must be provided together or neither")
	}

	log.Info().
		Int("port", cfg.GetBase().Port).
		Msg("Server initialized with middleware")
}

// @title			SharedDeps Server
// @version			v0.0.0
// @description	API Abstraction of common web-server duties.
// @license.name	GNU General Public License v3.0
// @license.url	https://www.gnu.org/licenses/gpl-3.0.html
func Start() {
	addr := fmt.Sprintf(":%d", config.Cfg.Port)
	if err := Server.Run(addr); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start server on %s", addr)
	}
}
