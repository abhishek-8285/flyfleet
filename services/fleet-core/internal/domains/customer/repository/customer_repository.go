package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/customer/models"
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
	GetShipmentByTripID(ctx context.Context, tripID uuid.UUID) (*models.Shipment, error)
	UpdateShipment(ctx context.Context, s *models.Shipment) error
	CountPendingShipments(ctx context.Context, tenantID uuid.UUID) (int, error)
	GetShipmentTracking(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)
	GetDashboardMetrics(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error)
	GetAnalytics(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error)
	SearchShipments(ctx context.Context, customerID uuid.UUID, filters map[string]interface{}) ([]models.Shipment, error)
	ListCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]models.Customer, error)
	CountCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}) (int, error)
	UpdateCustomer(ctx context.Context, c *models.Customer) error
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
		INSERT INTO customers (
			id, tenant_id, name, gst_number, billing_address, 
			contact_email, contact_phone, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	id := uuid.New()
	now := time.Now()

	var billingAddrJSON []byte
	if c.BillingAddress != "" {
		// If it doesn't look like JSON, wrap it in quotes to make it a valid JSON string
		addr := c.BillingAddress
		if !((len(addr) > 0 && addr[0] == '{') || (len(addr) > 0 && addr[0] == '[')) {
			// It's a plain string, JSON-quote it
			b, err := json.Marshal(addr)
			if err == nil {
				billingAddrJSON = b
			} else {
				billingAddrJSON = []byte(addr)
			}
		} else {
			billingAddrJSON = []byte(addr)
		}
	} else {
		billingAddrJSON = []byte("{}")
	}

	err := r.db.QueryRow(ctx, query,
		id, c.TenantID, c.Name, c.GSTNumber, billingAddrJSON,
		c.ContactEmail, c.ContactPhone, c.Status, now, now,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r customerV2Repository) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	query := `
		SELECT id, tenant_id, name, gst_number, billing_address,
			   contact_email, contact_phone, status, created_at, updated_at
		FROM customers WHERE id = $1`

	var c models.Customer
	var billingAddrJSON []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.GSTNumber, &billingAddrJSON,
		&c.ContactEmail, &c.ContactPhone, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" || err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get customer by ID: %w", err)
	}

	// Handle JSON decoding for address
	if len(billingAddrJSON) > 0 {
		if billingAddrJSON[0] == '"' {
			var s string
			if err := json.Unmarshal(billingAddrJSON, &s); err == nil {
				c.BillingAddress = s
			} else {
				c.BillingAddress = string(billingAddrJSON)
			}
		} else {
			c.BillingAddress = string(billingAddrJSON)
		}
	}

	return &c, nil
}

func (r *customerV2Repository) ListCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]models.Customer, error) {
	query := `
		SELECT id, tenant_id, name, gst_number, billing_address,
			   contact_email, contact_phone, status, created_at, updated_at
		FROM customers 
		WHERE tenant_id = $1`

	args := []interface{}{tenantID}
	argIdx := 2

	if name, ok := filters["name"].(string); ok && name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argIdx)
		args = append(args, "%"+name+"%")
		argIdx++
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		var billingAddrJSON []byte
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.GSTNumber, &billingAddrJSON,
			&c.ContactEmail, &c.ContactPhone, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}

		// Handle JSON decoding for address
		if len(billingAddrJSON) > 0 {
			if billingAddrJSON[0] == '"' {
				var s string
				if err := json.Unmarshal(billingAddrJSON, &s); err == nil {
					c.BillingAddress = s
				} else {
					c.BillingAddress = string(billingAddrJSON)
				}
			} else {
				c.BillingAddress = string(billingAddrJSON)
			}
		}

		c.NormalizeStatus()
		customers = append(customers, c)
	}

	return customers, nil
}

