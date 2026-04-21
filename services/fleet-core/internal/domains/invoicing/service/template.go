package service

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/invoicing"
	"github.com/skip2/go-qrcode"
)

// GenerateInvoiceHTML generates a high-density, professional HTML for the invoice
func GenerateInvoiceHTML(inv *invoicing.Invoice) string {
	if inv.TemplateType == invoicing.TemplateTripNote {
		return renderTripNote(inv)
	}
	return renderTaxInvoice(inv)
}

func renderTaxInvoice(inv *invoicing.Invoice) string {
	qrCode := getQRCodeBase64(fmt.Sprintf("Invoice:%s|Amount:%.2f", inv.InvoiceNumber, inv.TotalAmount))
	
	var itemsRows strings.Builder
	for i, item := range inv.Items {
		itemsRows.WriteString(fmt.Sprintf(`
			<tr class="item-row">
				<td class="text-center">%d</td>
				<td>
					<div class="item-desc">%s</div>
				</td>
				<td class="text-center">%s</td>
				<td class="text-center">%.2f %s</td>
				<td class="text-right">₹%.2f</td>
				<td class="text-right">%.1f%%</td>
				<td class="text-right">₹%.2f</td>
			</tr>`,
			i+1, item.Description, item.HSNSAC, item.Quantity, item.Unit, item.UnitPrice, item.TaxPercent, item.Amount,
		))
	}

	amountInWords := AmountToWords(inv.TotalAmount)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&family=Roboto+Mono:wght@400;500&display=swap');
        
        * { box-sizing: border-box; }
        body {
            font-family: 'Inter', system-ui, sans-serif;
            background: #ffffff;
            color: #1e293b;
            margin: 0;
            padding: 40px;
            font-size: 11px;
            line-height: 1.4;
        }
        
        .invoice-box {
            max-width: 800px;
            margin: auto;
            border: 1px solid #e2e8f0;
            padding: 30px;
        }
        
        .header-table { width: 100%%; border-collapse: collapse; margin-bottom: 20px; }
        .header-table td { vertical-align: top; }
        
        .brand { font-size: 24px; font-weight: 800; color: #0f172a; margin-bottom: 5px; }
        .invoice-title { font-size: 18px; font-weight: 700; text-transform: uppercase; color: #64748b; text-align: right; }
        
        .meta-section { margin-bottom: 30px; display: grid; grid-template-columns: 1fr 1fr; gap: 40px; }
        .section-label { font-weight: 800; text-transform: uppercase; font-size: 9px; color: #94a3b8; margin-bottom: 5px; letter-spacing: 0.05em; }
        .section-val { font-size: 12px; font-weight: 600; color: #334155; }
        .section-sub { color: #64748b; margin-top: 2px; }

        .transporter-meta { background: #f8fafc; padding: 15px; border-radius: 8px; margin-bottom: 20px; display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px; }
        
        .items-table { width: 100%%; border-collapse: collapse; margin-bottom: 30px; }
        .items-table th { background: #0f172a; color: white; padding: 10px; text-transform: uppercase; font-size: 9px; letter-spacing: 0.05em; text-align: left; }
        .items-table td { padding: 12px 10px; border-bottom: 1px solid #f1f5f9; }
        .item-row:last-child td { border-bottom: 2px solid #0f172a; }
        .text-right { text-align: right; }
        .text-center { text-align: center; }
        
        .summary-section { display: grid; grid-template-columns: 1.5fr 1fr; gap: 40px; }
        .bank-details { padding: 15px; border: 1px solid #e2e8f0; border-radius: 8px; }
        
        .total-row { display: flex; justify-content: space-between; padding: 5px 0; font-size: 12px; }
        .total-grand { font-size: 18px; font-weight: 900; color: #0f172a; border-top: 2px solid #0f172a; padding-top: 10px; margin-top: 10px; }
        
        .amount-words { font-style: italic; font-weight: 600; color: #475569; margin-top: 10px; padding: 10px; background: #f1f5f9; border-radius: 4px; }
        
        .footer-declaration { margin-top: 40px; font-size: 9px; color: #64748b; line-height: 1.6; }
        .qr-code { float: right; width: 80px; height: 80px; }
    </style>
</head>
<body>
    <div class="invoice-box">
        <table class="header-table">
            <tr>
                <td>
                    <div class="brand">FLYFLEET</div>
                    <div style="font-weight: 600;">%s</div>
                    <div style="color: #64748b;">%s</div>
                    <div style="color: #64748b;">GSTIN: %s | PAN: %s</div>
                </td>
                <td>
                    <div class="invoice-title">Tax Invoice</div>
                    <div style="text-align: right; margin-top: 10px;">
                        <div class="section-label">Invoice Number</div>
                        <div class="section-val">#%s</div>
                        <div class="section-label" style="margin-top: 10px;">Date</div>
                        <div class="section-val">%s</div>
                    </div>
                </td>
            </tr>
        </table>

        <div class="transporter-meta">
            <div>
                <div class="section-label">Vehicle No</div>
                <div class="section-val">%s</div>
            </div>
            <div>
                <div class="section-label">Driver</div>
                <div class="section-val">%s</div>
            </div>
            <div>
                <div class="section-label">Place of Supply</div>
                <div class="section-val">%s</div>
            </div>
        </div>

        <div class="meta-section">
            <div>
                <div class="section-label">Billed To</div>
                <div class="section-val">%s</div>
                <div class="section-sub">%s</div>
                <div class="section-sub">%s</div>
            </div>
            <div>
                <div class="section-label">Transport Documentation</div>
                <div class="section-val"><span style="color: #94a3b8;">E-Way Bill:</span> %s</div>
                <div class="section-val"><span style="color: #94a3b8;">DISTANCE:</span> %.1f KM</div>
            </div>
        </div>

        <table class="items-table">
            <thead>
                <tr>
                    <th>Sr.</th>
                    <th>Description</th>
                    <th class="text-center">HSN/SAC</th>
                    <th class="text-center">Qty/Unit</th>
                    <th class="text-right">Price</th>
                    <th class="text-right">GST</th>
                    <th class="text-right">Total</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>

        <div class="summary-section">
            <div>
                <div class="bank-details">
                    <div class="section-label">Bank Payment Instructions</div>
                    <div style="font-weight: 600; color: #334155;">%s</div>
                    <div style="font-family: 'Roboto Mono', monospace;">A/C: %s</div>
                    <div style="font-family: 'Roboto Mono', monospace;">IFSC: %s</div>
                    <div style="font-size: 10px; color: #64748b; margin-top: 4px;">%s</div>
                </div>
                <div class="amount-words">
                    <div class="section-label">Amount in Words</div>
                    %s
                </div>
            </div>
            <div>
                <div class="total-row">
                    <span>Subtotal</span>
                    <span class="section-val">₹%.2f</span>
                </div>
                <div class="total-row">
                    <span>GST Amount</span>
                    <span class="section-val">₹%.2f</span>
                </div>
                <div class="total-row total-grand">
                    <span>Total Amount</span>
                    <span>₹%.2f</span>
                </div>
                <div style="text-align: center; margin-top: 20px;">
                    <img src="data:image/png;base64,%s" class="qr-code">
                </div>
            </div>
        </div>

        <div class="footer-declaration">
            <strong>DECLARATION:</strong><br>
            I/We hereby declare that though our aggregate turnover in any preceding financial year from 2017-18 onwards is more than the aggregate turnover notified under sub-rule (4) of rule 48, we are not required to prepare an invoice in terms of the provisions of the said sub-rule. This is a computer-generated invoice.
        </div>
    </div>
</body>
</html>`,
		inv.GstinSeller, inv.BillingAddress, inv.GstinSeller, inv.PanSeller,
		inv.InvoiceNumber, inv.InvoiceDate.Format("02 Jan 2006"),
		inv.VehicleNo, inv.DriverName, inv.PlaceOfSupply,
		inv.CustomerName, inv.BillingAddress, formatPhoneNumber(inv.CustomerPhone),
		inv.EWayBillNo, inv.BilledDistance,
		itemsRows.String(),
		inv.BankName, inv.AccountNo, inv.IFSC, inv.Branch,
		amountInWords,
		inv.Subtotal, inv.TaxAmount, inv.TotalAmount,
		qrCode,
	)
}

func renderTripNote(inv *invoicing.Invoice) string {
	// Porter style: Visual summary with map and timeline
	qrCode := getQRCodeBase64(fmt.Sprintf("Trip:%s|Invoice:%s", inv.TripID, inv.InvoiceNumber))
	
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&display=swap');
        body { font-family: 'Inter', sans-serif; background: #f8fafc; margin: 0; padding: 40px; }
        .card { max-width: 600px; margin: auto; background: white; border-radius: 32px; overflow: hidden; box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1); }
        .header { padding: 40px; text-align: center; background: #eff6ff; }
        .trip-visual { padding: 30px; border-bottom: 1px solid #f1f5f9; position: relative; }
        .map-placeholder { width: 100%%; height: 180px; background: #e2e8f0; border-radius: 20px; display: flex; align-items: center; justify-content: center; overflow: hidden; }
        .timeline { padding: 30px; }
        .point { display: flex; gap: 20px; position: relative; padding-bottom: 30px; }
        .point:last-child { padding-bottom: 0; }
        .point::after { content: ''; position: absolute; left: 6px; top: 20px; bottom: 0; width: 2px; background: #e2e8f0; }
        .point:last-child::after { display: none; }
        .dot { width: 14px; height: 14px; border-radius: 50%%; background: #3b82f6; z-index: 1; margin-top: 4px; }
        .dot.drop { background: #ef4444; }
        .total-badge { position: absolute; top: -30px; right: 30px; background: #0f172a; color: white; padding: 20px 30px; border-radius: 24px; text-align: center; box-shadow: 0 10px 15px -3px rgba(0,0,0,0.2); }
        .label { font-size: 10px; font-weight: 800; text-transform: uppercase; color: #94a3b8; letter-spacing: 1px; margin-bottom: 4px; }
        .val { font-weight: 700; color: #1e293b; }
    </style>
</head>
<body>
    <div class="card">
        <div class="header">
            <div style="font-size: 28px; font-weight: 900; color: #2563eb;">FLYFLEET</div>
            <div style="margin-top: 10px; font-weight: 600; color: #64748b;">Consignment Note / Trip Summary</div>
        </div>
        
        <div class="trip-visual">
            <div class="total-badge">
                <div class="label" style="color: #64748b;">Total Fare</div>
                <div style="font-size: 24px; font-weight: 900;">₹%.0f</div>
            </div>
            <div class="map-placeholder">
                <img src="%s" style="width: 100%%; height: 100%%; object-fit: cover;">
            </div>
            <div style="margin-top: 20px; display: flex; justify-content: space-between; align-items: center;">
                <div>
                   <div class="label">Vehicle</div>
                   <div class="val">%s</div>
                </div>
                <div style="text-align: right;">
                   <div class="label">Driver</div>
                   <div class="val">%s</div>
                </div>
            </div>
        </div>

        <div class="timeline">
            <div class="point">
                <div class="dot"></div>
                <div>
                    <div class="label">Pickup Location</div>
                    <div class="val" style="font-size: 13px;">%s</div>
                    <div style="font-size: 11px; color: #94a3b8; margin-top: 4px;">%s</div>
                </div>
            </div>
            <div class="point">
                <div class="dot drop"></div>
                <div>
                    <div class="label">Drop Location</div>
                    <div class="val" style="font-size: 13px;">%s</div>
                    <div style="font-size: 11px; color: #94a3b8; margin-top: 4px;">%s</div>
                </div>
            </div>
        </div>

        <div style="padding: 30px; background: #f8fafc; display: flex; justify-content: space-between; align-items: center;">
            <div style="font-size: 11px; color: #64748b;">
                Invoice #%s<br>
                Trip ID: %s
            </div>
            <img src="data:image/png;base64,%s" style="width: 60px; height: 60px;">
        </div>
    </div>
</body>
</html>`,
		inv.TotalAmount,
		inv.MapUrl,
		inv.VehicleNo,
		inv.DriverName,
		inv.PickupLocation, formatTime(inv.PickupTime),
		inv.DropLocation, formatTime(inv.DropTime),
		inv.InvoiceNumber,
		inv.TripID,
		qrCode,
	)
}

func formatTime(t *time.Time) string {
	if t == nil {
		return "N/A"
	}
	return t.Format("15:04, 02 Jan 2006")
}

func getQRCodeBase64(data string) string {
	png, err := qrcode.Encode(data, qrcode.Medium, 256)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(png)
}

func formatPhoneNumber(phone string) string {
	digits := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digits += string(char)
		}
	}
	if len(digits) == 10 {
		return fmt.Sprintf("+91 %s %s", digits[:5], digits[5:])
	}
	if len(digits) == 12 && strings.HasPrefix(digits, "91") {
		return fmt.Sprintf("+91 %s %s", digits[2:7], digits[7:])
	}
	return phone
}
