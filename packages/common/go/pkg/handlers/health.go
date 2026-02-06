package handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status      string                 `json:"status"`
	Version     string                 `json:"version"`
	Service     string                 `json:"service"`
	SystemStats SystemStats            `json:"system_stats"`
	Components  map[string]interface{} `json:"components,omitempty"`
	Timestamp   string                 `json:"timestamp"`
}

// SystemStats represents system statistics
type SystemStats struct {
	Goroutines int    `json:"goroutines"`
	CPUThreads int    `json:"cpu_threads"`
	MemoryMB   uint64 `json:"memory_mb"`
}

// HealthHandler handles health check requests
type HealthHandler struct {
	logger  *zap.Logger
	service string
	version string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *zap.Logger, service, version string) *HealthHandler {
	return &HealthHandler{
		logger:  logger,
		service: service,
		version: version,
	}
}

// HealthCheck handles GET /health requests
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	stats := h.getSystemStats()
	
	response := HealthCheckResponse{
		Status:  "healthy",
		Version: h.version,
		Service: h.service,
		SystemStats: SystemStats{
			Goroutines: stats.Goroutines,
			CPUThreads: stats.CPUThreads,
			MemoryMB:   stats.MemoryMB,
		},
		Timestamp: c.GetHeader("Date"),
	}
	
	h.logger.Info("Health check",
		zap.String("service", h.service),
		zap.String("status", response.Status),
		zap.Int("goroutines", stats.Goroutines),
	)
	
	c.JSON(http.StatusOK, response)
}

// ReadyCheck handles GET /ready requests for readiness probes
func (h *HealthHandler) ReadyCheck(c *gin.Context) {
	// Check if service is ready to serve requests
	// Add your readiness logic here (database connections, external services, etc.)
	
	response := map[string]interface{}{
		"status":  "ready",
		"service": h.service,
		"version": h.version,
	}
	
	h.logger.Info("Readiness check passed", zap.String("service", h.service))
	
	c.JSON(http.StatusOK, response)
}

// LiveCheck handles GET /live requests for liveness probes
func (h *HealthHandler) LiveCheck(c *gin.Context) {
	// Basic liveness check - service is running
	response := map[string]interface{}{
		"status":  "alive",
		"service": h.service,
		"version": h.version,
	}
	
	h.logger.Debug("Liveness check passed", zap.String("service", h.service))
	
	c.JSON(http.StatusOK, response)
}

func (h *HealthHandler) getSystemStats() SystemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return SystemStats{
		Goroutines: runtime.NumGoroutine(),
		CPUThreads: runtime.GOMAXPROCS(0),
		MemoryMB:   m.Alloc / 1024 / 1024,
	}
}