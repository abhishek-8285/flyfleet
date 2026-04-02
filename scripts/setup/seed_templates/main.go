package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable"
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// 1. Professional Template
	profHTML := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: sans-serif; font-size: 9pt; padding: 20px; }
		.header { display: flex; justify-content: space-between; background: #3b82f6; color: white; padding: 20px; }
		.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin: 20px 0; }
		table { width: 100%; border-collapse: collapse; }
		th { background: #f1f5f9; padding: 8px; text-align: left; }
		td { padding: 8px; border-bottom: 1px solid #e2e8f0; }
		.total-box { text-align: right; margin-top: 20px; font-weight: bold; font-size: 12pt; }
	</style>
</head>
<body>
	<div class="header">
		<div>
			<h1>{{.CompanyName}}</h1>
			<p>{{.CompanyAddress}}</p>
		</div>
		<div style="text-align: right">
			<h2>INVOICE</h2>
			<p>#{{.InvoiceNumber}}</p>
		</div>
	</div>
	<div class="grid">
		<div>
			<strong>BILL TO:</strong>
			<p>{{.CustomerName}}<br>{{.CustomerAddress}}</p>
		</div>
		<div style="text-align: right">
			<p>Date: {{.InvoiceDate}}</p>
			<p>Due: {{.DueDate}}</p>
		</div>
	</div>
	<table>
		<thead>
			<tr>
				<th>Description</th>
				<th>Qty</th>
				<th>Price</th>
				<th>Total</th>
			</tr>
		</thead>
		<tbody>
			{{range .Items}}
			<tr>
				<td>{{.Description}}</td>
				<td>{{.Quantity}}</td>
				<td>₹{{.UnitPrice}}</td>
				<td>₹{{.Total}}</td>
			</tr>
			{{end}}
		</tbody>
	</table>
	<div class="total-box">
		Total: ₹{{.TotalAmount}}
	</div>
</body>
</html>`

	// 2. Minimalist Dark Template
	darkHTML := `<!DOCTYPE html>
<html>
<head>
	<style>
		body { background: #0f172a; color: #f8fafc; font-family: sans-serif; padding: 50px; }
		.card { background: #1e293b; padding: 40px; border-radius: 12px; }
		.brand { color: #38bdf8; font-size: 24pt; font-weight: bold; }
		.item { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #334155; }
		.total { font-size: 30pt; color: #38bdf8; margin-top: 20px; }
	</style>
</head>
<body>
	<div class="card">
		<div class="brand">FLYFLEET</div>
		<p>Invoice for {{.CustomerName}}</p>
		<p>#{{.InvoiceNumber}}</p>
		<div style="margin: 30px 0;">
			{{range .Items}}
			<div class="item">
				<span>{{.Description}}</span>
				<span>₹{{.Total}}</span>
			</div>
			{{end}}
		</div>
		<div class="total">₹{{.TotalAmount}}</div>
	</div>
</body>
</html>`

	// Seed templates
	insertTemplate(ctx, pool, "professional", profHTML, true)
	insertTemplate(ctx, pool, "minimalist", darkHTML, false)

	fmt.Println("✅ Templates seeded successfully into DB.")
}

func insertTemplate(ctx context.Context, pool *pgxpool.Pool, name, html string, isDefault bool) {
	_, err := pool.Exec(ctx, `
		INSERT INTO billing_invoice_templates (name, version, html_content, is_default)
		VALUES ($1, 1, $2, $3)
		ON CONFLICT (name, version) DO UPDATE SET html_content = EXCLUDED.html_content
	`, name, html, isDefault)
	if err != nil {
		log.Fatalf("Failed to seed template %s: %v", name, err)
	}
}
