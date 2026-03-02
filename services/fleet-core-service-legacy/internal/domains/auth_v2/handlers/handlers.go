package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	authservice "github.com/flyfleet/fleet-core-service/internal/domains/auth_v2/service"
	pb "github.com/flyfleet/proto/auth/v1"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handlers struct {
	svc    *authservice.AuthService
	logger *zap.Logger
	mu     sync.Mutex
	limits map[string]*rateState
}

func NewHandlers(svc *authservice.AuthService, logger *zap.Logger) *Handlers {
	return &Handlers{svc: svc, logger: logger, limits: map[string]*rateState{}}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/register", h.Register)
		r.Post("/validate-token", h.ValidateToken)
		r.Get("/users/{id}", h.GetUser)
		r.Get("/organizations/{id}", h.GetOrganization)
		r.Get("/organizations", h.ListOrganizations)
		r.Get("/stats", h.GetStats)
		r.Get("/security-telemetry", h.GetSecurityTelemetry)
		r.Get("/audit-logs", h.GetAuditLogs)
	})
}

func (h *Handlers) json(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handlers) errorJSON(w http.ResponseWriter, status int, msg string) {
	h.json(w, status, map[string]string{"error": msg})
}

type rateState struct {
	Count int
	Reset time.Time
}

func (h *Handlers) clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	return r.RemoteAddr
}

func (h *Handlers) allowRate(key string, limit int, window time.Duration) bool {
	now := time.Now()
	h.mu.Lock()
	defer h.mu.Unlock()

	st, ok := h.limits[key]
	if !ok || now.After(st.Reset) {
		h.limits[key] = &rateState{Count: 1, Reset: now.Add(window)}
		return true
	}
	if st.Count >= limit {
		return false
	}
	st.Count++
	return true
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if !h.allowRate("login:"+h.clientIP(r), 5, time.Minute) {
		h.errorJSON(w, http.StatusTooManyRequests, "too many login attempts")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error("auth login failed", zap.Error(err))
		h.errorJSON(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	h.json(w, http.StatusOK, resp)
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	if !h.allowRate("register:"+h.clientIP(r), 3, time.Minute) {
		h.errorJSON(w, http.StatusTooManyRequests, "too many registration attempts")
		return
	}
	var req pb.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		h.logger.Error("auth register failed", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to register user")
		return
	}
	h.json(w, http.StatusCreated, resp)
}

func (h *Handlers) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req pb.ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.svc.ValidateToken(r.Context(), req.Token)
	if err != nil {
		h.json(w, http.StatusOK, map[string]any{"valid": false})
		return
	}
	h.json(w, http.StatusOK, map[string]any{"valid": true, "user": user})
}

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.svc.GetUserByID(r.Context(), id)
	if err != nil {
		h.errorJSON(w, http.StatusNotFound, "user not found")
		return
	}
	h.json(w, http.StatusOK, user)
}

func (h *Handlers) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	org, err := h.svc.GetOrganizationByID(r.Context(), id)
	if err != nil {
		h.errorJSON(w, http.StatusNotFound, "organization not found")
		return
	}
	h.json(w, http.StatusOK, org)
}

func (h *Handlers) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := int32((page - 1) * limit)
	orgs, total, err := h.svc.ListOrganizations(r.Context(), int32(limit), offset)
	if err != nil {
		h.errorJSON(w, http.StatusInternalServerError, "failed to list organizations")
		return
	}
	h.json(w, http.StatusOK, map[string]any{"organizations": orgs, "total": total, "page": page, "limit": limit})
}

func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		h.errorJSON(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	h.json(w, http.StatusOK, stats)
}

func (h *Handlers) GetSecurityTelemetry(w http.ResponseWriter, r *http.Request) {
	telemetry, err := h.svc.GetSecurityTelemetry(r.Context())
	if err != nil {
		h.errorJSON(w, http.StatusInternalServerError, "failed to get security telemetry")
		return
	}
	h.json(w, http.StatusOK, telemetry)
}

func (h *Handlers) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := int32((page - 1) * limit)
	logs, total, err := h.svc.GetAuditLogs(r.Context(), int32(limit), offset)
	if err != nil {
		h.errorJSON(w, http.StatusInternalServerError, "failed to get audit logs")
		return
	}
	h.json(w, http.StatusOK, map[string]any{"logs": logs, "total": total, "page": page, "limit": limit})
}
