package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/dr/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeDrService struct {
	create            func(ctx context.Context, test *models.DRTest) (*models.DRTest, error)
	abort             func(ctx context.Context, id uuid.UUID) error
	getDetail         func(ctx context.Context, id string) (*models.DRTest, []*models.DRTestStep, []*models.DRTestLog, error)
	getMetrics        func(ctx context.Context) (*models.DRTestMetrics, error)
	getSettings       func(ctx context.Context) (*models.DRTestSettings, error)
	updateSettings    func(ctx context.Context, settings *models.DRTestSettings) error
	listTests         func(ctx context.Context, limit, offset int) ([]*models.DRTest, int, error)
	placeholderCreate func(ctx context.Context, backupType models.BackupType, createdBy string) (*models.Backup, error)
}

func (f *fakeDrService) InitiateBackup(ctx context.Context, backupType models.BackupType, createdBy uuid.UUID) (*models.Backup, error) {
	return nil, nil
}
func (f *fakeDrService) GetBackup(ctx context.Context, id uuid.UUID) (*models.Backup, error) {
	return nil, nil
}
func (f *fakeDrService) ListBackups(ctx context.Context, limit, offset int) ([]*models.Backup, int, error) {
	return nil, 0, nil
}
func (f *fakeDrService) VerifyBackup(ctx context.Context, backupID uuid.UUID, verifiedBy uuid.UUID) (*models.BackupVerification, error) {
	return nil, nil
}
func (f *fakeDrService) InitiateRestore(ctx context.Context, backupID uuid.UUID, targetEnv string, performedBy uuid.UUID) (*models.RestoreJob, error) {
	return nil, nil
}
func (f *fakeDrService) GetRestoreJob(ctx context.Context, id uuid.UUID) (*models.RestoreJob, error) {
	return nil, nil
}
func (f *fakeDrService) ListRestoreJobs(ctx context.Context, limit, offset int) ([]*models.RestoreJob, int, error) {
	return nil, 0, nil
}
func (f *fakeDrService) ListStoragePools(ctx context.Context) ([]*models.StoragePool, error) {
	return nil, nil
}

func (f *fakeDrService) CreateDRTest(ctx context.Context, test *models.DRTest) (*models.DRTest, error) {
	if f.create != nil {
		return f.create(ctx, test)
	}
	return nil, nil
}
func (f *fakeDrService) ListDRTests(ctx context.Context, limit, offset int) ([]*models.DRTest, int, error) {
	if f.listTests != nil {
		return f.listTests(ctx, limit, offset)
	}
	return nil, 0, nil
}
func (f *fakeDrService) GetDRTestDetails(ctx context.Context, id uuid.UUID) (*models.DRTest, []*models.DRTestStep, []*models.DRTestLog, error) {
	if f.getDetail != nil {
		return f.getDetail(ctx, id.String())
	}
	return nil, nil, nil, nil
}
func (f *fakeDrService) AbortDRTest(ctx context.Context, id uuid.UUID) error {
	if f.abort != nil {
		return f.abort(ctx, id)
	}
	return nil
}
func (f *fakeDrService) GetDRTestMetrics(ctx context.Context) (*models.DRTestMetrics, error) {
	if f.getMetrics != nil {
		return f.getMetrics(ctx)
	}
	return nil, nil
}
func (f *fakeDrService) GetDRTestSettings(ctx context.Context) (*models.DRTestSettings, error) {
	if f.getSettings != nil {
		return f.getSettings(ctx)
	}
	return nil, nil
}
func (f *fakeDrService) UpdateDRTestSettings(ctx context.Context, settings *models.DRTestSettings) error {
	if f.updateSettings != nil {
		return f.updateSettings(ctx, settings)
	}
	return nil
}

func TestCreateDRTestHandler(t *testing.T) {
	sent := map[string]any{}
	svc := &fakeDrService{
		create: func(ctx context.Context, test *models.DRTest) (*models.DRTest, error) {
			sent["name"] = test.Name
			return &models.DRTest{ID: uuid.New(), Status: "RUNNING"}, nil
		},
	}
	h := NewDrHandlers(svc, zap.NewNop())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/dr/tests" && r.Method == http.MethodPost {
			h.CreateDRTest(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	payload := map[string]any{"name": "test", "test_type": "FAILOVER", "target_env": "STAGING"}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(server.URL+"/api/v1/dr/tests", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "test", sent["name"])
}

func TestAbortDRTestHandler(t *testing.T) {
	called := ""
	svc := &fakeDrService{
		abort: func(ctx context.Context, id uuid.UUID) error {
			called = id.String()
			return nil
		},
	}
	h := NewDrHandlers(svc, zap.NewNop())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/dr/tests/abc/abort" && r.Method == http.MethodPost {
			h.AbortDRTest(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/v1/dr/tests/abc/abort", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "abc", called)
}
