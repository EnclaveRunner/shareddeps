package shareddeps

import (
	"context"
	"fmt"
	"net"

	"github.com/EnclaveRunner/shareddeps/api"
	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/EnclaveRunner/shareddeps/config"
	"github.com/EnclaveRunner/shareddeps/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Authentication struct {
	BasicAuthenticator middleware.BasicAuthenticator
}

func PopulateAppConfig[T config.HasBaseConfig](
	cfg T,
	serviceName, version string, defaultValues ...config.DefaultValue,
) {
	err := config.PopulateAppConfig(cfg, serviceName, version, defaultValues...)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to populate application config")
	}
}

func InitRESTServer(cfg config.HasBaseConfig) *gin.Engine {
	if cfg.GetBase().ProductionEnvironment {
		gin.SetMode(gin.ReleaseMode)
	}

	restServer := gin.New()

	// Add recovery middleware
	restServer.Use(gin.Recovery())

	// Add our custom zerolog middleware
	restServer.Use(middleware.Zerolog())

	server := api.NewServer()
	handler := api.NewStrictHandler(server, nil)
	api.RegisterHandlers(restServer, handler)

	log.Info().Msg("Server initialized with middleware")

	return restServer
}

func InitGRPCServer() *grpc.Server {
	// create the gRPC server
	grpcServer := grpc.NewServer()

	log.Info().Msg("gRPC server initialized")

	return grpcServer
}

func StartGRPCServer(cfg config.HasBaseConfig, server *grpc.Server) {
	lc := net.ListenConfig{}
	lis, err := lc.Listen(
		context.Background(),
		"tcp",
		fmt.Sprintf(":%d", cfg.GetBase().Port),
	)
	if err != nil {
		log.Fatal().
			Err(err).
			Int("port", cfg.GetBase().Port).
			Msg("Failed to create gRPC listener")
	}

	log.Info().
		Int("port", cfg.GetBase().Port).
		Msg("Setup finished. Starting to listen")
	addr := fmt.Sprintf(":%d", cfg.GetBase().Port)
	if err := server.Serve(lis); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start gRPC server on %s", addr)
	}
}

func StartRESTServer(cfg config.HasBaseConfig, server *gin.Engine) {
	log.Info().
		Int("port", cfg.GetBase().Port).
		Msg("Setup finished. Starting to listen")
	addr := fmt.Sprintf(":%d", cfg.GetBase().Port)
	if err := server.Run(addr); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start server on %s", addr)
	}
}

// InitGRPCClient initializes a gRPC client connection to the specified host and
// port.
func InitGRPCClient(host string, port int) *grpc.ClientConn {
	client, err := grpc.NewClient(
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

	return client
}

// AddAuth adds authentication and authorization middleware to the REST-Server.
// Must be called after InitRESTServer and before StartRESTServer.
func AddAuth(
	server *gin.Engine,
	authModule auth.AuthModule,
	authentication Authentication,
) {
	server.Use(middleware.Authentication(authentication.BasicAuthenticator))
	server.Use(authModule.Middleware())

	// Add policy to allow health checks without authentication
	err := authModule.CreateResourceGroup("health_INTERNAL")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create health_INTERNAL resource group")
	}
	err = authModule.AddResourceToGroup("/health", "health_INTERNAL")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to add /health to health_INTERNAL resource group")
	}
	err = authModule.AddPolicy("*", "health_INTERNAL", "GET")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to add policy for health_INTERNAL resource group")
	}

	log.Info().Msg("Authentication and Authorization middleware added")
}
