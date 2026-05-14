package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/fleet-core/internal/domains/invoicing"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Service interface {
	CreateInvoice(ctx context.Context, tenantID uuid.UUID, inv *invoicing.Invoice) (*invoicing.Invoice, error)
	GetInvoice(ctx context.Context, tenantID, id uuid.UUID) (*invoicing.Invoice, error)
	ListInvoices(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]invoicing.Invoice, int64, error)
	VoidInvoice(ctx context.Context, tenantID, id uuid.UUID) error
	GeneratePDF(ctx context.Context, tenantID, id uuid.UUID) ([]byte, string, error)
}

type Handler struct {
	service  Service
	validate *validator.Validate
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := ctxkeys.GetTenantID(ctx)
	if !ok {
		http.Error(w, "tenant context required", http.StatusUnauthorized)
		return
	}

	var req invoicing.CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	inv, err := h.service.CreateInvoice(ctx, tenantID, req.ToDomain())
	if err != nil {
		http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inv)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctxkeys.GetTenantID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	inv, err := h.service.GetInvoice(ctx, tenantID, id)
	if err != nil {
		http.Error(w, "invoice not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctxkeys.GetTenantID(ctx)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	invoices, total, err := h.service.ListInvoices(ctx, tenantID, limit, (page-1)*limit)
	if err != nil {
		http.Error(w, "internal server error: "+uuid.New().String(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invoices": invoices,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) Void(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctxkeys.GetTenantID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	if err := h.service.VoidInvoice(ctx, tenantID, id); err != nil {
		http.Error(w, "internal server error: "+uuid.New().String(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := ctxkeys.GetTenantID(ctx)
	if !ok {
		http.Error(w, "tenant context required", http.StatusUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid invoice ID", http.StatusBadRequest)
		return
	}

	pdfData, filename, err := h.service.GeneratePDF(ctx, tenantID, id)
	if err != nil {
		http.Error(w, "failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfData)))
	w.Write(pdfData)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Post("/{id}/void", h.Void)
	r.Get("/{id}/pdf", h.DownloadPDF)
}
