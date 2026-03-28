package tests

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestAPISpoke performs a basic health check on the FlyFleet API
func TestAPISpoke(t *testing.T) {
	// Skip if not in k8s environment
	if os.Getenv("SKIP_API_TESTS") == "true" {
		t.Skip("Skipping API test (SKIP_API_TESTS set)")
	}

	apiURL := os.Getenv("FLYFLEET_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8084"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL+"/health", nil)
	require.NoError(t, err, "Failed to create health check request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Health check request failed")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200 OK")
	t.Logf("✅ API health check passed: %s/health", apiURL)
}

// TestKubernetesDeployment verifies that FlyFleet pods are running
func TestKubernetesDeployment(t *testing.T) {
	// Requires kubectl to be configured
	if os.Getenv("SKIP_K8S_TESTS") == "true" {
		t.Skip("Skipping Kubernetes test (SKIP_K8S_TESTS set)")
	}

	namespace := os.Getenv("FLYFLEET_NAMESPACE")
	if namespace == "" {
		namespace = "flyfleet"
	}

	// Use kubectl command to check pods (simple integration test)
	// For more sophisticated tests, use client-go
	t.Logf("Checking pods in namespace %s...", namespace)

	// This test assumes kubectl is available
	// In real CI/CD, you'd use Kubernetes client library
	t.Skip("This test requires kubectl; run manually: kubectl get pods -n flyfleet")
}

// TestHelmRelease verifies Helm chart is installed and values are set
func TestHelmRelease(t *testing.T) {
	if os.Getenv("SKIP_HELM_TESTS") == "true" {
		t.Skip("Skipping Helm test (SKIP_HELM_TESTS set)")
	}

	releaseName := os.Getenv("HELM_RELEASE_NAME")
	if releaseName == "" {
		releaseName = "flyfleet"
	}

	namespace := os.Getenv("FLYFLEET_NAMESPACE")
	if namespace == "" {
		namespace = "flyfleet"
	}

	t.Logf("Checking Helm release %s/%s", namespace, releaseName)

	// This would use Helm CLI or Go Helm library
	// Skipping in this basic version
	t.Skip("Helm release check requires Helm CLI; run manually: helm status flyfleet -n flyfleet")
}

// TestDatabaseConnectivity verifies database is accessible from application pod
func TestDatabaseConnectivity(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database test (SKIP_DB_TESTS set)")
	}

	// This test would execute SQL query from within a pod
	// Using kubectl exec as simplest approach
	t.Skip("Database connectivity test requires running pod; execute manually: kubectl exec -n flyfleet deploy/flyfleet-fleet-core-service -- psql -U fleet_user -d fleet_db -c 'SELECT 1'")
}

// TestExternalSecrets verifies that ExternalSecret resources are synced
func TestExternalSecrets(t *testing.T) {
	if os.Getenv("SKIP_SECRETS_TESTS") == "true" {
		t.Skip("Skipping secrets test (SKIP_SECRETS_TESTS set)")
	}

	namespace := os.Getenv("FLYFLEET_NAMESPACE")
	if namespace == "" {
		namespace = "flyfleet"
	}

	t.Logf("Checking ExternalSecret status in %s", namespace)

	// Expect flyfleet-secrets to exist and be Ready
	// kubectl get externalsecret flyfleet-secrets -n flyfleet -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
	t.Skip("Secrets verification requires External Secrets Operator; run: kubectl get externalsecret -n flyfleet")
}

// Run all tests:
// go test ./tests/infrastructure -v
//
// With Docker:
// docker run --rm -it -v $(pwd):/workspace -w /workspace golang:1.21 go test ./tests/infrastructure -v
//
// Or via Makefile:
// make test:infra
