package models

type GSTBreakdown struct {
	CGSTRate float64 `json:"cgst_rate"`
	CGSTAmt  float64 `json:"cgst_amt"`
	SGSTRate float64 `json:"sgst_rate"`
	SGSTAmt  float64 `json:"sgst_amt"`
	IGSTRate float64 `json:"igst_rate"`
	IGSTAmt  float64 `json:"igst_amt"`
	TotalTax float64 `json:"total_tax"`
}

type GSTTripInvoice struct {
	TripID         string       `json:"trip_id"`
	SupplierGSTIN  string       `json:"supplier_gstin"`
	RecipientGSTIN string       `json:"recipient_gstin"`
	PlaceOfSupply  string       `json:"place_of_supply"`
	BaseAmount     float64      `json:"base_amount"`
	TaxBreakdown   GSTBreakdown `json:"tax_breakdown"`
	TotalAmount    float64      `json:"total_amount"`
}
