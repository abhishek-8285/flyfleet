package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/document/models"
	"github.com/flyfleet/fleet-core/internal/domains/document/service"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DocumentHandler struct {
	svc    *service.DocumentService
	logger *zap.Logger
}

func NewDocumentHandler(svc *service.DocumentService, logger *zap.Logger) *DocumentHandler {
	return &DocumentHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *DocumentHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/upload", h.UploadDocument)
	r.Get("/", h.GetDocuments)
	return r
}

func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get tenant and user IDs from context
	tenantID, ok := ctxkeys.GetTenantID(ctx)
	if !ok {
		h.logger.Error("tenant ID not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	userID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		// Fallback for testing/unauthenticated
		userID = uuid.Nil
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		h.logger.Error("failed to parse multipart form", zap.Error(err))
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("failed to get file from form", zap.Error(err))
		http.Error(w, "file missing", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Extract metadata
	entityType := r.FormValue("entity_type")
	entityIDStr := r.FormValue("entity_id")
	docType := r.FormValue("document_type")
	expiryStr := r.FormValue("expiry_date")

	if entityType == "" || entityIDStr == "" || docType == "" {
		http.Error(w, "missing required metadata", http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		http.Error(w, "invalid entity_id", http.StatusBadRequest)
		return
	}

	var expiryDate *time.Time
	if expiryStr != "" {
		t, err := time.Parse(time.RFC3339, expiryStr)
		if err == nil {
			expiryDate = &t
		}
	}

	req := &models.UploadDocumentRequest{
		EntityType:   entityType,
		EntityID:     entityID,
		DocumentType: docType,
		ExpiryDate:   expiryDate,
	}

	doc, err := h.svc.Upload(ctx, tenantID, userID, req, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if err != nil {
		h.logger.Error("failed to upload document", zap.Error(err))
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

func (h *DocumentHandler) GetDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	entityType := r.URL.Query().Get("entity_type")
	entityIDStr := r.URL.Query().Get("entity_id")

	if entityType == "" || entityIDStr == "" {
		http.Error(w, "missing entity parameters", http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		http.Error(w, "invalid entity_id", http.StatusBadRequest)
		return
	}

	docs, err := h.svc.GetDocuments(ctx, entityType, entityID)
	if err != nil {
		h.logger.Error("failed to get documents", zap.Error(err))
		http.Error(w, "retrieval failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": docs,
	})
}
