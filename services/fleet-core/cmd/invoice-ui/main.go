//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/pdf"
	"github.com/google/uuid"
)

type InvoiceItemRequest struct {
	ID           string                 `json:"id"`
	Description  string                 `json:"description"`
	Quantity     int                    `json:"quantity"`
	UnitPrice    float64                `json:"unitPrice"`
	TaxPercent   float64                `json:"taxPercent"`
	HSNCode      string                 `json:"hsnCode"`
	CustomFields map[string]interface{} `json:"customFields"`
}

type InvoiceRequest struct {
	InvoiceNumber  string                 `json:"invoiceNumber"`
	CustomerName   string                 `json:"customerName"`
	CustomerEmail  string                 `json:"customerEmail"`
	BillingAddress string                 `json:"billingAddress"`
	InvoiceDate    string                 `json:"invoiceDate"`
	DueDate        string                 `json:"dueDate"`
	Status         string                 `json:"status"`
	Currency       string                 `json:"currency"`
	Items          []InvoiceItemRequest   `json:"items"`
	Notes          string                 `json:"notes"`
	Comments       string                 `json:"comments"`
	Metadata       map[string]string      `json:"metadata"`
}

func main() {
	os.Setenv("CHROME_PATH", "/data/data/com.termux/files/usr/lib/chromium/chrome")
	generator := pdf.NewPDFGenerator(120 * time.Second)
	invoiceService := pdf.NewInvoiceService(generator)

	outputDir := filepath.Join(os.Getenv("HOME"), "flyfleet", "test-results")
	os.MkdirAll(outputDir, 0755)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, getIndexHTML())
	})

	http.HandleFunc("/generate-pdf", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req InvoiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		invoice := &models.Invoice{
			ID:             uuid.New(),
			OrgID:          uuid.New(),
			InvoiceNumber:  req.InvoiceNumber,
			CustomerName:   req.CustomerName,
			CustomerEmail:  req.CustomerEmail,
			BillingAddress: req.BillingAddress,
			Currency:       req.Currency,
			Status:         models.InvoiceStatus(req.Status),
			Notes:          req.Notes + "\n\nComments:\n" + req.Comments,
			Metadata:       make(map[string]interface{}),
		}

		if req.InvoiceDate != "" {
			invoice.InvoiceDate, _ = time.Parse(time.RFC3339, req.InvoiceDate)
		}
		if req.DueDate != "" {
			invoice.DueDate, _ = time.Parse(time.RFC3339, req.DueDate)
		}
		for k, v := range req.Metadata {
			invoice.Metadata[k] = v
		}

		items := make([]models.InvoiceItem, 0, len(req.Items))
		var subtotal, totalTax float64
		for _, itemReq := range req.Items {
			item := models.InvoiceItem{
				ID:           uuid.New(),
				Description:  itemReq.Description,
				Quantity:     itemReq.Quantity,
				UnitPrice:    itemReq.UnitPrice,
				TaxPercent:   itemReq.TaxPercent,
				HSNCode:      itemReq.HSNCode,
				CustomFields: itemReq.CustomFields,
			}
			itemSubtotal := float64(item.Quantity) * item.UnitPrice
			itemTax := itemSubtotal * (item.TaxPercent / 100)
			item.Total = itemSubtotal + itemTax
			subtotal += itemSubtotal
			totalTax += itemTax
			items = append(items, item)
		}

		invoice.Subtotal = subtotal
		invoice.TotalAmount = subtotal + totalTax
		invoice.BalanceDue = subtotal + totalTax

		ctx := context.Background()
		pdfData, err := invoiceService.GeneratePDF(ctx, invoice, items)
		if err != nil {
			http.Error(w, "PDF generation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("%s.pdf", invoice.InvoiceNumber))
		os.WriteFile(filename, pdfData, 0644)

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.pdf\"", invoice.InvoiceNumber))
		w.Write(pdfData)
	})

	port := ":8888"
	fmt.Printf("Invoice Generator UI running at http://localhost%s\n", port)
	fmt.Printf("Output directory: %s\n", outputDir)
	http.ListenAndServe(port, nil)
}

func getIndexHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invoice PDF Generator</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="min-h-screen bg-gray-50 py-8">
    <div class="max-w-4xl mx-auto px-4">
        <h1 class="text-3xl font-bold text-gray-900 mb-2">Invoice PDF Generator</h1>
        <p class="text-gray-600 mb-8">Upload invoice JSON, add comments, and generate PDF</p>

        <div id="status" class="hidden p-4 rounded-lg mb-6"></div>

        <div class="bg-white rounded-xl shadow-lg p-6 mb-6">
            <h2 class="text-xl font-semibold mb-4">Upload Invoice JSON</h2>
            
            <div class="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-blue-500 transition-colors">
                <input type="file" accept=".json" class="hidden" id="file-upload" />
                <label for="file-upload" class="cursor-pointer">
                    <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                    <p class="mt-2 text-sm text-gray-600" id="file-name">Drop JSON file here or click to browse</p>
                </label>
            </div>

            <div class="mt-4 flex gap-3">
                <button onclick="loadSample()" class="px-4 py-2 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors">
                    Load Sample (25 Items)
                </button>
                <button onclick="generatePDF()" id="generate-btn" class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50" disabled>
                    Generate PDF
                </button>
            </div>
        </div>

        <div class="bg-white rounded-xl shadow-lg p-6 mb-6">
            <h2 class="text-xl font-semibold mb-4">Add Comments / Notes</h2>
            <textarea id="comments" rows="4" class="w-full border border-gray-300 rounded-lg p-3 focus:ring-2 focus:ring-blue-500 focus:border-transparent" placeholder="Add any comments, notes, or special instructions for this invoice..."></textarea>
            <p class="text-sm text-gray-500 mt-2">These comments will be included in the generated PDF</p>
        </div>

        <div id="preview" class="bg-white rounded-xl shadow-lg p-6 hidden">
            <h2 class="text-xl font-semibold mb-4">Invoice Preview</h2>
            <div class="grid grid-cols-2 gap-4 text-sm" id="preview-data"></div>
            <div class="mt-6">
                <h3 class="font-medium mb-2">Line Items (<span id="item-count">0</span>)</h3>
                <div class="max-h-64 overflow-y-auto border rounded-lg">
                    <table class="w-full text-sm">
                        <thead class="bg-gray-50 sticky top-0">
                            <tr>
                                <th class="px-3 py-2 text-left">#</th>
                                <th class="px-3 py-2 text-left">Description</th>
                                <th class="px-3 py-2 text-right">Qty</th>
                                <th class="px-3 py-2 text-right">Price</th>
                                <th class="px-3 py-2 text-right">Tax %</th>
                                <th class="px-3 py-2 text-right">Total</th>
                            </tr>
                        </thead>
                        <tbody id="items-body"></tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>

    <script>
        let invoiceData = null;

        document.getElementById('file-upload').addEventListener('change', async (e) => {
            const file = e.target.files[0];
            if (!file) return;
            
            document.getElementById('file-name').textContent = file.name;
            
            const text = await file.text();
            try {
                invoiceData = JSON.parse(text);
                showPreview(invoiceData);
                showStatus('Loaded invoice with ' + (invoiceData.items?.length || 0) + ' items', 'success');
                document.getElementById('generate-btn').disabled = false;
            } catch (err) {
                showStatus('Invalid JSON file', 'error');
            }
        });

        function loadSample() {
            const items = [];
            for (let i = 0; i < 25; i++) {
                items.push({
                    id: 'item-' + (i + 1),
                    description: 'Industrial Product ' + (i + 1) + ' - Premium Quality Item with Extended Description',
                    quantity: Math.floor(Math.random() * 200) + 10,
                    unitPrice: Math.floor(Math.random() * 5000) + 500,
                    taxPercent: i % 2 === 0 ? 18 : 28,
                    hsnCode: '8482',
                    customFields: { category: 'Industrial', remarks: 'Batch ' + (i + 1) }
                });
            }

            invoiceData = {
                invoiceNumber: 'INV-2026-' + Date.now(),
                customerName: 'Acme Industries Pvt Ltd',
                customerEmail: 'billing@acme.com',
                billingAddress: '123 Industrial Area\\nMumbai, Maharashtra 400001\\nIndia',
                invoiceDate: new Date().toISOString(),
                dueDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
                status: 'PENDING',
                currency: 'INR',
                notes: 'Thank you for your business!',
                metadata: {
                    vehicle_no: 'MH02AB1234',
                    lr_no: 'LR-2026-001',
                    e_way_bill: 'EWB-2026-001',
                    place_of_supply: 'Maharashtra (27)'
                },
                items: items
            };

            showPreview(invoiceData);
            showStatus('Sample 25-item invoice loaded', 'info');
            document.getElementById('generate-btn').disabled = false;
        }

        function showPreview(data) {
            document.getElementById('preview').classList.remove('hidden');
            document.getElementById('preview-data').innerHTML = 
                '<div><span class="text-gray-500">Invoice Number:</span><p class="font-medium">' + data.invoiceNumber + '</p></div>' +
                '<div><span class="text-gray-500">Customer:</span><p class="font-medium">' + data.customerName + '</p></div>' +
                '<div><span class="text-gray-500">Items:</span><p class="font-medium">' + (data.items?.length || 0) + '</p></div>' +
                '<div><span class="text-gray-500">Currency:</span><p class="font-medium">' + data.currency + '</p></div>' +
                '<div><span class="text-gray-500">Status:</span><p class="font-medium">' + data.status + '</p></div>' +
                '<div><span class="text-gray-500">Due Date:</span><p class="font-medium">' + new Date(data.dueDate).toLocaleDateString() + '</p></div>';
            
            document.getElementById('item-count').textContent = data.items?.length || 0;
            
            const tbody = document.getElementById('items-body');
            let rows = '';
            data.items.forEach((item, idx) => {
                const total = item.quantity * item.unitPrice * (1 + item.taxPercent / 100);
                rows += '<tr class="border-t">' +
                    '<td class="px-3 py-2">' + (idx + 1) + '</td>' +
                    '<td class="px-3 py-2 truncate max-w-xs">' + item.description + '</td>' +
                    '<td class="px-3 py-2 text-right">' + item.quantity + '</td>' +
                    '<td class="px-3 py-2 text-right">₹' + item.unitPrice.toFixed(2) + '</td>' +
                    '<td class="px-3 py-2 text-right">' + item.taxPercent + '%</td>' +
                    '<td class="px-3 py-2 text-right font-medium">₹' + total.toFixed(2) + '</td>' +
                    '</tr>';
            });
            tbody.innerHTML = rows;
        }

        async function generatePDF() {
            if (!invoiceData) return;
            
            const btn = document.getElementById('generate-btn');
            btn.disabled = true;
            btn.textContent = 'Generating...';
            showStatus('Generating PDF...', 'info');

            const comments = document.getElementById('comments').value;
            const dataToSend = { ...invoiceData, comments: comments };

            try {
                const response = await fetch('/generate-pdf', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(dataToSend)
                });

                if (!response.ok) throw new Error('Generation failed');

                const blob = await response.blob();
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = (invoiceData.invoiceNumber || 'invoice') + '.pdf';
                document.body.appendChild(a);
                a.click();
                window.URL.revokeObjectURL(url);
                document.body.removeChild(a);

                showStatus('PDF generated successfully!', 'success');
            } catch (err) {
                showStatus('Failed to generate PDF: ' + err.message, 'error');
            } finally {
                btn.disabled = false;
                btn.textContent = 'Generate PDF';
            }
        }

        function showStatus(message, type) {
            const el = document.getElementById('status');
            el.classList.remove('hidden');
            const typeClasses = {
                success: 'bg-green-100 text-green-800',
                error: 'bg-red-100 text-red-800',
                info: 'bg-blue-100 text-blue-800'
            };
            el.className = 'p-4 rounded-lg mb-6 ' + typeClasses[type];
            el.textContent = message;
        }
    </script>
</body>
</html>`
}
