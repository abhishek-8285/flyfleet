package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupRedisContainer starts a Redis container for testing
func SetupRedisContainer(ctx context.Context, t *testing.T) (string, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	ip, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get redis container ip: %v", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get redis container port: %v", err)
	}

	addr := fmt.Sprintf("%s:%s", ip, port.Port())
	return addr, func() {
		container.Terminate(ctx)
	}
}

// SetupMinioContainer starts a MinIO container for testing
func SetupMinioContainer(ctx context.Context, t *testing.T) (string, string, string, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:latest",
		Cmd:          []string{"server", "/data"},
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     "admin",
			"MINIO_ROOT_PASSWORD": "password",
		},
		WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000/tcp"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start minio container: %v", err)
	}

	ip, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get minio container ip: %v", err)
	}

	port, err := container.MappedPort(ctx, "9000")
	if err != nil {
		t.Fatalf("failed to get minio container port: %v", err)
	}

	endpoint := fmt.Sprintf("%s:%s", ip, port.Port())
	return endpoint, "admin", "password", func() {
		container.Terminate(ctx)
	}
}
