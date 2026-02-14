package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/flyfleet/support-service/internal/models"
	"github.com/flyfleet/support-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TicketHandlers struct {
	service service.TicketService
	logger  *zap.Logger
}

func NewTicketHandlers(service service.TicketService, logger *zap.Logger) *TicketHandlers {
	return &TicketHandlers{service: service, logger: logger}
}

// Dashboard APIs
func (h *TicketHandlers) GetDashboardSummary(c *gin.Context) {
	orgID := h.getOrgID(c)
	
	summary, err := h.service.GetDashboardSummary(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get dashboard summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *TicketHandlers) GetRecentActivity(c *gin.Context) {
	orgID := h.getOrgID(c)
	
	limitStr := c.Query("limit")
	limit := 10
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	tickets, err := h.service.GetRecentActivity(c.Request.Context(), orgID, limit)
	if err != nil {
		h.logger.Error("Failed to get recent activity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent activity"})
		return
	}

	// Transform for dashboard display
	type ActivityItem struct {
		ID          string `json:"id"`
		Subject     string `json:"subject"`
		Reporter    string `json:"reporter"`
		Priority    string `json:"priority"`
		AssignedTo  string `json:"assigned_to"`
		LastUpdated string `json:"last_updated"`
		Status      string `json:"status"`
	}

	var activity []ActivityItem
	for _, ticket := range tickets {
		assignedTo := "Unassigned"
		if ticket.AssignedTo != nil {
			assignedTo = ticket.AssignedTo.String()[:8] // Short ID for display
		}
		
		activity = append(activity, ActivityItem{
			ID:          ticket.TicketID,
			Subject:     ticket.Title,
			Reporter:    "System", // Would come from user lookup
			Priority:    strings.ToUpper(ticket.Priority),
			AssignedTo:  assignedTo,
			LastUpdated: "10 mins ago", // Would calculate real time difference
			Status:      strings.ReplaceAll(strings.ToUpper(ticket.Status), "_", " "),
		})
	}

	c.JSON(http.StatusOK, activity)
}

// Incident Management APIs
func (h *TicketHandlers) CreateTicket(c *gin.Context) {
	var req models.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validation
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	orgID := h.getOrgID(c)
	userID := h.getUserID(c) // Would come from auth middleware

	ticket, err := h.service.CreateTicket(c.Request.Context(), &req, orgID, userID)
	if err != nil {
		h.logger.Error("Failed to create ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (h *TicketHandlers) GetTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	ticket, err := h.service.GetTicket(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandlers) ListTickets(c *gin.Context) {
	var filter models.TicketFilter
	
	// Parse query parameters
	if status := c.Query("status"); status != "" {
		filter.Status = status
	}
	if priority := c.Query("priority"); priority != "" {
		filter.Priority = priority
	}
	if category := c.Query("category"); category != "" {
		filter.Category = category
	}
	
	pageStr := c.Query("page")
	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = page
		}
	}
	
	pageSizeStr := c.Query("page_size")
	if pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			filter.PageSize = pageSize
		}
	}

	tickets, totalCount, err := h.service.ListTickets(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list tickets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tickets":   tickets,
		"total":     totalCount,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

func (h *TicketHandlers) UpdateTicketStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var req models.UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	userID := h.getUserID(c)
	err = h.service.UpdateTicketStatus(c.Request.Context(), id, req.Status, userID)
	if err != nil {
		h.logger.Error("Failed to update ticket status", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

func (h *TicketHandlers) AssignTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var req models.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := h.getUserID(c)
	err = h.service.AssignTicket(c.Request.Context(), id, req.AssignedTo, userID)
	if err != nil {
		h.logger.Error("Failed to assign ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket assigned successfully"})
}

func (h *TicketHandlers) AddComment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var req models.AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
		return
	}

	userID := h.getUserID(c)
	authorName := "Support Agent" // Would come from user context
	
	err = h.service.AddComment(c.Request.Context(), id, userID, &req, authorName)
	if err != nil {
		h.logger.Error("Failed to add comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment added successfully"})
}

func (h *TicketHandlers) GetTicketComments(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	comments, err := h.service.GetTicketComments(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// Helper methods
func (h *TicketHandlers) getOrgID(_ *gin.Context) uuid.UUID {
	// In real implementation, this would come from auth context
	// For now, using a default org ID
	orgID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	return orgID
}

func (h *TicketHandlers) getUserID(_ *gin.Context) uuid.UUID {
	// In real implementation, this would come from auth context
	// For now, using a default user ID
	userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000002")
	return userID
}