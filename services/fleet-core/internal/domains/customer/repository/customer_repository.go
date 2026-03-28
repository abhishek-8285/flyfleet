package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerV2Repository interface {
	CreateCustomer(ctx context.Context, c *models.Customer) error
	GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error)
	CreateBookingRequest(ctx context.Context, b *models.BookingRequest) error
	GetBookingRequestByID(ctx context.Context, id uuid.UUID) (*models.BookingRequest, error)
	CreateShipment(ctx context.Context, s *models.Shipment) error
	GetShipmentByID(ctx context.Context, id uuid.UUID) (*models.Shipment, error)
	GetShipmentTracking(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)
	GetDashboardMetrics(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error)
	GetAnalytics(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error)
	SearchShipments(ctx context.Context, customerID uuid.UUID, filters map[string]interface{}) ([]models.Shipment, error)
}

type customerV2Repository struct {
	db *pgxpool.Pool
}

func NewCustomerV2Repository(db *pgxpool.Pool) CustomerV2Repository {
	return &customerV2Repository{db: db}
}

// -- Customer Methods --

func (r *customerV2Repository) CreateCustomer(ctx context.Context, c *models.Customer) error {
	query := `
		INSERT INTO customer_v2_customers (
			id, organization_id, company_name, gst_number, billing_address, 
			contact_email, contact_phone, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	id := uuid.New()
	now := time.Now()

	var billingAddrJSON []byte
	var err error
	if len(c.BillingAddress) > 0 {
		billingAddrJSON = c.BillingAddress
	} else {
		billingAddrJSON = []byte("{}")
	}

	err = r.db.QueryRow(ctx, query,
		id, c.OrganizationID, c.CompanyName, c.GSTNumber, billingAddrJSON,
		c.ContactEmail, c.ContactPhone, c.Status, now, now,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer_v2 customer: %w", err)
	}

	return nil
}

func (r customerV2Repository) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	query := `
		SELECT id, organization_id, company_name, gst_number, billing_address,
			   contact_email, contact_phone, status, created_at, updated_at
		FROM customer_v2_customers WHERE id = $1`

	var c models.Customer
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.OrganizationID, &c.CompanyName, &c.GSTNumber, &c.BillingAddress,
		&c.ContactEmail, &c.ContactPhone, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" || err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get customer by ID: %w", err)
	}

	return &c, nil
}

// -- Booking Request Methods --

func (r customerV2Repository) CreateBookingRequest(ctx context.Context, b *models.BookingRequest) error {
	query := `
		INSERT INTO customer_v2_booking_requests (
			id, customer_id, pickup_location, drop_location, cargo_type,
			weight, volume, vehicle_type, requested_pickup_time, status,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	id := uuid.New()
	now := time.Now()

	err := r.db.QueryRow(ctx, query,
		id, b.CustomerID, b.PickupLocation, b.DropLocation, b.CargoType,
		b.Weight, b.Volume, b.VehicleType, b.RequestedPickupTime, b.Status,
		now, now,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create booking request: %w", err)
	}

	return nil
}

func (r customerV2Repository) GetBookingRequestByID(ctx context.Context, id uuid.UUID) (*models.BookingRequest, error) {
	query := `
		SELECT id, customer_id, pickup_location, drop_location, cargo_type,
			   weight, volume, vehicle_type, requested_pickup_time, status,
			   created_at, updated_at
		FROM customer_v2_booking_requests WHERE id = $1`

	var b models.BookingRequest
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.CustomerID, &b.PickupLocation, &b.DropLocation, &b.CargoType,
		&b.Weight, &b.Volume, &b.VehicleType, &b.RequestedPickupTime, &b.Status,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" || err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get booking request by ID: %w", err)
	}

	return &b, nil
}

// -- Shipment Methods --

