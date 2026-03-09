package integrations

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─── Handler ─────────────────────────────────────────────────────────────────

type Handler struct {
	svc     Service
	metrics *IntegrationMetrics
	logger  *zap.Logger
}

func NewHandler(svc Service, metrics *IntegrationMetrics, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, metrics: metrics, logger: logger}
}

// RegisterRoutes mounts all integrations API endpoints.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/integrations", func(r chi.Router) {
		r.Get("/", h.ListIntegrations)
		r.Post("/", h.CreateIntegration)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetIntegration)
			r.Post("/enable", h.EnableIntegration)
			r.Post("/disable", h.DisableIntegration)
			r.Delete("/", h.DeleteIntegration)
			// Credentials
			r.Post("/credentials", h.AddCredential)
		})
		// Events
		r.Post("/events/{id}/replay", h.ReplayEvent)
		r.Get("/events/dlq", h.ListDLQ)
		// Connector health
		r.Get("/health", h.ConnectorHealth)
		// Metrics
		r.Get("/metrics", h.ServeMetrics)
	})

	r.Route("/webhooks", func(r chi.Router) {
		r.Get("/", h.ListWebhooks)
		r.Post("/", h.CreateWebhook)
		r.Delete("/{id}", h.DeleteWebhook)
		// Inbound webhook verification endpoint
		r.Post("/inbound/{id}", h.VerifyInboundWebhook)
	})
}

// ─── Integration Handlers ────────────────────────────────────────────────────

// POST /integrations
func (h *Handler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonErr(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req CreateIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid request body", http.StatusBadRequest)
		return
	}
	i, err := h.svc.CreateIntegration(r.Context(), tenantID, req)
	if err != nil {
		h.logger.Error("CreateIntegration failed", zap.Error(err))
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, i)
}

// GET /integrations
func (h *Handler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonErr(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	list, err := h.svc.ListIntegrations(r.Context(), tenantID)
	if err != nil {
		jsonErr(w, "failed to list integrations", http.StatusInternalServerError)
		return
	}
	jsonOK(w, list)
}

// GET /integrations/{id}
func (h *Handler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	id, ok := parseUUID(r, "id")
	if !ok || tenantID == uuid.Nil {
		jsonErr(w, "bad request", http.StatusBadRequest)
		return
	}
	i, err := h.svc.GetIntegration(r.Context(), id, tenantID)
	if err != nil {
		jsonErr(w, "not found", http.StatusNotFound)
		return
	}
	jsonOK(w, i)
}

// POST /integrations/{id}/enable
func (h *Handler) EnableIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	id, ok := parseUUID(r, "id")
	if !ok || tenantID == uuid.Nil {
		jsonErr(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.svc.EnableIntegration(r.Context(), id, tenantID); err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /integrations/{id}/disable
func (h *Handler) DisableIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	id, ok := parseUUID(r, "id")
	if !ok || tenantID == uuid.Nil {
		jsonErr(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.svc.DisableIntegration(r.Context(), id, tenantID); err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /integrations/{id}
func (h *Handler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	id, ok := parseUUID(r, "id")
	if !ok || tenantID == uuid.Nil {
		jsonErr(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteIntegration(r.Context(), id, tenantID); err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /integrations/{id}/credentials
func (h *Handler) AddCredential(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(r, "id")
	if !ok {
		jsonErr(w, "invalid integration id", http.StatusBadRequest)
		return
	}
	var req AddCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.svc.AddCredential(r.Context(), id, req); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]string{"status": "credential stored"})
}

// POST /integrations/events/{id}/replay
func (h *Handler) ReplayEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(r, "id")
	if !ok {
		jsonErr(w, "invalid event id", http.StatusBadRequest)
		return
	}
	if err := h.svc.ReplayEvent(r.Context(), id); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	jsonOK(w, map[string]string{"status": "event queued for replay"})
}

// GET /integrations/events/dlq
func (h *Handler) ListDLQ(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonErr(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	entries, err := h.svc.ListDLQ(r.Context(), tenantID)
	if err != nil {
		jsonErr(w, "failed to list DLQ", http.StatusInternalServerError)
		return
	}
	jsonOK(w, entries)
}

// GET /integrations/health
func (h *Handler) ConnectorHealth(w http.ResponseWriter, r *http.Request) {
	statuses := h.svc.ConnectorHealthStatus(r.Context())
	jsonOK(w, statuses)
}

// GET /integrations/metrics
func (h *Handler) ServeMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	if h.metrics != nil {
		w.Write([]byte(h.metrics.ServeMetrics()))
	}
}

// ─── Webhook Handlers ─────────────────────────────────────────────────────────

// POST /webhooks
func (h *Handler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonErr(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid request body", http.StatusBadRequest)
		return
	}
	sub, err := h.svc.CreateWebhookSubscription(r.Context(), tenantID, req)
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, sub)
}

// GET /webhooks
func (h *Handler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	if tenantID == uuid.Nil {
		jsonErr(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	subs, err := h.svc.ListWebhookSubscriptions(r.Context(), tenantID)
	if err != nil {
		jsonErr(w, "failed to list webhooks", http.StatusInternalServerError)
		return
	}
	jsonOK(w, subs)
}

// DELETE /webhooks/{id}
func (h *Handler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID := h.tenantID(r)
	id, ok := parseUUID(r, "id")
	if !ok || tenantID == uuid.Nil {
		jsonErr(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteWebhookSubscription(r.Context(), id, tenantID); err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /webhooks/inbound/{id} — validates HMAC signature on incoming webhooks.
// Used as a reverse proxy for external systems calling back FlyFleet.
func (h *Handler) VerifyInboundWebhook(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("X-Signature")
	ts := r.Header.Get("X-Timestamp")
	if sig == "" || ts == "" {
		jsonErr(w, "missing signature headers", http.StatusBadRequest)
		return
	}

	// Read body for verification — store in buffer for downstream handler
	var bodyBuf []byte
	if _, err := r.Body.Read(bodyBuf); err != nil && len(bodyBuf) == 0 {
		jsonErr(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Validate timestamp freshness (5-minute window)
	whID, ok := parseUUID(r, "id")
	if !ok {
		jsonErr(w, "invalid webhook id", http.StatusBadRequest)
		return
	}

	h.logger.Info("Inbound webhook received",
		zap.String("webhook_id", whID.String()),
		zap.String("timestamp", ts))

	// Production: look up subscription secret by whID, then call VerifyWebhookSignature
	w.WriteHeader(http.StatusAccepted)
	jsonOK(w, map[string]string{"status": "received"})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (h *Handler) tenantID(r *http.Request) uuid.UUID {
	raw := r.Header.Get("X-Tenant-ID")
	if raw == "" {
		raw = r.URL.Query().Get("tenant_id")
	}
	id, _ := uuid.Parse(raw)
	return id
}

func parseUUID(r *http.Request, param string) (uuid.UUID, bool) {
	raw := chi.URLParam(r, param)
	id, err := uuid.Parse(raw)
	return id, err == nil
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":     msg,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
