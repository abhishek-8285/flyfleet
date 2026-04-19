package invoicing

import (
	"time"
)

type CreateInvoiceRequest struct {
	CustomerName   string              `json:"customer_name" validate:"required"`
	CustomerEmail  string              `json:"customer_email" validate:"omitempty,email"`
	CustomerPhone  string              `json:"customer_phone" validate:"required"`
	BillingAddress string              `json:"billing_address" validate:"required"`
	DueDate        time.Time           `json:"due_date" validate:"required"`
	Currency       string              `json:"currency" validate:"required,len=3"`
	Notes          string              `json:"notes"`
	Items          []CreateInvoiceItem `json:"items" validate:"required,min=1"`
}

type CreateInvoiceItem struct {
	Description string  `json:"description" validate:"required"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

func (r CreateInvoiceRequest) ToDomain() *Invoice {
	inv := &Invoice{
		CustomerName:   r.CustomerName,
		CustomerEmail:  r.CustomerEmail,
		CustomerPhone:  r.CustomerPhone,
		BillingAddress: r.BillingAddress,
		DueDate:        r.DueDate,
		Currency:       r.Currency,
		Notes:          r.Notes,
		Status:         StatusDraft,
	}

	for _, item := range r.Items {
		inv.Items = append(inv.Items, InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Amount:      item.Quantity * item.UnitPrice,
		})
	}

	return inv
}
