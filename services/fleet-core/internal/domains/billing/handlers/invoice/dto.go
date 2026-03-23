package invoice

import (
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/domain/invoice"
)

type CreateInvoiceRequest struct {
	CustomerName   string              `json:"customer_name" validate:"required"`
	CustomerEmail  string              `json:"customer_email" validate:"required,email"`
	BillingAddress string              `json:"billing_address"`
	DueDate        time.Time           `json:"due_date" validate:"required"`
	Currency       string              `json:"currency" validate:"required,len=3"`
	Notes          string              `json:"notes"`
	Items          []CreateItemRequest `json:"items" validate:"required,min=1"`
}

type CreateItemRequest struct {
	Description string  `json:"description" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gte=0"`
}

func (r CreateInvoiceRequest) ToDomain() *invoice.Invoice {
	items := make([]invoice.InvoiceItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = invoice.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	return &invoice.Invoice{
		CustomerName:   r.CustomerName,
		CustomerEmail:  r.CustomerEmail,
		BillingAddress: r.BillingAddress,
		DueDate:        r.DueDate,
		Currency:       r.Currency,
		Notes:          r.Notes,
		Items:          items,
	}
}
