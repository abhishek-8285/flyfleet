package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/auth-service/internal/middleware"
	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GET /api/v1/admin/vehicles/{id}/documents
func (h *AdminHandlers) ListDocuments(w http.ResponseWriter, r *http.Request) {
	vehicleIDStr := r.URL.Query().Get("vehicle_id") // or from path? Check router.
	// Router usually puts path params in context or we parse from URL if using simple mux without path helpers
	// main.go uses standard http.ServeMux or similar?
	// main.go seems to use `gorilla/mux` or similar if path checking. 
	// Let's assume standard query param or path parsing logic used in other handlers.
	// AdminHandlers.GetVehicle uses `r.URL.Query().Get("id")`.
	// So I will expect `vehicle_id` as query param or generic ID.
	
	if vehicleIDStr == "" {
		// It might be in the path /vehicles/{id}/documents.
		// If using standard library mux in Go 1.22+, it supports path values.
		// But let's check assumptions. 
		// If the route is `/api/v1/admin/vehicles/{id}/documents`, the `id` is the vehicle ID.
		vehicleIDStr = r.PathValue("id")
	}
	// Fallback to query
	if vehicleIDStr == "" {
		vehicleIDStr = r.URL.Query().Get("vehicle_id")
	}

	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid Vehicle ID"}`, http.StatusBadRequest)
		return
	}

	documents, err := h.service.ListDocumentsByVehicleID(r.Context(), vehicleID)
	if err != nil {
		h.logger.Error("Failed to list documents", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	
	middleware.WriteSuccess(w, documents, 1, len(documents), len(documents))
}

// POST /api/v1/admin/vehicles/{id}/documents
func (h *AdminHandlers) CreateDocument(w http.ResponseWriter, r *http.Request) {
	vehicleIDStr := r.PathValue("id")
	if vehicleIDStr == "" {
		vehicleIDStr = r.URL.Query().Get("vehicle_id")
	}
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid Vehicle ID"}`, http.StatusBadRequest)
		return
	}

	var d models.Document
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, `{"error": "Invalid Body"}`, http.StatusBadRequest)
		return
	}
	d.VehicleID = vehicleID

	doc, err := h.service.CreateDocument(r.Context(), &d)
	if err != nil {
		h.logger.Error("Failed to create document", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, doc, 1, 1, 1)
}

// GET /api/v1/admin/documents/{id}
func (h *AdminHandlers) GetDocument(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	doc, err := h.service.GetDocumentByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get document", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "Document not found"}`, http.StatusNotFound)
		return
	}

	middleware.WriteSuccess(w, doc, 1, 1, 1)
}

// PUT /api/v1/admin/documents/{id}
func (h *AdminHandlers) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var d models.Document
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, `{"error": "Invalid Body"}`, http.StatusBadRequest)
		return
	}
	d.ID = id

	doc, err := h.service.UpdateDocument(r.Context(), &d)
	if err != nil {
		h.logger.Error("Failed to update document", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, doc, 1, 1, 1)
}

// DELETE /api/v1/admin/documents/{id}
func (h *AdminHandlers) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteDocument(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete document", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, nil, 0, 0, 0)
}
