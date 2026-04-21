package accounting

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
)

type QuickBooksExporter struct{}

func NewQuickBooksExporter() *QuickBooksExporter {
	return &QuickBooksExporter{}
}

func (e *QuickBooksExporter) GetName() models.ExternalSystem {
	return models.SystemQuickBooks
}

func (e *QuickBooksExporter) ExportLedger(tenantID uuid.UUID, entries []models.LedgerEntry) ([]byte, string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	_ = w.Write([]string{"TransactionID", "Date", "Type", "Account", "Amount", "Currency", "Description"})

	for _, entry := range entries {
		desc := ""
		if entry.Description != nil {
			desc = *entry.Description
		}
		_ = w.Write([]string{
			entry.ID.String(),
			entry.CreatedAt.Format("2006-01-02"),
			string(entry.EntryType),
			entry.LedgerAccount,
			fmt.Sprintf("%.2f", entry.Amount),
			entry.Currency,
			desc,
		})
	}

	w.Flush()
	return buf.Bytes(), "ledger_quickbooks.csv", nil
}

func (e *QuickBooksExporter) ExportInvoices(tenantID uuid.UUID, invoices []models.Invoice) ([]byte, string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	_ = w.Write([]string{"InvoiceNumber", "Customer", "Date", "DueDate", "Subtotal", "Tax", "Total", "Currency", "Status"})

	for _, inv := range invoices {
		taxAmount := inv.TotalAmount - inv.Subtotal
		_ = w.Write([]string{
			inv.InvoiceNumber,
			inv.TenantID.String(),
			inv.InvoiceDate.Format("2006-01-02"),
			inv.DueDate.Format("2006-01-02"),
			fmt.Sprintf("%.2f", inv.Subtotal),
			fmt.Sprintf("%.2f", taxAmount),
			fmt.Sprintf("%.2f", inv.TotalAmount),
			inv.Currency,
			string(inv.Status),
		})
	}

	w.Flush()
	return buf.Bytes(), "invoices_quickbooks.csv", nil
}

type XeroExporter struct{}

func NewXeroExporter() *XeroExporter {
	return &XeroExporter{}
}

func (e *XeroExporter) GetName() models.ExternalSystem {
	return models.SystemXero
}

func (e *XeroExporter) ExportLedger(tenantID uuid.UUID, entries []models.LedgerEntry) ([]byte, string, error) {
	// For now, Xero also gets CSV, but structured for Xero import
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"*ContactName", "*EmailAddress", "Reference", "*InvoiceNumber", "*InvoiceDate", "*DueDate", "*Quantity", "*UnitAmount", "*AccountCode", "*TaxType", "Description"})

	for _, entry := range entries {
		desc := ""
		if entry.Description != nil {
			desc = *entry.Description
		}
		_ = w.Write([]string{
			"FlyFleet Org",
			"",
			entry.ID.String(),
			"",
			entry.CreatedAt.Format("02/01/2006"),
			entry.CreatedAt.Format("02/01/2006"),
			"1",
			fmt.Sprintf("%.2f", entry.Amount),
			entry.LedgerAccount,
			"NONE",
			desc,
		})
	}

	w.Flush()
	return buf.Bytes(), "ledger_xero.csv", nil
}

func (e *XeroExporter) ExportInvoices(tenantID uuid.UUID, invoices []models.Invoice) ([]byte, string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"*ContactName", "*EmailAddress", "Reference", "*InvoiceNumber", "*InvoiceDate", "*DueDate", "*Quantity", "*UnitAmount", "*AccountCode", "*TaxType", "Description"})

	for _, inv := range invoices {
		_ = w.Write([]string{
			"Organization " + inv.TenantID.String(),
			"",
			inv.ID.String(),
			inv.InvoiceNumber,
			inv.InvoiceDate.Format("02/01/2006"),
			inv.DueDate.Format("02/01/2006"),
			"1",
			fmt.Sprintf("%.2f", inv.TotalAmount),
			"200", // Default Sales Account
			"OUTPUT",
			"Monthly Subscription - " + inv.InvoiceNumber,
		})
	}

	w.Flush()
	return buf.Bytes(), "invoices_xero.csv", nil
}
