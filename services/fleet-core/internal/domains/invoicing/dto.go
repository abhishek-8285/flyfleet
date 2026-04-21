package invoicing

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type JSONDate time.Time

func (j *JSONDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*j = JSONDate(t)
	return nil
}

func (j JSONDate) ToTime() time.Time {
	return time.Time(j)
}

type CreateInvoiceRequest struct {
	CustomerName   string              `json:"customer_name" validate:"required"`
	CustomerEmail  string              `json:"customer_email" validate:"omitempty,email"`
	CustomerPhone  string              `json:"customer_phone" validate:"required,min=10"`
	BillingAddress string              `json:"billing_address" validate:"required"`
	InvoiceDate    JSONDate            `json:"invoice_date" validate:"required"`
	DueDate        JSONDate            `json:"due_date" validate:"required"`
	Currency       string              `json:"currency" validate:"required"`
	Notes          string              `json:"notes"`
	TemplateType   string              `json:"template_type" validate:"omitempty,oneof=TAX_INVOICE TRIP_NOTE"`
	
	// Transporter Details
	TripID         *uuid.UUID `json:"trip_id"`
	VehicleNo      string     `json:"vehicle_no"`
	DriverName     string     `json:"driver_name"`
	DriverMobile   string     `json:"driver_mobile"`
	PickupLocation string     `json:"pickup_location"`
	DropLocation   string     `json:"drop_location"`
	BilledDistance float64    `json:"billed_distance"`

	// Compliance
	PlaceOfSupply string `json:"place_of_supply"`
	ReverseCharge bool   `json:"reverse_charge"`
	EWayBillNo    string `json:"eway_bill_no"`
	IRN           string `json:"irn"`

	// Bank
	BankName    string `json:"bank_name"`
	AccountNo   string `json:"account_no"`
	IFSC        string `json:"ifsc_code"`
	Branch      string `json:"branch_name"`
	Beneficiary string `json:"beneficiary_name"`

	Items []CreateInvoiceItem `json:"items" validate:"required,min=1"`
}

type CreateInvoiceItem struct {
	Description    string  `json:"description" validate:"required"`
	Quantity       float64 `json:"quantity" validate:"required,gt=0"`
	UnitPrice      float64 `json:"unit_price" validate:"required,gt=0"`
	TaxPercent     float64 `json:"tax_percent"`
	HSNSAC         string  `json:"hsn_sac"`
	Unit           string  `json:"unit"`
	ListPrice      float64 `json:"list_price"`
	DiscountAmount float64 `json:"discount_amount"`
}

func (r CreateInvoiceRequest) ToDomain() *Invoice {
	inv := &Invoice{
		CustomerName:   r.CustomerName,
		CustomerEmail:  r.CustomerEmail,
		CustomerPhone:  r.CustomerPhone,
		BillingAddress: r.BillingAddress,
		InvoiceDate:    r.InvoiceDate.ToTime(),
		DueDate:        r.DueDate.ToTime(),
		Currency:       r.Currency,
		Notes:          r.Notes,
		Status:         StatusDraft,
		TemplateType:   TemplateType(r.TemplateType),
		TripID:         r.TripID,
		VehicleNo:      r.VehicleNo,
		DriverName:     r.DriverName,
		DriverMobile:   r.DriverMobile,
		PickupLocation: r.PickupLocation,
		DropLocation:   r.DropLocation,
		BilledDistance: r.BilledDistance,
		PlaceOfSupply:  r.PlaceOfSupply,
		ReverseCharge:  r.ReverseCharge,
		EWayBillNo:     r.EWayBillNo,
		IRN:            r.IRN,
		BankName:       r.BankName,
		AccountNo:      r.AccountNo,
		IFSC:           r.IFSC,
		Branch:         r.Branch,
		Beneficiary:    r.Beneficiary,
	}

	if inv.TemplateType == "" {
		inv.TemplateType = TemplateTaxInvoice
	}

	for _, item := range r.Items {
		inv.Items = append(inv.Items, InvoiceItem{
			Description:    item.Description,
			Quantity:       item.Quantity,
			UnitPrice:      item.UnitPrice,
			TaxPercent:     item.TaxPercent,
			Amount:         item.Quantity * item.UnitPrice,
			HSNSAC:         item.HSNSAC,
			Unit:           item.Unit,
			ListPrice:      item.ListPrice,
			DiscountAmount: item.DiscountAmount,
		})
	}

	return inv
}