func (r customerV2Repository) CreateShipment(ctx context.Context, s *models.Shipment) error {
	query := `
		INSERT INTO customer_v2_shipments (
			id, customer_id, trip_id, booking_id, status, origin,
			destination, distance_km, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	id := uuid.New()
	now := time.Now()

	err := r.db.QueryRow(ctx, query,
		id, s.CustomerID, s.TripID, s.BookingID, s.Status, s.Origin,
		s.Destination, s.DistanceKm, now, now,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create shipment: %w", err)
	}

	return nil
}

func (r customerV2Repository) GetShipmentByID(ctx context.Context, id uuid.UUID) (*models.Shipment, error) {
	query := `
		SELECT id, customer_id, trip_id, booking_id, status, origin,
			   destination, distance_km, created_at, updated_at
		FROM customer_v2_shipments WHERE id = $1`

	var s models.Shipment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.CustomerID, &s.TripID, &s.BookingID, &s.Status, &s.Origin,
		&s.Destination, &s.DistanceKm, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" || err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get shipment by ID: %w", err)
	}

	return &s, nil
}

// -- Shipment Tracking Method (simplified) --
func (r customerV2Repository) GetShipmentTracking(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	// The tracking_v2 telemetry data logic will enrich this in the service layer,
	// but we fetch the base shipment data here.
	shipment, err := r.GetShipmentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if shipment == nil {
		return nil, nil
	}

	// Let's also fetch events or stops
	eventsQuery := `
		SELECT event_type, timestamp, location, metadata 
		FROM customer_v2_shipment_events 
		WHERE shipment_id = $1 
		ORDER BY timestamp DESC`

	rows, err := r.db.Query(ctx, eventsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var eventType string
		var timestamp time.Time
		var location, metadata []byte

		if err := rows.Scan(&eventType, &timestamp, &location, &metadata); err != nil {
			continue // skip on error
		}

		var locMap, metaMap map[string]interface{}
		json.Unmarshal(location, &locMap)
		json.Unmarshal(metadata, &metaMap)

		events = append(events, map[string]interface{}{
			"event_type": eventType,
			"timestamp":  timestamp,
			"location":   locMap,
			"metadata":   metaMap,
		})
	}

	result := map[string]interface{}{
		"shipment_id": id,
		"status":      shipment.Status,
		"events":      events,
		// to be enriched with telemetry tracking by calling tracking_v2 service
	}

	return result, nil
}

// -- Dashboard & Analytics Methods --

func (r customerV2Repository) GetDashboardMetrics(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) {
	// Active shipments count
	var activeShipments int
	err := r.db.QueryRow(ctx, "SELECT count(*) FROM customer_v2_shipments WHERE customer_id = $1 AND status NOT IN ('DELIVERED', 'CANCELLED')", customerID).Scan(&activeShipments)
	if err != nil {
		return nil, err
	}

	// Delivered shipments count
	var deliveredShipments int
	err = r.db.QueryRow(ctx, "SELECT count(*) FROM customer_v2_shipments WHERE customer_id = $1 AND status = 'DELIVERED'", customerID).Scan(&deliveredShipments)
	if err != nil {
		return nil, err
	}

	// Pending bookings count
	var pendingBookings int
	err = r.db.QueryRow(ctx, "SELECT count(*) FROM customer_v2_booking_requests WHERE customer_id = $1 AND status = 'REQUESTED'", customerID).Scan(&pendingBookings)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"active_shipments":    activeShipments,
		"delivered_shipments": deliveredShipments,
		"pending_bookings":    pendingBookings,
	}, nil
}

func (r customerV2Repository) GetAnalytics(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) {
	// Stub for rich analytics. Represents the metrics the user wants:
	// delivery success rate, average delivery time, monthly shipment volume, customer spend

	return map[string]interface{}{
		"delivery_success_rate":   98.5,
		"average_delivery_time":   "24h 15m",
		"monthly_shipment_volume": 1420,
		"total_customer_spend":    1500000.0,
	}, nil
}

func (r customerV2Repository) SearchShipments(ctx context.Context, customerID uuid.UUID, filters map[string]interface{}) ([]models.Shipment, error) {
	query := `
		SELECT id, customer_id, trip_id, booking_id, status, origin,
			   destination, distance_km, created_at, updated_at
		FROM customer_v2_shipments WHERE customer_id = $1`

	// Simplistic filter application
	args := []interface{}{customerID}

	if status, ok := filters["status"]; ok && status != "" {
		args = append(args, status)
		query += fmt.Sprintf(" AND status = $%d", len(args))
	}

	query += " ORDER BY created_at DESC LIMIT 50"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search shipments: %w", err)
	}
	defer rows.Close()

	var shipments []models.Shipment
	for rows.Next() {
		var s models.Shipment
		if err := rows.Scan(
			&s.ID, &s.CustomerID, &s.TripID, &s.BookingID, &s.Status, &s.Origin,
			&s.Destination, &s.DistanceKm, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			continue
		}
		shipments = append(shipments, s)
	}

	return shipments, nil
}
