package handlers

import (
	"net/http"
	"strconv"

	"github.com/flyfleet/route-optimization-service/internal/models"
	"github.com/flyfleet/route-optimization-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RouteOptimizationHandler handles HTTP requests for route optimization
type RouteOptimizationHandler struct {
	service service.RouteOptimizationService
}

// NewRouteOptimizationHandler creates a new instance of RouteOptimizationHandler
func NewRouteOptimizationHandler(svc service.RouteOptimizationService) *RouteOptimizationHandler {
	return &RouteOptimizationHandler{
		service: svc,
	}
}

// CreateOptimizationRequest handles POST /api/v1/route-optimization/requests
func (h *RouteOptimizationHandler) CreateOptimizationRequest(c *gin.Context) {
	var request models.RouteOptimizationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateOptimizationRequest(ctx, &request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create optimization request", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, request)
}

// GetOptimizationRequest handles GET /api/v1/route-optimization/requests/:id
func (h *RouteOptimizationHandler) GetOptimizationRequest(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	request, err := h.service.GetOptimizationRequest(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Optimization request not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, request)
}

// UpdateOptimizationRequest handles PUT /api/v1/route-optimization/requests/:id
func (h *RouteOptimizationHandler) UpdateOptimizationRequest(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateOptimizationRequest(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update optimization request", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Optimization request updated"})
}

// DeleteOptimizationRequest handles DELETE /api/v1/route-optimization/requests/:id
func (h *RouteOptimizationHandler) DeleteOptimizationRequest(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteOptimizationRequest(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete optimization request", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Optimization request deleted"})
}

// ListOptimizationRequests handles GET /api/v1/route-optimization/requests
func (h *RouteOptimizationHandler) ListOptimizationRequests(c *gin.Context) {
	orgIDStr := c.Query("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "org_id query parameter is required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid org_id", "details": err.Error()})
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := c.Request.Context()
	requests, totalCount, err := h.service.ListOptimizationRequests(ctx, orgID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list optimization requests", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requests":     requests,
		"total_count":  totalCount,
		"limit":        limit,
		"offset":       offset,
		"returned_count": len(requests),
	})
}

// GetOptimizationResult handles GET /api/v1/route-optimization/results/:id
func (h *RouteOptimizationHandler) GetOptimizationResult(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	result, err := h.service.GetOptimizationResult(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Optimization result not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetOptimizationResultByRequestID handles GET /api/v1/route-optimization/results/request/:request_id
func (h *RouteOptimizationHandler) GetOptimizationResultByRequestID(c *gin.Context) {
	requestID := c.Param("request_id")

	ctx := c.Request.Context()
	result, err := h.service.GetOptimizationResultByRequestID(ctx, requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Optimization result not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListOptimizationResults handles GET /api/v1/route-optimization/results
func (h *RouteOptimizationHandler) ListOptimizationResults(c *gin.Context) {
	requestID := c.Query("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id query parameter is required"})
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := c.Request.Context()
	results, totalCount, err := h.service.ListOptimizationResults(ctx, requestID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list optimization results", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":        results,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(results),
	})
}

// SubmitOptimizationJob handles POST /api/v1/route-optimization/jobs
func (h *RouteOptimizationHandler) SubmitOptimizationJob(c *gin.Context) {
	var req struct {
		RequestID string `json:"request_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing request_id", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.SubmitOptimizationJob(ctx, req.RequestID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit optimization job", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "accepted",
		"message": "Optimization job submitted successfully",
		"request_id": req.RequestID,
	})
}

// GetOptimizationJob handles GET /api/v1/route-optimization/jobs/:id
func (h *RouteOptimizationHandler) GetOptimizationJob(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	job, err := h.service.GetOptimizationJob(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Optimization job not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListOptimizationJobs handles GET /api/v1/route-optimization/jobs
func (h *RouteOptimizationHandler) ListOptimizationJobs(c *gin.Context) {
	requestID := c.Query("request_id")
	status := c.Query("status") // optional filter

	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id query parameter is required"})
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx := c.Request.Context()
	jobs, totalCount, err := h.service.ListOptimizationJobs(ctx, requestID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list optimization jobs", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":           jobs,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(jobs),
	})
}

// RegisterRoutes registers all route optimization API routes
func (h *RouteOptimizationHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		optimization := v1.Group("/route-optimization")
		{
			// Requests
			optimization.POST("/requests", h.CreateOptimizationRequest)
			optimization.GET("/requests/:id", h.GetOptimizationRequest)
			optimization.PUT("/requests/:id", h.UpdateOptimizationRequest)
			optimization.DELETE("/requests/:id", h.DeleteOptimizationRequest)
			optimization.GET("/requests", h.ListOptimizationRequests)

			// Results
			optimization.GET("/results/:id", h.GetOptimizationResult)
			optimization.GET("/results/request/:request_id", h.GetOptimizationResultByRequestID)
			optimization.GET("/results", h.ListOptimizationResults)

			// Jobs
			optimization.POST("/jobs", h.SubmitOptimizationJob)
			optimization.GET("/jobs/:id", h.GetOptimizationJob)
			optimization.GET("/jobs", h.ListOptimizationJobs)
		}
	}
}