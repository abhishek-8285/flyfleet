package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *BillingHandler) IssueCreditNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		models.CreditNote
		Items []models.CreditNoteItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	req.CreditNote.TenantID = h.getTenantID(r)
	if req.CreditNote.TenantID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	err := h.creditNoteService.IssueCreditNote(r.Context(), &req.CreditNote, req.Items)
	if err != nil {
		h.logger.Error("Failed to issue credit note", zap.Error(err))
		h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, req.CreditNote, http.StatusCreated)
}

func (h *BillingHandler) ListCreditNotes(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	page := h.readInt(r, "page", 1)
	size := h.readInt(r, "size", 20)

	results, total, err := h.creditNoteService.ListCreditNotes(r.Context(), tenantID, page, size)
	if err != nil {
		h.logger.Error("Failed to list credit notes", zap.Error(err))
		h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"data":  results,
		"total": total,
		"page":  page,
		"size":  size,
	}, http.StatusOK)
}
