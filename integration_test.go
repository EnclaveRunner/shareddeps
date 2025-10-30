package shareddeps_test

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/EnclaveRunner/shareddeps"
	"github.com/EnclaveRunner/shareddeps/client"
	"github.com/EnclaveRunner/shareddeps/config"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/stretchr/testify/assert"
)

func startServer(t *testing.T, port int) {
	defaults := []config.DefaultValue{
		{Key: "port", Value: port},
		{Key: "production_environment", Value: false},
		{Key: "log_level", Value: "debug"},
		{Key: "human_readable_output", Value: true},
	}

	tmpDir := t.TempDir()
	err := os.WriteFile(tmpDir+"/policies.csv", []byte(""), 0o644)
	assert.NoError(t, err)

	shareddeps.Init(config.Cfg, "test-service", "v0.6.0", defaults...)
	shareddeps.AddAuth(
		fileadapter.NewAdapter(tmpDir+"/policies.csv"),
		shareddeps.Authentication{
			BasicAuthenticator: func(ctx context.Context, username, password string) (string, error) {
				return "test-user-id", nil
			},
		},
	)
	go shareddeps.Start()
	time.Sleep(3 * time.Second)
}

func TestHealthCheck(t *testing.T) {
	t.Parallel()
	//nolint:gosec // Random port for testing
	port := rand.Intn(10000) + 8000
	startServer(t, port)
	c, _ := client.NewClientWithResponses(
		"http://localhost:" + strconv.Itoa(port),
	)
	resp, err := c.GetHealthWithResponse(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())
}
