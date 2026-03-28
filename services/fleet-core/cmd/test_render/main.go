package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable"
	ctx := context.Background()
	pool, _ := pgxpool.New(ctx, dbURL)
	defer pool.Close()

	// THIS IS THE POWER: We provide EVERY possible data point to the template
	masterData := map[string]interface{}{
		"Invoice": map[string]interface{}{
			"Number": "29",
			"Date":   time.Now().Format("02-Jan-2026"),
			"Amount": 84525.00,
		},
		"Logistics": map[string]interface{}{
			"VehicleNo":   "DL01LAH8783",
			"EWayBill":    "751590248946",
			"LRNo":        "17261920",
			"Transporter": "Khushi Textiles Logistics",
		},
		"GST": map[string]interface{}{
			"HSN":     "5208",
			"Taxable": 80499.51,
			"IGST":    4024.98,
			"Total":   84525.00,
			"InWords": "Rupees Eighty Four Thousand Five Hundred Twenty Five Only",
		},
		"Custom": map[string]interface{}{
			"AnyNewField": "This is accessible to the template!",
		},
	}

	// The User has 100% Freedom here
	dynamicTemplate := `
	<!DOCTYPE html>
	<html>
	<body style="font-family: monospace; padding: 50px;">
		<h1>INDUSTRIAL DOCUMENT GENERATOR</h1>
		<div style="border: 2px solid #000; padding: 20px;">
			<p>VEHICLE: {{.Logistics.VehicleNo}}</p>
			<p>LR NO: {{.Logistics.LRNo}}</p>
			<p>E-WAY BILL: {{.Logistics.EWayBill}}</p>
			<hr>
			<h3>GST BREAKUP (HSN {{.GST.HSN}})</h3>
			<p>TAXABLE: ₹{{.GST.Taxable}}</p>
			<p>IGST: ₹{{.GST.IGST}}</p>
			<hr>
			<h2>GRAND TOTAL: ₹{{.Invoice.Amount}}</h2>
			<p><i>{{.GST.InWords}}</i></p>
		</div>
	</body>
	</html>`

	tmpl, _ := template.New("master").Parse(dynamicTemplate)
	var buf bytes.Buffer
	tmpl.Execute(&buf, masterData)

	os.WriteFile("master_access_invoice.html", buf.Bytes(), 0644)
	fmt.Println("✅ SUCCESS: Generated master_access_invoice.html with Full Data Access!")
}
