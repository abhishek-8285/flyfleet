package invoice

import (
	"encoding/json"
	"net/http"
	"strconv"


	"github.com/flyfleet/fleet-core-service/internal/domains/billing/domain/invoice"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	service  invoice.Service
	validate *validator.Validate
}

func NewHandler(service invoice.Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)

	var req CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeValidationProblem(w, err)
		return
	}

	inv, err := h.service.CreateInvoice(ctx, orgID, req.ToDomain())
	if err != nil {
		http.Error(w, "internal server error: "+uuid.New().String(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inv)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	inv, err := h.service.GetInvoice(ctx, orgID, id)
	if err != nil {
		http.Error(w, "invoice not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	invoices, total, err := h.service.ListInvoices(ctx, orgID, limit, (page-1)*limit)
	if err != nil {
		http.Error(w, "internal server error: "+uuid.New().String(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": invoices,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) Void(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	if err := h.service.VoidInvoice(ctx, orgID, id); err != nil {
		http.Error(w, "internal server error: "+uuid.New().String(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeValidationProblem(w http.ResponseWriter, err error) {
	http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
}
