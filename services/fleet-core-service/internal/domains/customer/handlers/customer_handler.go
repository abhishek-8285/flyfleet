package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type CustomerHandlers struct {
	service   *service.CustomerService
	logger    *zap.Logger
	validator *validator.Validate
	jwtSecret []byte
}

func NewCustomerHandlers(service *service.CustomerService, logger *zap.Logger, jwtSecret string) *CustomerHandlers {
	return &CustomerHandlers{
		service:   service,
		logger:    logger,
		validator: validator.New(),
		jwtSecret: []byte(jwtSecret),
	}
}

func (h *CustomerHandlers) RegisterRoutes(r chi.Router) {
	// Public Routes
	r.Post("/auth/token", h.Authenticate)
	r.Get("/shipments/{tracking_number}", h.TrackShipment) // Public tracking

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(h.CustomerAuthMiddleware)
		r.Post("/bookings", h.CreateBooking)
		r.Get("/invoices", h.GetInvoices)
	})
}

// Authenticate: POST /auth/token
func (h *CustomerHandlers) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey string `json:"api_key" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, customer, err := h.service.Authenticate(r.Context(), req.APIKey)
	if err != nil {
		h.logger.Warn("auth failure", zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"customer": map[string]string{
			"id":   customer.ID,
			"name": customer.Name,
		},
	})
}

// TrackShipment: GET /shipments/{tracking_number}
func (h *CustomerHandlers) TrackShipment(w http.ResponseWriter, r *http.Request) {
	trackingNumber := chi.URLParam(r, "tracking_number")
	
	shipment, err := h.service.TrackShipment(r.Context(), trackingNumber)
	if err != nil {
		h.logger.Error("failed to track shipment", zap.Error(err))
		http.Error(w, "Shipment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": shipment,
	})
}

// CreateBooking: POST /bookings
func (h *CustomerHandlers) CreateBooking(w http.ResponseWriter, r *http.Request) {
	customerID := r.Context().Value("customer_id").(string)

	var req models.BookingRequestCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	booking, err := h.service.CreateBooking(r.Context(), customerID, &req)
	if err != nil {
		h.logger.Error("failed to create booking", zap.Error(err))
		http.Error(w, "Failed to create booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": booking,
	})
}

// GetInvoices: GET /invoices
func (h *CustomerHandlers) GetInvoices(w http.ResponseWriter, r *http.Request) {
	customerID := r.Context().Value("customer_id").(string)

	invoices, err := h.service.GetCustomerInvoices(r.Context(), customerID)
	if err != nil {
		h.logger.Error("failed to get invoices", zap.Error(err))
		http.Error(w, "Failed to retrieve invoices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": invoices,
	})
}

// CustomerAuthMiddleware validates the JWT
func (h *CustomerHandlers) CustomerAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return h.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		customerID, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Invalid customer ID in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "customer_id", customerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
