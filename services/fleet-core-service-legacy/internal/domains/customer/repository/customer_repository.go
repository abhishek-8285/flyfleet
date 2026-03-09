package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// -- Customer Methods --

func (r *CustomerRepository) Create(ctx context.Context, c *models.Customer) error {
	query := `
		INSERT INTO customers (
			id, org_id, name, email, phone, api_key, api_secret_hash,
			address, billing_info, status, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	addressJSON, err := json.Marshal(c.Address)
	if err != nil {
		return fmt.Errorf("failed to marshal address: %w", err)
	}
	billingJSON, err := json.Marshal(c.BillingInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal billing info: %w", err)
	}
	metadataJSON, err := json.Marshal(c.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.db.QueryRow(ctx, query,
		id, c.OrgID, c.Name, c.Email, c.Phone, c.APIKey, c.APISecretHash,
		addressJSON, billingJSON, c.Status, metadataJSON, now, now,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *CustomerRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Customer, error) {
	query := `
		SELECT id, org_id, name, email, phone, api_key, api_secret_hash,
			   address, billing_info, status, metadata, created_at, updated_at
		FROM customers WHERE api_key = $1 AND deleted_at IS NULL`

	var c models.Customer
	var addressJSON, billingJSON, metadataJSON []byte

	err := r.db.QueryRow(ctx, query, apiKey).Scan(
		&c.ID, &c.OrgID, &c.Name, &c.Email, &c.Phone, &c.APIKey, &c.APISecretHash,
		&addressJSON, &billingJSON, &c.Status, &metadataJSON, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer by API key: %w", err)
	}

	if err := json.Unmarshal(addressJSON, &c.Address); err != nil {
		return nil, fmt.Errorf("failed to unmarshal address: %w", err)
	}
	if err := json.Unmarshal(billingJSON, &c.BillingInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal billing info: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &c.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &c, nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, id string) (*models.Customer, error) {
	query := `
		SELECT id, org_id, name, email, phone, api_key, api_secret_hash,
			   address, billing_info, status, metadata, created_at, updated_at
		FROM customers WHERE id = $1 AND deleted_at IS NULL`

	var c models.Customer
	var addressJSON, billingJSON, metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.OrgID, &c.Name, &c.Email, &c.Phone, &c.APIKey, &c.APISecretHash,
		&addressJSON, &billingJSON, &c.Status, &metadataJSON, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer by ID: %w", err)
	}

	if err := json.Unmarshal(addressJSON, &c.Address); err != nil {
		return nil, fmt.Errorf("failed to unmarshal address: %w", err)
	}
	if err := json.Unmarshal(billingJSON, &c.BillingInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal billing info: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &c.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &c, nil
}

// -- Shipment Methods --

func (r *CustomerRepository) CreateShipment(ctx context.Context, s *models.Shipment) error {
	query := `
		INSERT INTO shipments (
			id, customer_id, trip_id, tracking_number, status,
			origin, destination, current_location, estimated_delivery,
			actual_delivery, cargo_details, pod_url, pod_uploaded_at,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	originJSON, err := json.Marshal(s.Origin)
	if err != nil {
		return fmt.Errorf("failed to marshal origin: %w", err)
	}
	destJSON, err := json.Marshal(s.Destination)
	if err != nil {
		return fmt.Errorf("failed to marshal destination: %w", err)
	}
	currLocJSON, err := json.Marshal(s.CurrentLocation)
	if err != nil {
		return fmt.Errorf("failed to marshal current location: %w", err)
	}
	cargoJSON, err := json.Marshal(s.CargoDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal cargo details: %w", err)
	}
	metadataJSON, err := json.Marshal(s.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.db.QueryRow(ctx, query,
		id, s.CustomerID, s.TripID, s.TrackingNumber, s.Status,
		originJSON, destJSON, currLocJSON, s.EstimatedDelivery,
		s.ActualDelivery, cargoJSON, s.PODURL, s.PODUploadedAt,
		metadataJSON, now, now,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create shipment: %w", err)
	}

	return nil
}

func (r *CustomerRepository) GetShipment(ctx context.Context, id string) (*models.Shipment, error) {
	query := `
		SELECT id, customer_id, trip_id, tracking_number, status,
			   origin, destination, current_location, estimated_delivery,
			   actual_delivery, cargo_details, pod_url, pod_uploaded_at,
			   metadata, created_at, updated_at
		FROM shipments WHERE id = $1 AND deleted_at IS NULL`

	var s models.Shipment
	var originJSON, destJSON, currLocJSON, cargoJSON, metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.CustomerID, &s.TripID, &s.TrackingNumber, &s.Status,
		&originJSON, &destJSON, &currLocJSON, &s.EstimatedDelivery,
		&s.ActualDelivery, &cargoJSON, &s.PODURL, &s.PODUploadedAt,
		&metadataJSON, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}

	if err := json.Unmarshal(originJSON, &s.Origin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
	}
	if err := json.Unmarshal(destJSON, &s.Destination); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
	}
	if err := json.Unmarshal(currLocJSON, &s.CurrentLocation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal current location: %w", err)
	}
	if err := json.Unmarshal(cargoJSON, &s.CargoDetails); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cargo details: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &s.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &s, nil
}

