package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/fleet-platform/shared/pkg/security"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
)

type MonitoringHandler struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

type SystemMetrics struct {
	Timestamp    time.Time `json:"timestamp"`
	CPUUsage     float64   `json:"cpu_usage_percent"`
	MemoryUsage  float64   `json:"memory_usage_percent"`
	MemoryUsed   uint64    `json:"memory_used_bytes"`
	MemoryTotal  uint64    `json:"memory_total_bytes"`
	Goroutines   int       `json:"goroutines"`
}

type DatabaseMetrics struct {
	Connections      int     `json:"active_connections"`
	MaxConnections   int     `json:"max_connections"`
	ConnectionsUsed  float64 `json:"connections_used_percent"`
	CacheHitRatio    float64 `json:"cache_hit_ratio"`
	Deadlocks        int64   `json:"deadlocks"`
	TempFiles        int64   `json:"temp_files"`
	SlowQueries      int64   `json:"slow_queries"`
}

type CombinedMetrics struct {
	System   SystemMetrics   `json:"system"`
	Database DatabaseMetrics `json:"database"`
}

func NewMonitoringHandler(db *pgxpool.Pool, logger *zap.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		db:     db,
		logger: logger,
	}
}

func (h *MonitoringHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := h.collectSystemMetrics()
	if err != nil {
		h.logger.Error("Failed to collect system metrics", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *MonitoringHandler) GetDatabaseMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := h.collectDatabaseMetrics()
	if err != nil {
		h.logger.Error("Failed to collect database metrics", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *MonitoringHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	systemMetrics, err := h.collectSystemMetrics()
	if err != nil {
		h.logger.Error("Failed to collect system metrics", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	databaseMetrics, err := h.collectDatabaseMetrics()
	if err != nil {
		h.logger.Error("Failed to collect database metrics", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	combined := CombinedMetrics{
		System:   *systemMetrics,
		Database: *databaseMetrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(combined)
}

func (h *MonitoringHandler) collectSystemMetrics() (*SystemMetrics, error) {
	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Memory usage
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	metrics := &SystemMetrics{
		Timestamp:   time.Now(),
		CPUUsage:    cpuPercent[0],
		MemoryUsage: memStats.UsedPercent,
		MemoryUsed:  memStats.Used,
		MemoryTotal: memStats.Total,
		Goroutines:  runtime.NumGoroutine(),
	}

	return metrics, nil
}

func (h *MonitoringHandler) collectDatabaseMetrics() (*DatabaseMetrics, error) {
	metrics := &DatabaseMetrics{}
	
	// Create context for database queries
	ctx := context.Background()

	// Active connections
	var activeConnections int
	err := h.db.QueryRow(ctx, `
		SELECT count(*) FROM pg_stat_activity 
		WHERE datname = current_database() AND state = 'active'
	`).Scan(&activeConnections)
	if err != nil {
		h.logger.Warn("Failed to get active connections", zap.Error(err))
	}

	// Max connections
	var maxConnections int
	err = h.db.QueryRow(ctx, `
		SELECT setting::int FROM pg_settings WHERE name = 'max_connections'
	`).Scan(&maxConnections)
	if err != nil {
		h.logger.Warn("Failed to get max connections", zap.Error(err))
	}

	// Calculate connection usage percentage
	if maxConnections > 0 {
		metrics.ConnectionsUsed = float64(activeConnections) / float64(maxConnections) * 100
	}

	metrics.Connections = activeConnections
	metrics.MaxConnections = maxConnections

	// Cache hit ratio
	var cacheHitRatio float64
	err = h.db.QueryRow(ctx, `
		SELECT round(sum(blks_hit)*100/sum(blks_hit+blks_read), 2) 
		FROM pg_stat_database 
		WHERE datname = current_database()
	`).Scan(&cacheHitRatio)
	if err != nil {
		h.logger.Warn("Failed to get cache hit ratio", zap.Error(err))
	}
	metrics.CacheHitRatio = cacheHitRatio

	// Deadlocks
	var deadlocks int64
	err = h.db.QueryRow(ctx, `
		SELECT deadlocks FROM pg_stat_database 
		WHERE datname = current_database()
	`).Scan(&deadlocks)
	if err != nil {
		h.logger.Warn("Failed to get deadlock count", zap.Error(err))
	}
	metrics.Deadlocks = deadlocks

	// Temporary files
	var tempFiles int64
	err = h.db.QueryRow(ctx, `
		SELECT temp_files FROM pg_stat_database 
		WHERE datname = current_database()
	`).Scan(&tempFiles)
	if err != nil {
		h.logger.Warn("Failed to get temp files count", zap.Error(err))
	}
	metrics.TempFiles = tempFiles

	// Slow queries (queries taking more than 1 second)
	var slowQueries int64
	err = h.db.QueryRow(ctx, `
		SELECT count(*) FROM pg_stat_statements 
		WHERE mean_time > 1000 AND dbid = (SELECT oid FROM pg_database WHERE datname = current_database())
	`).Scan(&slowQueries)
	if err != nil {
		h.logger.Warn("Failed to get slow queries count", zap.Error(err))
	}
	metrics.SlowQueries = slowQueries

	return metrics, nil
}

func (h *MonitoringHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.db.Ping(ctx)
	if err != nil {
		h.logger.Error("Database health check failed", zap.Error(err))
		http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}

	response := map[string]interface{}{
		"status":                 "healthy",
		"timestamp":              time.Now(),
		"service":                "auth-service",
		"security_policy_hash":   security.GetPolicyVersion(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}