package pdf

import (
	"fmt"
	"strings"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
)

// GenerateMinimalistInvoiceHTML generates a sleek, dark-themed minimalist invoice
func GenerateMinimalistInvoiceHTML(invoice *models.Invoice, items []models.InvoiceItem) string {
	var itemsRows strings.Builder
	for _, item := range items {
		itemsRows.WriteString(fmt.Sprintf(`
			<div style="display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #334155;">
				<div style="flex: 2;">
					<div style="font-weight: 600; color: #f8fafc;">%s</div>
					<div style="font-size: 8pt; color: #94a3b8;">Qty: %d x ₹%.2f</div>
				</div>
				<div style="flex: 1; text-align: right; color: #f8fafc; font-weight: 600;">
					₹%.2f
				</div>
			</div>`,
			item.Description, item.Quantity, item.UnitPrice, item.Total,
		))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body {
			font-family: 'Inter', system-ui, sans-serif;
			background: #0f172a;
			color: #cbd5e1;
			margin: 0;
			padding: 40px;
		}
		.card {
			background: #1e293b;
			border-radius: 16px;
			padding: 40px;
			max-width: 800px;
			margin: auto;
			box-shadow: 0 25px 50px -12px rgba(0,0,0,0.5);
		}
		.header {
			display: flex;
			justify-content: space-between;
			margin-bottom: 40px;
		}
		.brand {
			font-size: 24pt;
			font-weight: 900;
			color: #38bdf8;
			letter-spacing: -1px;
		}
		.status {
			background: #064e3b;
			color: #34d399;
			padding: 4px 12px;
			border-radius: 99px;
			font-size: 9pt;
			text-transform: uppercase;
			font-weight: 700;
		}
		.total-box {
			margin-top: 40px;
			padding-top: 20px;
			border-top: 2px solid #38bdf8;
			display: flex;
			justify-content: space-between;
			align-items: center;
		}
		.amount {
			font-size: 28pt;
			font-weight: 800;
			color: #f8fafc;
		}
	</style>
</head>
<body>
	<div class="card">
		<div class="header">
			<div class="brand">FLYFLEET</div>
			<div><span class="status">%s</span></div>
		</div>
		
		<div style="margin-bottom: 40px;">
			<div style="font-size: 9pt; color: #94a3b8; text-transform: uppercase; letter-spacing: 1px;">Invoice To</div>
			<div style="font-size: 18pt; font-weight: 700; color: #f8fafc; margin-top: 5px;">%s</div>
			<div style="font-size: 10pt; margin-top: 5px;">Invoice #%s</div>
		</div>

		<div style="margin-top: 20px;">
			%s
		</div>

		<div class="total-box">
			<div>
				<div style="font-size: 9pt; color: #94a3b8; text-transform: uppercase;">Total Balance Due</div>
				<div style="font-size: 10pt; margin-top: 5px; color: #38bdf8;">Due by %s</div>
			</div>
			<div class="amount">₹%.2f</div>
		</div>

		<div style="margin-top: 40px; font-size: 8pt; color: #64748b; text-align: center;">
			Thank you for choosing FlyFleet Global Logistics.
		</div>
	</div>
</body>
</html>`,
		invoice.Status,
		invoice.CustomerName,
		invoice.InvoiceNumber,
		itemsRows.String(),
		invoice.DueDate.Format("Jan 02, 2006"),
		invoice.BalanceDue,
	)

	return html
}
