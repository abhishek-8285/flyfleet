package repository

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/invoicing"
	"github.com/google/uuid"
)

type InvoiceRepository interface {
	Create(ctx context.Context, inv *invoicing.Invoice) (*invoicing.Invoice, error)
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*invoicing.Invoice, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]invoicing.Invoice, int64, error)
	UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status invoicing.InvoiceStatus, updatedBy uuid.UUID) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}
