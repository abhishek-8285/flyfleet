package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type LogHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewLogHandlers(cfg *config.Config, logger *zap.Logger) *LogHandlers {
	return &LogHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// 1. FETCH LOGS (MAIN TABLE)
func (h *LogHandlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	module := r.URL.Query().Get("module")

	h.logger.Info("Getting troubleshooting logs", zap.String("search", search), zap.String("module", module))

	// Mock logs
	logs := []map[string]interface{}{
		{
			"timestamp": "2023-10-24 14:32:01.452",
			"severity":  "ERROR",
			"module":    "AuthService",
			"message":   "Failed to validate token signature: key expired",
			"traceId":   "tid_8f92a1",
		},
		{
			"timestamp": "2023-10-24 14:31:58.102",
			"severity":  "WARN",
			"module":    "RateLimiter",
			"message":   "User 4921 approaching API limit (95%)",
			"traceId":   "tid_7b12c4",
		},
		{
			"timestamp": "2023-10-24 14:31:55.004",
			"severity":  "INFO",
			"module":    "FleetManager",
			"message":   "Vehicle status updated: V-8821 -> Active",
			"traceId":   "tid_6d88e2",
		},
		{
			"timestamp": "2023-10-24 14:31:52.331",
			"severity":  "INFO",
			"module":    "PaymentGateway",
			"message":   "Webhook received from Stripe: payment_intent.succeeded",
			"traceId":   "tid_5a44f9",
		},
		{
			"timestamp": "2023-10-24 14:31:45.122",
			"severity":  "ERROR",
			"module":    "Database",
			"message":   "Connection timeout: pgbouncer-main-02",
			"traceId":   "tid_4c33d1",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": logs,
		"total": 2492,
	})
}

// 2. SEARCH LOGS
func (h *LogHandlers) SearchLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	h.logger.Info("Searching logs", zap.String("q", q))
	h.GetLogs(w, r)
}

// 3. FILTER METADATA
func (h *LogHandlers) GetLogFilters(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting log filters")

	res := map[string]interface{}{
		"modules":    []string{"AuthService", "FleetAPI", "PaymentGateway", "RateLimiter", "Database", "MapService"},
		"severities": []string{"ERROR", "WARN", "INFO", "DEBUG"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 4. LOG DETAIL
func (h *LogHandlers) GetLogDetail(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("traceId")
	h.logger.Info("Getting log detail", zap.String("traceId", traceID))

	res := map[string]interface{}{
		"traceId":     traceID,
		"message":     "Failed to validate token signature: key expired",
		"service":     "AuthService",
		"environment": "Production",
		"severity":    "ERROR",
		"userId":      "usr_99281",
		"ip":          "192.168.1.42",
		"timestamp":   "2023-10-24 14:32:01.452",
		"stackTrace":  "io.jsonwebtoken.ExpiredJwtException: JWT expired at 2023-10-24T14:30:00Z. Current time: 2023-10-24T14:32:01Z\n    at io.jsonwebtoken.impl.DefaultJwtParser.parse(DefaultJwtParser.java:401)\n    at com.fleet.auth.TokenValidator.validate(TokenValidator.java:45)\n    at com.fleet.auth.AuthFilter.doFilter(AuthFilter.java:28)",
		"requestContext": map[string]interface{}{
			"method": "POST",
			"url":    "/api/v1/vehicles/status",
			"headers": map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
				"User-Agent":    "FleetMobile/2.4.1 (iOS 16.2)",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 6. EXPORT LOGS
func (h *LogHandlers) ExportLogs(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Exporting logs")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=logs_evidence.csv")
	w.Write([]byte("Timestamp,Severity,Module,Message,TraceID\n"))
	w.Write([]byte("2023-10-24 14:32:01.452,ERROR,AuthService,Failed to validate token,tid_8f92a1\n"))
}

// 7. COPY TRACE CONTEXT
func (h *LogHandlers) GetLogContext(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("traceId")
	h.logger.Info("Getting log context", zap.String("traceId", traceID))
	w.Write([]byte(fmt.Sprintf("Context for %s: {env: prod, app: support}", traceID)))
}

// 8. CREATE JIRA TICKET
func (h *LogHandlers) CreateLogJiraTicket(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("traceId")
	h.logger.Info("Creating Jira ticket for log", zap.String("traceId", traceID))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"ticketId": "JIRA-101", "url": "https://jira.fleet.com/browse/JIRA-101"}`))
}

// 10. COUNT SUMMARY / STATS
func (h *LogHandlers) GetLogStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting log stats")
	res := map[string]interface{}{
		"total":    2492,
		"errors":   142,
		"warnings": 311,
		"info":     2039,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
