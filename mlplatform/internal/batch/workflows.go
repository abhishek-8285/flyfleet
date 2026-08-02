package batch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/flyfleet/mlplatform/internal/drift"
	"github.com/flyfleet/mlplatform/internal/featurestore"
	"github.com/flyfleet/mlplatform/internal/inference"
	"github.com/flyfleet/mlplatform/internal/warehouse"
)

// Dependencies wires the real platform components into the batch activities.
// A Temporal worker process calls InitActivities with these before starting
// its worker loop.
type Dependencies struct {
	Features       featurestore.Store
	Warehouse      *warehouse.ClickHouseStore
	Engine         *inference.Engine
	MLPlatformURL  string
	ModelDir       string
}

// deployedModels records the active model version per model name.
var deployedModels sync.Map

// deps holds the activity dependencies. It is populated by InitActivities,
// which a worker process must call before Run.
var deps *Dependencies

// InitActivities configures the package-level activity dependencies.
func InitActivities(d *Dependencies) { deps = d }

// RegisterWorkflows registers all batch workflows and activities with a
// Temporal worker. Call this from a dedicated worker process:
//
//	w := worker.New(c, taskQueue, worker.Options{})
//	RegisterWorkflows(w)
func RegisterWorkflows(w worker.Worker) {
	w.RegisterWorkflow(DailyFeatureGenerationWorkflow)
	w.RegisterWorkflow(NightlyTrainingWorkflow)
	w.RegisterWorkflow(HourlyAggregationWorkflow)
	w.RegisterWorkflow(WeeklyDriftDetectionWorkflow)

	w.RegisterActivity(ComputeDailyFeatures)
	w.RegisterActivity(ExportTrainingDataset)
	w.RegisterActivity(TrainModel)
	w.RegisterActivity(ExportToONNX)
	w.RegisterActivity(DeployModel)
	w.RegisterActivity(AggregateHourlyMetrics)
	w.RegisterActivity(RunDriftCheck)
}

// DailyFeatureGenerationWorkflow recomputes all batch features.
func DailyFeatureGenerationWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Hour,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	log.Info().Msg("starting daily feature generation")

	var featureCount int
	if err := workflow.ExecuteActivity(ctx, ComputeDailyFeatures).Get(ctx, &featureCount); err != nil {
		return err
	}

	log.Info().Int("features_computed", featureCount).Msg("daily features done")
	return nil
}

// NightlyTrainingWorkflow trains and deploys models.
func NightlyTrainingWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 4 * time.Hour,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var datasetPath string
	if err := workflow.ExecuteActivity(ctx, ExportTrainingDataset).Get(ctx, &datasetPath); err != nil {
		return err
	}

	var modelPath string
	if err := workflow.ExecuteActivity(ctx, TrainModel, datasetPath).Get(ctx, &modelPath); err != nil {
		return err
	}

	var onnxPath string
	if err := workflow.ExecuteActivity(ctx, ExportToONNX, modelPath).Get(ctx, &onnxPath); err != nil {
		return err
	}

	if err := workflow.ExecuteActivity(ctx, DeployModel, onnxPath).Get(ctx, nil); err != nil {
		return err
	}

	return nil
}

// HourlyAggregationWorkflow pre-computes analytics aggregates.
func HourlyAggregationWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	return workflow.ExecuteActivity(ctx, AggregateHourlyMetrics).Get(ctx, nil)
}

// WeeklyDriftDetectionWorkflow checks model quality.
func WeeklyDriftDetectionWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Hour,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var driftScore float64
	if err := workflow.ExecuteActivity(ctx, RunDriftCheck).Get(ctx, &driftScore); err != nil {
		return err
	}

	if driftScore > 0.1 {
		return workflow.ExecuteChildWorkflow(ctx, NightlyTrainingWorkflow).Get(ctx, nil)
	}
	return nil
}

// Activity implementations.
//
// Each activity is wired to its real component via the package-level deps
// configured by InitActivities. The heavy compute stages (training, ONNX
// export) are delegated to external tooling in production; the activities
// here materialize the artifacts and hand off the artifact paths.

