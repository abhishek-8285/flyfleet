package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/support/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/support/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
)

type TicketHandlers struct {
	service service.TicketService
	logger  *zap.Logger
}

func NewTicketHandlers(service service.TicketService, logger *zap.Logger) *TicketHandlers {
	return &TicketHandlers{service: service, logger: logger}
}

func (h *TicketHandlers) RegisterRoutes(r chi.Router) {
	// Dashboard APIs
	r.Get("/dashboard/summary", h.GetDashboardSummary)
	r.Get("/dashboard/recent-activity", h.GetRecentActivity)

	// Incident Management
	r.Get("/incidents", h.ListTickets)
	r.Post("/incidents", h.CreateTicket)
	r.Get("/incidents/{id}", h.GetTicket)
	r.Patch("/incidents/{id}/status", h.UpdateTicketStatus)
	r.Post("/incidents/{id}/assign", h.AssignTicket)
	r.Post("/incidents/{id}/comment", h.AddComment)
	r.Get("/incidents/{id}/comments", h.GetTicketComments)
}

// Dashboard APIs
func (h *TicketHandlers) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	
	summary, err := h.service.GetDashboardSummary(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get dashboard summary", zap.Error(err))
		h.errorJSON(w, "Failed to get dashboard summary", http.StatusInternalServerError)
		return
	}

	h.json(w, summary, http.StatusOK)
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
		h.errorJSON(w, "Failed to get recent activity", http.StatusInternalServerError)
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

	h.json(w, activity, http.StatusOK)
}

// Incident Management APIs
func (h *TicketHandlers) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.errorJSON(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	orgID := h.getOrgID(r)
	userID := h.getUserID(r)

	ticket, err := h.service.CreateTicket(r.Context(), &req, orgID, userID)
	if err != nil {
		h.logger.Error("Failed to create ticket", zap.Error(err))
		h.errorJSON(w, "Failed to create ticket", http.StatusInternalServerError)
		return
	}

	h.json(w, ticket, http.StatusCreated)
}

func (h *TicketHandlers) GetTicket(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	ticket, err := h.service.GetTicket(r.Context(), id)
	if err != nil {
		h.errorJSON(w, "Ticket not found", http.StatusNotFound)
		return
	}

	h.json(w, ticket, http.StatusOK)
}

func (h *TicketHandlers) ListTickets(w http.ResponseWriter, r *http.Request) {
	var filter models.TicketFilter
	
	// Parse query parameters
	query := r.URL.Query()
	if status := query.Get("status"); status != "" {
		filter.Status = status
	}
	if priority := query.Get("priority"); priority != "" {
		filter.Priority = priority
	}
	if category := query.Get("category"); category != "" {
		filter.Category = category
	}
	if q := query.Get("q"); q != "" {
		filter.Query = q
	}
	
	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = page
		}
	}
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			filter.PageSize = pageSize
		}
	}

	tickets, totalCount, err := h.service.ListTickets(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list tickets", zap.Error(err))
		h.errorJSON(w, "Failed to list tickets", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"tickets":   tickets,
		"total":     totalCount,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	}, http.StatusOK)
}

func (h *TicketHandlers) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateTicketStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.errorJSON(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	err = h.service.UpdateTicketStatus(r.Context(), id, req.Status, userID)
	if err != nil {
		h.logger.Error("Failed to update ticket status", zap.Error(err))
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.json(w, map[string]string{"message": "Status updated successfully"}, http.StatusOK)
}

func (h *TicketHandlers) AssignTicket(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	var req models.AssignTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.errorJSON(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	err = h.service.AssignTicket(r.Context(), id, req.AssignedTo, userID)
	if err != nil {
		h.logger.Error("Failed to assign ticket", zap.Error(err))
		h.errorJSON(w, "Failed to assign ticket", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"message": "Ticket assigned successfully"}, http.StatusOK)
}

func (h *TicketHandlers) AddComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	var req models.AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.errorJSON(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	authorName := "Support Agent" // Would come from user context

	err = h.service.AddComment(r.Context(), id, userID, &req, authorName)
	if err != nil {
		h.logger.Error("Failed to add comment", zap.Error(err))
		h.errorJSON(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"message": "Comment added successfully"}, http.StatusOK)
}

func (h *TicketHandlers) GetTicketComments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	comments, err := h.service.GetTicketComments(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get comments", zap.Error(err))
		h.errorJSON(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}

	h.json(w, comments, http.StatusOK)
}

// Helper methods
func (h *TicketHandlers) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *TicketHandlers) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}

func (h *TicketHandlers) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(ctxkeys.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func (h *TicketHandlers) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(ctxkeys.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}