func (r *customerV2Repository) CountCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}) (int, error) {
	query := `SELECT COUNT(*) FROM customers WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIdx := 2

	if name, ok := filters["name"].(string); ok && name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argIdx)
		args = append(args, "%"+name+"%")
		argIdx++
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count customers: %w", err)
	}
	return count, nil
}

func (r *customerV2Repository) UpdateCustomer(ctx context.Context, c *models.Customer) error {
	query := `
		UPDATE customers SET 
			name = $1, gst_number = $2, billing_address = $3, 
			contact_email = $4, contact_phone = $5, status = $6, updated_at = $7
		WHERE id = $8 AND tenant_id = $9`

	now := time.Now()
	var billingAddrJSON []byte
	if c.BillingAddress != "" {
		addr := c.BillingAddress
		if !((len(addr) > 0 && addr[0] == '{') || (len(addr) > 0 && addr[0] == '[')) {
			b, err := json.Marshal(addr)
			if err == nil {
				billingAddrJSON = b
			} else {
				billingAddrJSON = []byte(addr)
			}
		} else {
			billingAddrJSON = []byte(addr)
		}
	} else {
		billingAddrJSON = []byte("{}")
	}

	cmd, err := r.db.Exec(ctx, query,
		c.Name, c.GSTNumber, billingAddrJSON,
		c.ContactEmail, c.ContactPhone, c.Status, now,
		c.ID, c.TenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("customer not found or access denied")
	}

	return nil
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

// GetShipmentByTripID retrieves a shipment by its linked trip ID.
func (r customerV2Repository) GetShipmentByTripID(ctx context.Context, tripID uuid.UUID) (*models.Shipment, error) {
	query := `
		SELECT id, customer_id, trip_id, booking_id, status, origin,
			   destination, distance_km, created_at, updated_at
		FROM customer_v2_shipments WHERE trip_id = $1`

	var s models.Shipment
	err := r.db.QueryRow(ctx, query, tripID).Scan(
		&s.ID, &s.CustomerID, &s.TripID, &s.BookingID, &s.Status, &s.Origin,
		&s.Destination, &s.DistanceKm, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" || err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get shipment by trip ID: %w", err)
	}
	return &s, nil
}

// UpdateShipment updates the trip_id, status, and timestamps of a shipment.
func (r customerV2Repository) UpdateShipment(ctx context.Context, s *models.Shipment) error {
	query := `
		UPDATE customer_v2_shipments
		   SET trip_id = $1, status = $2, updated_at = NOW()
		 WHERE id = $3`

	cmd, err := r.db.Exec(ctx, query, s.TripID, s.Status, s.ID)
	if err != nil {
		return fmt.Errorf("failed to update shipment: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("shipment not found or update affected 0 rows")
	}
	return nil
}

// CountPendingShipments returns the number of shipments in CONFIRMED or BOOKED
// status that are not yet linked to a trip, for a given tenant.
func (r customerV2Repository) CountPendingShipments(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM customer_v2_shipments s
		JOIN customers c ON s.customer_id = c.id
		WHERE c.tenant_id = $1
		  AND s.status IN ('CONFIRMED', 'BOOKED')
		  AND s.trip_id IS NULL`

	var count int
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending shipments: %w", err)
	}
	return count, nil
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
	// 1. Shipment Counts
	var activeShipments, deliveredShipments, totalShipments int
	
	// Check ACTIVE shipments (not delivered/cancelled)
	err := r.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE status NOT IN ('DELIVERED', 'CANCELLED')),
			COUNT(*) FILTER (WHERE status = 'DELIVERED'),
			COUNT(*)
		FROM customer_v2_shipments 
		WHERE customer_id = $1
	`, customerID).Scan(&activeShipments, &deliveredShipments, &totalShipments)
	
	if err != nil {
		// Log and default to 0 if table missing/error
		fmt.Printf("Warning: failed to query shipments metrics: %v\n", err)
		activeShipments = 0
		deliveredShipments = 0
		totalShipments = 0
	}

	// 2. Pending bookings count
	var pendingBookings int
	err = r.db.QueryRow(ctx, "SELECT count(*) FROM customer_v2_booking_requests WHERE customer_id = $1 AND status = 'REQUESTED'", customerID).Scan(&pendingBookings)
	if err != nil {
		fmt.Printf("Warning: failed to query bookings metrics: %v\n", err)
		pendingBookings = 0
	}

	// 3. Revenue
	var totalRevenue float64
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(q.total_amount), 0)
		FROM customer_v2_pricing_quotes q
		JOIN customer_v2_booking_requests b ON q.booking_id = b.id
		WHERE b.customer_id = $1
	`, customerID).Scan(&totalRevenue)
	if err != nil {
		fmt.Printf("Warning: failed to query revenue metrics: %v\n", err)
		totalRevenue = 0
	}

	// 4. Unpaid Invoices (using credit usage as proxy)
	var unpaidInvoices float64
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(credit_limit - available_credit, 0)
		FROM customer_v2_customer_credit_limits
		WHERE customer_id = $1
	`, customerID).Scan(&unpaidInvoices)
	if err != nil {
		fmt.Printf("Warning: failed to query invoice metrics: %v\n", err)
		unpaidInvoices = 0
	}

	return map[string]interface{}{
		"active_shipments":    activeShipments,
		"delivered_shipments": deliveredShipments,
		"total_shipments":     totalShipments,
		"pending_bookings":    pendingBookings,
		"total_revenue":       totalRevenue,
		"unpaid_invoices":     unpaidInvoices, 
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
