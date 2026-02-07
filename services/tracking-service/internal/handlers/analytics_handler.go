package handlers

import (
	"net/http"
	"time"

	"github.com/flyfleet/tracking-service/internal/analytics"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AnalyticsHandler struct {
	analytics *analytics.TelemetryAnalytics
	logger    *zap.Logger
}

func NewAnalyticsHandler(analytics *analytics.TelemetryAnalytics, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analytics: analytics,
		logger:    logger,
	}
}

// GetVehicleAnalytics handles GET /analytics/vehicle/:id
func (ah *AnalyticsHandler) GetVehicleAnalytics(c *gin.Context) {
	vehicleIDStr := c.Param("id")
	
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID"})
		return
	}

	// Get query parameters for time range
	startStr := c.Query("start")
	endStr := c.Query("end")
	
	startTime := time.Now().AddDate(0, 0, -7) // Default to last 7 days
	endTime := time.Now()
	
	if startStr != "" {
		if parsedStart, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = parsedStart
		}
	}
	
	if endStr != "" {
		if parsedEnd, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = parsedEnd
		}
	}

	// Use the variables to avoid 'declared but not used' errors
	_ = startTime
	_ = endTime

	// Process analytics for the specified time range
	err = ah.analytics.ProcessAdvancedAnalytics(vehicleID, startTime, endTime)
	if err != nil {
		ah.logger.Error("Failed to process analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Analytics processed"})
}

// GetAnalyticsSummary handles GET /analytics/summary
func (ah *AnalyticsHandler) GetAnalyticsSummary(c *gin.Context) {
	// This would typically aggregate analytics across multiple vehicles
	// For now, returning a placeholder response
	summary := map[string]interface{}{
		"total_analyzed_vehicles": 0,
		"anomalies_detected":      0,
		"report_period":           "last_24_hours",
		"generated_at":            time.Now(),
	}

	c.JSON(http.StatusOK, summary)
}

// GetVehicleAnomalies handles GET /analytics/anomalies/:id
func (ah *AnalyticsHandler) GetVehicleAnomalies(c *gin.Context) {
	vehicleIDStr := c.Param("id")
	
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID"})
		return
	}

	// Get query parameters for time range
	startStr := c.Query("start")
	endStr := c.Query("end")
	
	startTime := time.Now().AddDate(0, 0, -7) // Default to last 7 days
	endTime := time.Now()
	
	if startStr != "" {
		if parsedStart, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = parsedStart
		}
	}
	
	if endStr != "" {
		if parsedEnd, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = parsedEnd
		}
	}

	// This would typically query the database for stored anomalies
	// For now, returning a placeholder response
	// Using the variables to avoid 'declared but not used' errors
	_ = vehicleID
	_ = startTime
	_ = endTime
	
	anomalies := []map[string]interface{}{
		{
			"type":         "speed_anomaly",
			"description":  "Unusual speed patterns detected",
			"timestamp":    time.Now(),
			"confidence":   0.85,
			"score":        85.5,
			"recommendation": "Review driving behavior",
		},
	}

	c.JSON(http.StatusOK, anomalies)
}

// RegisterRoutes registers analytics-related routes
func (ah *AnalyticsHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/analytics/vehicle/:id", ah.GetVehicleAnalytics)
	router.GET("/analytics/summary", ah.GetAnalyticsSummary)
	router.GET("/analytics/anomalies/vehicle/:id", ah.GetVehicleAnomalies)
}