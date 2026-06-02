package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func GetGateLogsHandler(pool *pgxpool.Pool, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := ctxkeys.GetTenantID(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized - tenant required"}`))
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		offsetStr := r.URL.Query().Get("offset")
		offset := 0
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		query := `
			SELECT 
				g.id, g.tenant_id, g.vehicle_id, g.gate_name, g.gatekeeper_id, g.action_type, g.cargo_status, g.weight_measured, g.created_at,
				COALESCE(v.registration, 'Unknown') AS vehicle_registration
			FROM yard_gate_logs g
			LEFT JOIN vehicles v ON g.vehicle_id = v.id
			WHERE g.tenant_id = $1
			ORDER BY g.created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err := pool.Query(r.Context(), query, tenantID, limit, offset)
		if err != nil {
			logger.Error("Failed to fetch yard gate logs", zap.Error(err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal server error"}`))
			return
		}
		defer rows.Close()

		type GateLogItem struct {
			ID                  string   `json:"id"`
			TenantID            string   `json:"tenant_id"`
			VehicleID           string   `json:"vehicle_id"`
			VehicleRegistration string   `json:"vehicle_registration"`
			GateName            string   `json:"gate_name"`
			GatekeeperID        *string  `json:"gatekeeper_id"`
			ActionType          string   `json:"action_type"`
			CargoStatus         *string  `json:"cargo_status"`
			WeightMeasured      *float64 `json:"weight_measured"`
			CreatedAt           string   `json:"created_at"`
		}

		logs := []GateLogItem{}
		for rows.Next() {
			var item GateLogItem
			var gatekeeperID, cargoStatus *string
			var weightMeasured *float64
			var id, tenantID, vehicleID, vehicleRegistration, gateName, actionType string
			var createdAt time.Time

			err := rows.Scan(
				&id, &tenantID, &vehicleID, &gateName, &gatekeeperID, &actionType, &cargoStatus, &weightMeasured, &createdAt,
				&vehicleRegistration,
			)
			if err != nil {
				logger.Error("Failed to scan yard gate log row", zap.Error(err))
				continue
			}

			item.ID = id
			item.TenantID = tenantID
			item.VehicleID = vehicleID
			item.GateName = gateName
			item.GatekeeperID = gatekeeperID
			item.ActionType = actionType
			item.CargoStatus = cargoStatus
			item.WeightMeasured = weightMeasured
			item.CreatedAt = createdAt.Format(time.RFC3339)
			item.VehicleRegistration = vehicleRegistration

			logs = append(logs, item)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	}
}

func GetFintechTransactionsHandler(pool *pgxpool.Pool, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := ctxkeys.GetTenantID(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized - tenant required"}`))
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		offsetStr := r.URL.Query().Get("offset")
		offset := 0
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		query := `
			SELECT 
				f.id, f.tenant_id, f.vehicle_id, f.service_type, f.transaction_amount, f.commission_earned, f.payment_status, f.created_at,
				COALESCE(v.registration, 'Unknown') AS vehicle_registration
			FROM fintech_transactions f
			LEFT JOIN vehicles v ON f.vehicle_id = v.id
			WHERE f.tenant_id = $1
			ORDER BY f.created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err := pool.Query(r.Context(), query, tenantID, limit, offset)
		if err != nil {
			logger.Error("Failed to fetch fintech transactions", zap.Error(err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal server error"}`))
			return
		}
		defer rows.Close()

		type FintechTxItem struct {
			ID                  string  `json:"id"`
			TenantID            string  `json:"tenant_id"`
			VehicleID           string  `json:"vehicle_id"`
			VehicleRegistration string  `json:"vehicle_registration"`
			ServiceType         string  `json:"service_type"`
			TransactionAmount   float64 `json:"transaction_amount"`
			CommissionEarned    float64 `json:"commission_earned"`
			PaymentStatus       string  `json:"payment_status"`
			CreatedAt           string  `json:"created_at"`
		}

		transactions := []FintechTxItem{}
		for rows.Next() {
			var item FintechTxItem
			var id, tenantID, vehicleID, serviceType, paymentStatus, vehicleRegistration string
			var transactionAmount, commissionEarned float64
			var createdAt time.Time

			err := rows.Scan(
				&id, &tenantID, &vehicleID, &serviceType, &transactionAmount, &commissionEarned, &paymentStatus, &createdAt,
				&vehicleRegistration,
			)
			if err != nil {
				logger.Error("Failed to scan fintech transaction row", zap.Error(err))
				continue
			}

			item.ID = id
			item.TenantID = tenantID
			item.VehicleID = vehicleID
			item.ServiceType = serviceType
			item.TransactionAmount = transactionAmount
			item.CommissionEarned = commissionEarned
			item.PaymentStatus = paymentStatus
			item.CreatedAt = createdAt.Format(time.RFC3339)
			item.VehicleRegistration = vehicleRegistration

			transactions = append(transactions, item)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	}
}
