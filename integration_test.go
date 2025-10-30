package shareddeps_test

import (
	"context"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/EnclaveRunner/shareddeps"
	"github.com/EnclaveRunner/shareddeps/client"
	"github.com/EnclaveRunner/shareddeps/config"
	pb "github.com/EnclaveRunner/shareddeps/proto_gen"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var serverInitMu sync.Mutex

func startRESTServer(t *testing.T, port int) {
	tmpDir := t.TempDir()

	err := os.WriteFile(tmpDir+"/policies.csv", []byte(""), 0o644)
	assert.NoError(t, err)
	// Serialize initialization to avoid races on global config.Cfg
	serverInitMu.Lock()
	defer serverInitMu.Unlock()

	// Apply per-test viper settings while holding the lock
	viper.Set("port", port)
	viper.Set("production_environment", false)
	viper.Set("log_level", "debug")
	viper.Set("human_readable_output", true)

	// Create a new config instance for this test
	cfg := &config.BaseConfig{Port: port}
	shareddeps.InitRESTServer(cfg, "test-REST-service", "v0.6.0")
	shareddeps.AddAuth(
		fileadapter.NewAdapter(tmpDir+"/policies.csv"),
		shareddeps.Authentication{
			BasicAuthenticator: func(ctx context.Context, username, password string) (string, error) {
				return "test-user-id", nil
			},
		},
	)
	go shareddeps.StartRESTServer()
	time.Sleep(3 * time.Second)
}

func TestRESTHealthCheck(t *testing.T) {
	t.Parallel()
	port := 8901
	startRESTServer(t, port)
	c, _ := client.NewClientWithResponses(
		"http://localhost:" + strconv.Itoa(port),
	)
	resp, err := c.GetHealthWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())
}

func startGRPCServer(t *testing.T, port int) {
	tmpDir := t.TempDir()
	err := os.WriteFile(tmpDir+"/policies.csv", []byte(""), 0o644)
	assert.NoError(t, err)
	// Serialize initialization to avoid races on global config.Cfg
	serverInitMu.Lock()
	defer serverInitMu.Unlock()

	// Apply per-test viper settings while holding the lock
	viper.Set("port", port)
	viper.Set("log_level", "debug")
	viper.Set("human_readable_output", true)

	// Create a new config instance for this test
	cfg := &config.BaseConfig{Port: port}
	shareddeps.InitGRPCServer(cfg, "test-GRPC-service", "v0.6.0")

	pb.RegisterHealthServiceServer(
		shareddeps.GRPCServer,
		&server{},
	)
	go shareddeps.StartGRPCServer()
	time.Sleep(3 * time.Second)
}

type server struct {
	pb.HealthServiceServer
}

func (s *server) CheckHealth(
	ctx context.Context,
	in *pb.HealthCheckRequest,
) (*pb.HealthCheckResponse, error) {
	// You can add more detailed health check logic here
	return &pb.HealthCheckResponse{Status: "SERVING"}, nil
}

func TestGRPCHealthCheck(t *testing.T) {
	t.Parallel()
	port := 8902
	startGRPCServer(t, port)

	conn, err := grpc.NewClient(
		"localhost:"+strconv.Itoa(port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)
	defer func() {
		err := conn.Close()
		assert.NoError(t, err)
	}()

	healthClient := pb.NewHealthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := healthClient.CheckHealth(ctx, &pb.HealthCheckRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "SERVING", resp.Status)
}
