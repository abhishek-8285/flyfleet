package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DocumentHandlers handles vehicle document upload and listing.
type DocumentHandlers struct {
	docsSvc  *service.DocumentsService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewDocumentHandlers(svc *service.DocumentsService, logger *zap.Logger) *DocumentHandlers {
	return &DocumentHandlers{docsSvc: svc, logger: logger, validate: validator.New()}
}

// RegisterRoutes mounts document routes inside the vehicles/{id} subrouter.
func (h *DocumentHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/{id}/documents", h.ListDocuments)
	r.Post("/{id}/documents", h.UploadDocument)
	r.Get("/{id}/documents/{docId}/verify", h.VerifyDocumentHash)
}

// UploadDocument godoc — POST /vehicles/{id}/documents
func (h *DocumentHandlers) UploadDocument(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Organization ID required")
		return
	}

	var req models.UploadDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	uploaderID := h.getUserID(r)
	doc, err := h.docsSvc.Upload(r.Context(), vehicleID, tenantID, &req, uploaderID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		h.logger.Error("upload document failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to upload document")
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{"data": doc})
}

// ListDocuments godoc — GET /vehicles/{id}/documents
func (h *DocumentHandlers) ListDocuments(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Organization ID required")
		return
	}

	docs, err := h.docsSvc.ListDocuments(r.Context(), vehicleID, tenantID)
	if err != nil {
		h.logger.Error("list documents failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to list documents")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"documents": docs, "total": len(docs)})
}

// VerifyDocumentHash godoc — GET /vehicles/{id}/documents/{docId}/verify?hash=<sha256hex>
// Compares the caller-provided hash against the stored hash. Used for integrity checks.
func (h *DocumentHandlers) VerifyDocumentHash(w http.ResponseWriter, r *http.Request) {
	docIDStr := chi.URLParam(r, "docId")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Organization ID required")
		return
	}

	providedHash := r.URL.Query().Get("hash")
	if len(providedHash) != 64 {
		h.writeError(w, http.StatusBadRequest, "hash query param must be a 64-character SHA-256 hex string")
		return
	}

	doc, err := h.docsSvc.GetDocument(r.Context(), docID, tenantID)
	if err != nil {
		h.logger.Error("get document failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve document")
		return
	}
	if doc == nil {
		h.writeError(w, http.StatusNotFound, "Document not found")
		return
	}

	// Normalize to lowercase before comparison (SHA-256 hex is case-insensitive)
	storedHash := strings.ToLower(doc.Hash)
	providedHash = strings.ToLower(providedHash)

	// SECURITY: constant-time compare prevents timing side-channel attacks.
	// Both slices are equal length (64 hex chars) after normalization.
	match := subtle.ConstantTimeCompare([]byte(storedHash), []byte(providedHash)) == 1

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"document_id":   doc.ID,
		"provided_hash": providedHash, // echo normalized form
		"match":         match,
		"integrity":     map[bool]string{true: "VERIFIED", false: "MISMATCH"}[match],
		// stored_hash intentionally omitted from response to avoid oracle attacks
	})
}

func (h *DocumentHandlers) writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (h *DocumentHandlers) writeError(w http.ResponseWriter, code int, msg string) {
	h.writeJSON(w, code, map[string]string{"error": msg})
}

func (h *DocumentHandlers) getTenantID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.TenantIDKey); val != nil {
		if s, ok := val.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func (h *DocumentHandlers) getUserID(r *http.Request) *uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if s, ok := val.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				return &id
			}
		}
	}
	return nil
}