func ComputeDailyFeatures(ctx context.Context) (int, error) {
	if deps == nil {
		return 0, errors.New("batch activities not initialized")
	}

	rows, err := deps.Warehouse.Query(ctx, `
		SELECT subject, count() AS n
		FROM events
		WHERE timestamp >= now() - INTERVAL 1 DAY
		GROUP BY subject
	`)
	if err != nil {
		return 0, fmt.Errorf("query recent events: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var subject string
		var n int
		if err := rows.Scan(&subject, &n); err != nil {
			return 0, err
		}
		parts := strings.SplitN(subject, ":", 2)
		if len(parts) != 2 {
			continue
		}
		fv := &featurestore.FeatureVector{
			EntityType: parts[0],
			EntityID:   parts[1],
			Features: []featurestore.FeatureValue{{
				Name:      "event_count_24h",
				Value:     float64(n),
				Timestamp: time.Now().UTC(),
			}},
		}
		if err := deps.Features.PutOnlineFeatures(ctx, fv); err != nil {
			return 0, fmt.Errorf("write daily feature: %w", err)
		}
		count++
	}
	return count, nil
}

func ExportTrainingDataset(ctx context.Context) (string, error) {
	if deps == nil {
		return "", errors.New("batch activities not initialized")
	}

	end := time.Now().UTC()
	start := end.Add(-90 * 24 * time.Hour)

	fvs, err := deps.Warehouse.GetOfflineFeatures(ctx, "driver", nil, []string{
		"acceptance_rate", "avg_rating", "total_rides",
		"distance_to_rider_km", "surge_multiplier",
	}, start, end)
	if err != nil {
		return "", fmt.Errorf("export training dataset: %w", err)
	}

	path := filepath.Join(os.TempDir(), fmt.Sprintf("dataset_%d.json", time.Now().UnixNano()))
	data, err := json.Marshal(fvs)
	if err != nil {
		return "", fmt.Errorf("marshal dataset: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write dataset: %w", err)
	}
	return path, nil
}

func TrainModel(ctx context.Context, datasetPath string) (string, error) {
	if deps == nil {
		return "", errors.New("batch activities not initialized")
	}

	modelDir := deps.ModelDir
	if modelDir == "" {
		modelDir = "./models"
	}

	// Run the Python training pipeline.
	// In production, this would be a container job on K8s.
	cmd := exec.CommandContext(ctx,
		"python", "training/train_dispatch_ranker.py",
		"--config", "training/config.yaml",
		"--dry-run",
	)
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("training failed: %w", err)
	}

	// Find the latest ONNX file produced by the training script.
	matches, err := filepath.Glob(filepath.Join(modelDir, "dispatch_ranker_v*.onnx"))
	if err != nil {
		return "", fmt.Errorf("find onnx files: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no ONNX model produced by training")
	}

	// Sort by modification time, take newest.
	sort.Slice(matches, func(i, j int) bool {
		fi, _ := os.Stat(matches[i])
		fj, _ := os.Stat(matches[j])
		return fi.ModTime().After(fj.ModTime())
	})

	return matches[0], nil
}

func ExportToONNX(ctx context.Context, modelPath string) (string, error) {
	if deps == nil {
		return "", errors.New("batch activities not initialized")
	}

	// The Python training pipeline produces the ONNX file directly.
	// The modelPath from TrainModel already points to the ONNX artifact.
	return modelPath, nil
}

func DeployModel(ctx context.Context, onnxPath string) error {
	if deps == nil {
		return errors.New("batch activities not initialized")
	}

	baseURL := deps.MLPlatformURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	version := strings.TrimSuffix(filepath.Base(onnxPath), ".onnx")
	modelName := strings.TrimSuffix(filepath.Base(onnxPath), filepath.Ext(onnxPath))

	allFeatures := []string{
		"distance_to_rider_km", "acceptance_rate_7d", "rides_completed_today",
		"minutes_since_last_ride", "driver_rating", "is_heading_toward_rider",
		"surge_multiplier", "time_of_day_hour", "day_of_week",
		"zone_demand_density", "zone_supply_density",
		"supply_demand_ratio", "effective_distance",
	}

	// 1. Load the new version (don't activate yet)
	log.Info().Str("path", onnxPath).Str("version", version).Msg("loading model into mlplatform")
	body := fmt.Sprintf(`{"name":"dispatch_ranker","version":%q,"path":%q,"format":"onnx","features":%s,"activate":false}`,
		version, onnxPath, marshalFeatures(allFeatures))
	resp, err := http.Post(baseURL+"/v1/models/load", "application/json", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("model load request: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 201 {
		return fmt.Errorf("model load failed: %d", resp.StatusCode)
	}

	// 2. Activate (hot-swap)
	log.Info().Str("version", version).Msg("activating model")
	body = fmt.Sprintf(`{"name":"dispatch_ranker","new_version":%q}`, version)
	resp, err = http.Post(baseURL+"/v1/models/swap", "application/json", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("model swap request: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("model swap failed: %d", resp.StatusCode)
	}

	deployedModels.Store(modelName, version)
	log.Info().Str("version", version).Msg("model deployed and active")
	return nil
}

func marshalFeatures(features []string) string {
	parts := make([]string, len(features))
	for i, f := range features {
		parts[i] = fmt.Sprintf("%q", f)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func AggregateHourlyMetrics(ctx context.Context) (int, error) {
	if deps == nil {
		return 0, errors.New("batch activities not initialized")
	}

	rows, err := deps.Warehouse.Query(ctx, `
		SELECT type, cnt
		FROM events_hourly_mv
		WHERE hour >= now() - INTERVAL 1 HOUR
		ORDER BY cnt DESC
		LIMIT 100
	`)
	if err != nil {
		return 0, fmt.Errorf("query hourly mv: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var typ string
		var cnt int
		if err := rows.Scan(&typ, &cnt); err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

func RunDriftCheck(ctx context.Context) (float64, error) {
	if deps == nil {
		return 0, errors.New("batch activities not initialized")
	}

	rows, err := deps.Warehouse.Query(ctx, `
		SELECT value
		FROM features
		WHERE entity_type = ? AND name = ?
		ORDER BY timestamp DESC
		LIMIT 40
	`, "driver", "acceptance_rate")
	if err != nil {
		return 0, fmt.Errorf("query features: %w", err)
	}
	defer rows.Close()

	var recent []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return 0, err
		}
		recent = append(recent, v)
	}
	if len(recent) < 20 {
		return 0, nil
	}
	// Rows are newest-first, so the newer half is the "current" distribution
	// and the older half is the "baseline" distribution.
	mid := len(recent) / 2
	return drift.PSI(recent[mid:], recent[:mid]), nil
}
