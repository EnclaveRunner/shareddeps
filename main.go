package shareddeps

import (
	"context"
	"fmt"
	"net"

	"github.com/EnclaveRunner/shareddeps/api"
	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/EnclaveRunner/shareddeps/config"
	"github.com/EnclaveRunner/shareddeps/middleware"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var RESTServer *gin.Engine

var GRPCServer *grpc.Server

var GRPCClient *grpc.ClientConn

type Authentication struct {
	BasicAuthenticator middleware.BasicAuthenticator
}

// GetConfig is the main function for consumers to load and get their
// configuration.
// It takes a pointer to any struct type that defines the configuration schema.
// The struct should have appropriate mapstructure and validate tags.
// Must be called before Init.
func InitRESTServer[T config.HasBaseConfig](
	cfg T,
	serviceName, version string, defaultValues ...config.DefaultValue,
) {
	err := config.LoadAppConfig(cfg, serviceName, version, defaultValues...)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	if config.Cfg.ProductionEnvironment {
		gin.SetMode(gin.ReleaseMode)
	}

	RESTServer = gin.New()

	// Add recovery middleware
	RESTServer.Use(gin.Recovery())

	// Add our custom zerolog middleware
	RESTServer.Use(middleware.Zerolog())

	server := api.NewServer()
	handler := api.NewStrictHandler(server, nil)
	api.RegisterHandlers(RESTServer, handler)

	log.Info().Msg("Server initialized with middleware")
}

func InitGRPCServer[T config.HasBaseConfig](
	cfg T,
	serviceName, version string, defaultValues ...config.DefaultValue,
) {
	err := config.LoadAppConfig(cfg, serviceName, version, defaultValues...)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// create the gRPC server and assign to the package-level variable
	GRPCServer = grpc.NewServer()

	log.Info().Msg("gRPC server initialized")
}

func StartGRPCServer() {
	lc := net.ListenConfig{}
	lis, err := lc.Listen(
		context.Background(),
		"tcp",
		fmt.Sprintf(":%d", config.Cfg.Port),
	)
	if err != nil {
		log.Fatal().
			Err(err).
			Int("port", config.Cfg.Port).
			Msg("Failed to create gRPC listener")
	}

	log.Info().
		Int("port", config.Cfg.Port).
		Msg("Setup finished. Starting to listen")
	addr := fmt.Sprintf(":%d", config.Cfg.Port)
	if err := GRPCServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start gRPC server on %s", addr)
	}
}

func StartRESTServer() {
	log.Info().
		Int("port", config.Cfg.Port).
		Msg("Setup finished. Starting to listen")
	addr := fmt.Sprintf(":%d", config.Cfg.Port)
	if err := RESTServer.Run(addr); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start server on %s", addr)
	}
}

// InitGRPCClient initializes a gRPC client connection to the specified host and
// port.
func InitGRPCClient(host string, port int) {
	var err error
	GRPCClient, err = grpc.NewClient(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create gRPC client")
	}
	log.Info().
		Int("port", port).
		Str("host", host).
		Msg("gRPC client initialized successfully")
}

// AddAuth adds authentication and authorization middleware to the REST-Server.
// Must be called after InitRESTServer and before StartRESTServer.
func AddAuth(
	policyAdapter persist.Adapter,
	authentication Authentication,
) {
	enforcer := auth.InitAuth(policyAdapter)
	RESTServer.Use(middleware.Authentication(authentication.BasicAuthenticator))
	RESTServer.Use(middleware.Authz(enforcer))

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
