package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
)

// UpdateDutyStatus godoc
// @Summary Update driver duty status (HOOS)
// @Description Updates the current duty status of the driver
// @Tags driver-mobile
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param body body models.DutyStatusLog true "Duty Status Log"
// @Success 200 {object} models.DutyStatusLog
// @Router /drivers/{id}/duty-status [post]
func (h *DriverHandler) UpdateDutyStatus(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	var req models.DutyStatusLog
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateDutyStatus(r.Context(), driverID, &req); err != nil {
		h.logger.Error("failed to update duty status", zap.Error(err))
		http.Error(w, "Failed to update duty status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
}

// SendMessage godoc
// @Summary Send a message to dispatch
// @Description Sends a text, audio, or image message from the driver
// @Tags driver-mobile
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param body body models.DriverMessage true "Message content"
// @Success 201 {object} models.DriverMessage
// @Router /drivers/{id}/messages [post]
func (h *DriverHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	var req struct {
		Content       string `json:"content"`
		Type          string `json:"type"` // TEXT, AUDIO, IMAGE
		AttachmentURL string `json:"attachment_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Default type to TEXT if empty
	if req.Type == "" {
		req.Type = "TEXT"
	}

	msg, err := h.service.SendMessage(r.Context(), driverID, req.Content, req.Type, req.AttachmentURL)
	if err != nil {
		h.logger.Error("failed to send message", zap.Error(err))
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

// GetMessages godoc
// @Summary Get driver messages
// @Description Retrieves a paginated list of messages for the driver
// @Tags driver-mobile
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {array} models.DriverMessage
// @Router /drivers/{id}/messages [get]
func (h *DriverHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 { page = 1 }
	
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 { pageSize = 20 }

	messages, err := h.service.GetMessages(r.Context(), driverID, page, pageSize)
	if err != nil {
		h.logger.Error("failed to get messages", zap.Error(err))
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}

// SubmitExpense godoc
// @Summary Submit an expense
// @Description Submits a new expense claim for the driver
// @Tags driver-mobile
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param body body models.DriverExpense true "Expense details"
// @Success 201 {object} models.DriverExpense
// @Router /drivers/{id}/expenses [post]
func (h *DriverHandler) SubmitExpense(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	var req models.DriverExpense
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	expense, err := h.service.SubmitExpense(r.Context(), driverID, &req)
	if err != nil {
		h.logger.Error("failed to submit expense", zap.Error(err))
		http.Error(w, "Failed to submit expense", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

// GetExpenses godoc
// @Summary Get driver expenses
// @Description Retrieves a list of expenses for the driver
// @Tags driver-mobile
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {array} models.DriverExpense
// @Router /drivers/{id}/expenses [get]
func (h *DriverHandler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	expenses, err := h.service.GetExpenses(r.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get expenses", zap.Error(err))
		http.Error(w, "Failed to get expenses", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expenses)
}

// TriggerSOS godoc
// @Summary Trigger SOS
// @Description Triggers an emergency SOS alert
// @Tags driver-mobile
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param body body models.Location true "Current location"
// @Success 200 {object} map[string]string
// @Router /drivers/{id}/sos [post]
func (h *DriverHandler) TriggerSOS(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	var loc models.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.TriggerSOS(r.Context(), driverID, &loc); err != nil {
		h.logger.Error("failed to trigger SOS", zap.Error(err))
		http.Error(w, "Failed to trigger SOS", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "SOS triggered"})
}
