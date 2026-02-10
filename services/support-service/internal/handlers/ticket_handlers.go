package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/flyfleet/support-service/internal/models"
	"github.com/flyfleet/support-service/internal/service"
	"github.com/go-chi/chi/v5"
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
func (h *TicketHandlers) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	
	summary, err := h.service.GetDashboardSummary(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get dashboard summary", zap.Error(err))
		http.Error(w, "Failed to get dashboard summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *TicketHandlers) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	tickets, err := h.service.GetRecentActivity(r.Context(), orgID, limit)
	if err != nil {
		h.logger.Error("Failed to get recent activity", zap.Error(err))
		http.Error(w, "Failed to get recent activity", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}

// Incident Management APIs
func (h *TicketHandlers) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	orgID := h.getOrgID(r)
	userID := h.getUserID(r) // Would come from auth middleware

	ticket, err := h.service.CreateTicket(r.Context(), &req, orgID, userID)
	if err != nil {
		h.logger.Error("Failed to create ticket", zap.Error(err))
		http.Error(w, "Failed to create ticket", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

func (h *TicketHandlers) GetTicket(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	ticket, err := h.service.GetTicket(r.Context(), id)
	if err != nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

func (h *TicketHandlers) ListTickets(w http.ResponseWriter, r *http.Request) {
	var filter models.TicketFilter
	
	// Parse query parameters
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		filter.Priority = priority
	}
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Category = category
	}
	
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = page
		}
	}
	
	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			filter.PageSize = pageSize
		}
	}

	tickets, totalCount, err := h.service.ListTickets(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list tickets", zap.Error(err))
		http.Error(w, "Failed to list tickets", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tickets": tickets,
		"total":   totalCount,
		"page":    filter.Page,
		"page_size": filter.PageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TicketHandlers) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateTicketStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	err = h.service.UpdateTicketStatus(r.Context(), id, req.Status, userID)
	if err != nil {
		h.logger.Error("Failed to update ticket status", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Status updated successfully"}`))
}

func (h *TicketHandlers) AssignTicket(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	var req models.AssignTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	err = h.service.AssignTicket(r.Context(), id, req.AssignedTo, userID)
	if err != nil {
		h.logger.Error("Failed to assign ticket", zap.Error(err))
		http.Error(w, "Failed to assign ticket", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Ticket assigned successfully"}`))
}

func (h *TicketHandlers) AddComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	var req models.AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	authorName := "Support Agent" // Would come from user context
	
	err = h.service.AddComment(r.Context(), id, userID, &req, authorName)
	if err != nil {
		h.logger.Error("Failed to add comment", zap.Error(err))
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Comment added successfully"}`))
}

func (h *TicketHandlers) GetTicketComments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	comments, err := h.service.GetTicketComments(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get comments", zap.Error(err))
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// Helper methods
func (h *TicketHandlers) getOrgID(r *http.Request) uuid.UUID {
	// In real implementation, this would come from auth context
	// For now, using a default org ID
	orgID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	return orgID
}

func (h *TicketHandlers) getUserID(r *http.Request) uuid.UUID {
	// In real implementation, this would come from auth context
	// For now, using a default user ID
	userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000002")
	return userID
}