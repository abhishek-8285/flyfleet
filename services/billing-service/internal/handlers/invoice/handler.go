package invoice

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/billing-service/internal/domain/invoice"
	"github.com/flyfleet/billing-service/internal/middleware"
	"github.com/flyfleet/billing-service/pkg/errors"
	"github.com/flyfleet/billing-service/pkg/respond"
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
		errors.WriteProblem(w, errors.BadRequest("invalid JSON"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeValidationProblem(w, err)
		return
	}

	inv, err := h.service.CreateInvoice(ctx, orgID, req.ToDomain())
	if err != nil {
		errors.WriteProblem(w, errors.Internal(uuid.New().String()))
		return
	}

	respond.JSON(w, http.StatusCreated, inv)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	inv, err := h.service.GetInvoice(ctx, orgID, id)
	if err != nil {
		errors.WriteProblem(w, errors.NotFound("invoice not found"))
		return
	}

	respond.JSON(w, http.StatusOK, inv)
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
		errors.WriteProblem(w, errors.Internal(uuid.New().String()))
		return
	}

	respond.Paginated(w, http.StatusOK, invoices, total, page, limit)
}

func (h *Handler) Void(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	if err := h.service.VoidInvoice(ctx, orgID, id); err != nil {
		errors.WriteProblem(w, errors.Internal(uuid.New().String()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeValidationProblem(w http.ResponseWriter, err error) {
	var validationErrors []errors.ValidationError
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			validationErrors = append(validationErrors, errors.ValidationError{
				Field:   e.Field(),
				Message: e.Tag(),
			})
		}
	}
	errors.WriteProblem(w, errors.ValidationFailed(validationErrors))
}

