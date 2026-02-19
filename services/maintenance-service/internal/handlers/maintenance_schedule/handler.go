package maintenance_schedule

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/maintenance-service/internal/domain/maintenance_schedule"
	"github.com/flyfleet/maintenance-service/internal/middleware"
	"github.com/flyfleet/maintenance-service/pkg/errors"
	"github.com/flyfleet/maintenance-service/pkg/respond"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	service  maintenance_schedule.Service
	validate *validator.Validate
}

func NewHandler(service maintenance_schedule.Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, errors.BadRequest("invalid JSON"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeValidationProblem(w, err)
		return
	}

	schedule, err := h.service.Create(ctx, orgID, req.ToDomain())
	if err != nil {
		errors.WriteProblem(w, errors.Internal(uuid.New().String()))
		return
	}

	respond.JSON(w, http.StatusCreated, schedule)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	schedule, err := h.service.GetByID(ctx, orgID, id)
	if err != nil {
		errors.WriteProblem(w, errors.NotFound(err.Error()))
		return
	}

	respond.JSON(w, http.StatusOK, schedule)
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

	schedules, total, err := h.service.List(ctx, orgID, limit, (page-1)*limit)
	if err != nil {
		errors.WriteProblem(w, errors.Internal(uuid.New().String()))
		return
	}

	respond.Paginated(w, http.StatusOK, schedules, total, page, limit)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, errors.BadRequest("invalid JSON"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeValidationProblem(w, err)
		return
	}

	schedule, err := h.service.Update(ctx, orgID, id, req.ToDomain())
	if err != nil {
		errors.WriteProblem(w, errors.Internal(uuid.New().String()))
		return
	}

	respond.JSON(w, http.StatusOK, schedule)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	if err := h.service.Delete(ctx, orgID, id); err != nil {
		errors.WriteProblem(w, errors.Internal(uuid.New().String()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, _ := middleware.GetOrgID(ctx)
	id, _ := uuid.Parse(chi.URLParam(r, "id"))

	var req CompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, errors.BadRequest("invalid JSON"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeValidationProblem(w, err)
		return
	}

	schedule, err := h.service.MarkComplete(ctx, orgID, id, req.ActualCost, req.Notes)
	if err != nil {
		errors.WriteProblem(w, errors.Internal(uuid.New().String()))
		return
	}

	respond.JSON(w, http.StatusOK, schedule)
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

