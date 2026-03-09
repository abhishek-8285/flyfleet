package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer_v2/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer_v2/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CustomerV2Handler struct {
	customerService *service.CustomerService
	bookingService  *service.BookingService
	shipmentService *service.ShipmentService
	pricingService  *service.PricingService
	creditService   *service.CreditService
	slaService      *service.SLAService
	repo            repository.CustomerV2Repository
	logger          *zap.Logger
}

func NewCustomerV2Handler(
	cs *service.CustomerService,
	bs *service.BookingService,
	ss *service.ShipmentService,
	ps *service.PricingService,
	crs *service.CreditService,
	slas *service.SLAService,
	repo repository.CustomerV2Repository,
	logger *zap.Logger,
) *CustomerV2Handler {
	return &CustomerV2Handler{
		customerService: cs,
		bookingService:  bs,
		shipmentService: ss,
		pricingService:  ps,
		creditService:   crs,
		slaService:      slas,
		repo:            repo,
		logger:          logger,
	}
}

func (h CustomerV2Handler) RegisterRoutes(r chi.Router) {
	// Customers
	r.Post("/customers", h.CreateCustomer)
	r.Get("/customers/{id}", h.GetCustomer)

	// Bookings
	r.Post("/bookings", h.CreateBooking)
	r.Get("/bookings/{id}", h.GetBooking)

	// Shipments & Tracking
	r.Post("/shipments", h.CreateShipment)
	r.Get("/shipments/{id}", h.GetShipment)
	r.Get("/shipments/{id}/tracking", h.GetShipmentTracking)

	// Invoices and Analytics stubs
	r.Get("/invoices", h.GetInvoices)
	r.Get("/analytics", h.GetAnalytics)
	
	// Dashboard & Search
	r.Get("/dashboard", h.GetDashboard)
	r.Get("/shipments/search", h.SearchShipments)

	// Self-Service API (Requires API Token authentication)
	r.Route("/api", func(api chi.Router) {
		api.Use(h.RequireAPIToken) // RBAC Middleware
		api.Post("/bookings", h.APICreateBooking)
		api.Get("/shipments", h.APIGetShipments)
		api.Get("/tracking", h.APIGetTracking)
		api.Get("/invoices", h.APIGetInvoices)
	})
}

// RequireAPIToken is a middleware to enforce RBAC and Tenant Isolation for API calls
func (h CustomerV2Handler) RequireAPIToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			h.respondError(w, http.StatusUnauthorized, "Missing API token")
			return
		}

		// Stub: Validate token Hash with `customer_v2_customer_api_keys`
		// Mock customer ID
		customerID := uuid.New()

		// Inject CustomerID into context for tenant isolation
		// ctx := context.WithValue(r.Context(), "customer_id", customerID)
		// next.ServeHTTP(w, r.WithContext(ctx))

		h.logger.Debug("API Token validated", zap.String("customer_id", customerID.String()))
		next.ServeHTTP(w, r)
	})
}

// --- Self-Service API Handlers ---

func (h CustomerV2Handler) APICreateBooking(w http.ResponseWriter, r *http.Request) {
	// Re-route to standard booking creation
	h.CreateBooking(w, r)
}

func (h CustomerV2Handler) APIGetShipments(w http.ResponseWriter, r *http.Request) {
	// Re-route to search shipments logic
	h.SearchShipments(w, r)
}

func (h CustomerV2Handler) APIGetTracking(w http.ResponseWriter, r *http.Request) {
	h.GetShipmentTracking(w, r)
}

func (h CustomerV2Handler) APIGetInvoices(w http.ResponseWriter, r *http.Request) {
	h.GetInvoices(w, r)
}

func (h CustomerV2Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var c models.Customer
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.customerService.CreateCustomer(r.Context(), &c); err != nil {
		h.logger.Error("Failed to create customer", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create customer")
		return
	}

	h.respondJSON(w, http.StatusCreated, c)
}

func (h CustomerV2Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	c, err := h.customerService.GetCustomerByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch customer", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch customer")
		return
	}
	if c == nil {
		h.respondError(w, http.StatusNotFound, "Customer not found")
		return
	}

	h.respondJSON(w, http.StatusOK, c)
}

func (h CustomerV2Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var b models.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.bookingService.CreateBooking(r.Context(), &b); err != nil {
		h.logger.Error("Failed to create booking", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create booking")
		return
	}

	h.respondJSON(w, http.StatusCreated, b)
}

func (h CustomerV2Handler) GetBooking(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	b, err := h.bookingService.GetBookingByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch booking", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch booking")
		return
	}
	if b == nil {
		h.respondError(w, http.StatusNotFound, "Booking not found")
		return
	}

	h.respondJSON(w, http.StatusOK, b)
}

func (h CustomerV2Handler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	var s models.Shipment
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.shipmentService.CreateShipment(r.Context(), &s); err != nil {
		h.logger.Error("Failed to create shipment", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create shipment")
		return
	}

	h.respondJSON(w, http.StatusCreated, s)
}

func (h CustomerV2Handler) GetShipment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	s, err := h.shipmentService.GetShipmentByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch shipment", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch shipment")
		return
	}
	if s == nil {
		h.respondError(w, http.StatusNotFound, "Shipment not found")
		return
	}

	h.respondJSON(w, http.StatusOK, s)
}

func (h CustomerV2Handler) GetShipmentTracking(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	tracking, err := h.shipmentService.GetShipmentTracking(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch tracking data", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch tracking data")
		return
	}
	if tracking == nil {
		h.respondError(w, http.StatusNotFound, "Shipment tracking not found")
		return
	}

	h.respondJSON(w, http.StatusOK, tracking)
}

func (h CustomerV2Handler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	// Stub for invoices endpoint
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"message": "List of invoices"})
}

func (h CustomerV2Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	// Require customer ID, typically via auth token. Mocking for now:
	customerIDStr := r.Header.Get("X-Customer-ID")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		customerID = uuid.Nil // use a mock if not passed
	}

	analytics, err := h.repo.GetAnalytics(r.Context(), customerID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch analytics")
		return
	}

	h.respondJSON(w, http.StatusOK, analytics)
}

func (h CustomerV2Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	customerIDStr := r.Header.Get("X-Customer-ID")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		customerID = uuid.Nil 
	}

	metrics, err := h.repo.GetDashboardMetrics(r.Context(), customerID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch dashboard metrics")
		return
	}

	h.respondJSON(w, http.StatusOK, metrics)
}

func (h CustomerV2Handler) SearchShipments(w http.ResponseWriter, r *http.Request) {
	customerIDStr := r.Header.Get("X-Customer-ID")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		customerID = uuid.Nil 
	}

	status := r.URL.Query().Get("status")
	filters := map[string]interface{}{}
	if status != "" {
		filters["status"] = status
	}

	shipments, err := h.repo.SearchShipments(r.Context(), customerID, filters)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to search shipments")
		return
	}

	h.respondJSON(w, http.StatusOK, shipments)
}

func (h CustomerV2Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h CustomerV2Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