func (r *CustomerRepository) GetShipmentByTrackingNumber(ctx context.Context, trackingNumber string) (*models.Shipment, error) {
	query := `
		SELECT id, customer_id, trip_id, tracking_number, status,
			   origin, destination, current_location, estimated_delivery,
			   actual_delivery, cargo_details, pod_url, pod_uploaded_at,
			   metadata, created_at, updated_at
		FROM shipments WHERE tracking_number = $1 AND deleted_at IS NULL`

	var s models.Shipment
	var originJSON, destJSON, currLocJSON, cargoJSON, metadataJSON []byte

	err := r.db.QueryRow(ctx, query, trackingNumber).Scan(
		&s.ID, &s.CustomerID, &s.TripID, &s.TrackingNumber, &s.Status,
		&originJSON, &destJSON, &currLocJSON, &s.EstimatedDelivery,
		&s.ActualDelivery, &cargoJSON, &s.PODURL, &s.PODUploadedAt,
		&metadataJSON, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get shipment by tracking number: %w", err)
	}

	if err := json.Unmarshal(originJSON, &s.Origin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
	}
	if err := json.Unmarshal(destJSON, &s.Destination); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
	}
	if err := json.Unmarshal(currLocJSON, &s.CurrentLocation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal current location: %w", err)
	}
	if err := json.Unmarshal(cargoJSON, &s.CargoDetails); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cargo details: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &s.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &s, nil
}

func (r *CustomerRepository) ListShipments(ctx context.Context, filter models.ShipmentListFilter) ([]*models.Shipment, int, error) {
	baseQuery := `
		SELECT id, customer_id, trip_id, tracking_number, status,
			   origin, destination, current_location, estimated_delivery,
			   actual_delivery, cargo_details, pod_url, pod_uploaded_at,
			   metadata, created_at, updated_at
		FROM shipments WHERE deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM shipments WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.CustomerID != "" {
		conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argIndex))
		args = append(args, filter.CustomerID)
		argIndex++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Count
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery + whereClause, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count shipments: %w", err)
	}

	// List
	orderBy := "created_at DESC"
	limit := filter.PageSize
	offset := (filter.Page - 1) * filter.PageSize

	limitClause := fmt.Sprintf(" ORDER BY %s LIMIT $%d OFFSET $%d", orderBy, argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, baseQuery + whereClause + limitClause, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list shipments: %w", err)
	}
	defer rows.Close()

	var shipments []*models.Shipment
	for rows.Next() {
		var s models.Shipment
		var originJSON, destJSON, currLocJSON, cargoJSON, metadataJSON []byte

		err := rows.Scan(
			&s.ID, &s.CustomerID, &s.TripID, &s.TrackingNumber, &s.Status,
			&originJSON, &destJSON, &currLocJSON, &s.EstimatedDelivery,
			&s.ActualDelivery, &cargoJSON, &s.PODURL, &s.PODUploadedAt,
			&metadataJSON, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan shipment: %w", err)
		}

		if err := json.Unmarshal(originJSON, &s.Origin); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal origin: %w", err)
		}
		if err := json.Unmarshal(destJSON, &s.Destination); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal destination: %w", err)
		}
		if err := json.Unmarshal(currLocJSON, &s.CurrentLocation); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal current location: %w", err)
		}
		if err := json.Unmarshal(cargoJSON, &s.CargoDetails); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal cargo details: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &s.Metadata); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		shipments = append(shipments, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("shipment iteration error: %w", err)
	}

	return shipments, totalCount, nil
}

// -- Invoice Methods --

func (r *CustomerRepository) GetInvoices(ctx context.Context, customerID string) ([]*models.Invoice, error) {
	query := `
		SELECT id, customer_id, shipment_id, invoice_number, status,
			   amount, currency, issue_date, due_date, paid_date, pdf_url,
			   metadata, created_at, updated_at
		FROM invoices WHERE customer_id = $1 ORDER BY issue_date DESC`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		var i models.Invoice
		var metadataJSON []byte
		err := rows.Scan(
			&i.ID, &i.CustomerID, &i.ShipmentID, &i.InvoiceNumber, &i.Status,
			&i.Amount, &i.Currency, &i.IssueDate, &i.DueDate, &i.PaidDate,
			&i.PDFURL, &metadataJSON, &i.CreatedAt, &i.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &i.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		invoices = append(invoices, &i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("invoice iteration error: %w", err)
	}

	return invoices, nil
}

// -- Booking Request Methods --

func (r *CustomerRepository) CreateBookingRequest(ctx context.Context, b *models.BookingRequest) error {
	query := `
		INSERT INTO booking_requests (
			id, customer_id, request_type, origin, destination,
			requested_pickup_time, cargo_details, status, notes,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	originJSON, err := json.Marshal(b.Origin)
	if err != nil {
		return fmt.Errorf("failed to marshal origin: %w", err)
	}
	destJSON, err := json.Marshal(b.Destination)
	if err != nil {
		return fmt.Errorf("failed to marshal destination: %w", err)
	}
	cargoJSON, err := json.Marshal(b.CargoDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal cargo details: %w", err)
	}
	metadataJSON, err := json.Marshal(b.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.db.QueryRow(ctx, query,
		id, b.CustomerID, b.RequestType, originJSON, destJSON,
		b.RequestedPickupTime, cargoJSON, b.Status, b.Notes,
		metadataJSON, now, now,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create booking request: %w", err)
	}

	return nil
}
