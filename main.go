package shareddeps

import (
	"fmt"

	"github.com/EnclaveRunner/shareddeps/api"
	"github.com/EnclaveRunner/shareddeps/auth"
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

// GetConfig is the main function for consumers to load and get their
// configuration.
// It takes a pointer to any struct type that defines the configuration schema.
// The struct should have appropriate mapstructure and validate tags.
// Must be called before Init.
func Init[T config.HasBaseConfig](
	cfg T,
	serviceName, version string,
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

	server := api.NewServer()
	handler := api.NewStrictHandler(server, nil)
	api.RegisterHandlers(Server, handler)

	log.Info().Msg("Server initialized with middleware")
}

// AddAuth adds authentication and authorization middleware to the server.
// Must be called after Init and before Start.
func AddAuth(
	policyAdapter persist.Adapter,
	authentication Authentication,
) {
	enforcer := auth.InitAuth(policyAdapter)
	Server.Use(middleware.Authentication(authentication.BasicAuthenticator))
	Server.Use(middleware.Authz(enforcer))

	// Add policy to allow health checks without authentication
	err := auth.CreateResourceGroup("health_INTERNAL")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create health_INTERNAL resource group")
	}
	err = auth.AddResourceToGroup("/health", "health_INTERNAL")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to add /health to health_INTERNAL resource group")
	}
	err = auth.AddPolicy("*", "health_INTERNAL", "GET")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to add policy for health_INTERNAL resource group")
	}

	log.Info().Msg("Authentication and Authorization middleware added")
}

func Start() {
	log.Info().
		Int("port", config.Cfg.Port).
		Msg("Setup finished. Starting to listen")
	addr := fmt.Sprintf(":%d", config.Cfg.Port)
	if err := Server.Run(addr); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start server on %s", addr)
	}
}
