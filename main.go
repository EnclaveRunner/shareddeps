package shareddeps

import (
	"fmt"

	"github.com/EnclaveRunner/shareddeps/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var Server *gin.Engine = gin.Default()

// GetConfig is the main function for consumers to load and get their configuration.
// It takes a pointer to any struct type that defines the configuration schema.
// The struct should have appropriate mapstructure and validate tags.
//
// Example usage:
//
//	shareddeps.Init(&MyConfig{}, "myservice", "1.0.0")
//	fmt.Printf("Port: %s\n", MyConfig.Port)
func Init[T config.HasBaseConfig](cfg T, serviceName, version string) {
	err := config.LoadAppConfig(cfg, serviceName, version)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}
}

// Start starts the gin server on the specified port.
// If no port is provided, it defaults to ":8080".
func Start() {
	addr := fmt.Sprintf(":%d", config.Cfg.Port)
	if err := Server.Run(addr); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start server on %s", addr)
	}
}
