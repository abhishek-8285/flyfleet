'use client';

import { useState, useCallback } from 'react';

interface InvoiceItem {
  id: string;
  description: string;
  quantity: number;
  unitPrice: number;
  taxPercent: number;
  hsnCode?: string;
  customFields?: Record<string, unknown>;
}

interface InvoiceData {
  invoiceNumber: string;
  customerName: string;
  customerEmail: string;
  billingAddress: string;
  invoiceDate: string;
  dueDate: string;
  status: 'PAID' | 'PENDING' | 'OVERDUE' | 'PARTIAL';
  currency: string;
  items: InvoiceItem[];
  notes?: string;
  metadata?: Record<string, string>;
}

export default function InvoiceGeneratorPage() {
  const [jsonFile, setJsonFile] = useState<File | null>(null);
  const [invoiceData, setInvoiceData] = useState<InvoiceData | null>(null);
  const [generating, setGenerating] = useState(false);
  const [status, setStatus] = useState<{ type: 'success' | 'error' | 'info'; message: string } | null>(null);
  const [itemCount, setItemCount] = useState(0);

  const handleFileUpload = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    if (!file.name.endsWith('.json')) {
      setStatus({ type: 'error', message: 'Please upload a JSON file' });
      return;
    }

    setJsonFile(file);
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = JSON.parse(e.target?.result as string);
        setInvoiceData(data);
        setItemCount(data.items?.length || 0);
        setStatus({ type: 'success', message: `Loaded invoice with ${data.items?.length || 0} items` });
      } catch (err) {
        setStatus({ type: 'error', message: 'Invalid JSON file' });
      }
    };
    reader.readAsText(file);
  }, []);

  const handleGeneratePDF = useCallback(async () => {
    if (!invoiceData) return;
    
    setGenerating(true);
    setStatus({ type: 'info', message: 'Generating PDF...' });

    try {
      const response = await fetch('/api/invoice/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(invoiceData),
      });

      if (!response.ok) throw new Error('Generation failed');

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${invoiceData.invoiceNumber || 'invoice'}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      setStatus({ type: 'success', message: 'PDF generated successfully!' });
    } catch (err) {
      setStatus({ type: 'error', message: 'Failed to generate PDF' });
    } finally {
      setGenerating(false);
    }
  }, [invoiceData]);

  const handleLoadSample = useCallback(() => {
    const sample: InvoiceData = {
      invoiceNumber: 'INV-2026-001',
      customerName: 'Acme Industries Pvt Ltd',
      customerEmail: 'billing@acme.com',
      billingAddress: '123 Industrial Area\nMumbai, Maharashtra 400001\nIndia',
      invoiceDate: new Date().toISOString(),
      dueDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
      status: 'PENDING',
      currency: 'INR',
      notes: 'Thank you for your business!',
      metadata: {
        vehicle_no: 'MH02AB1234',
        lr_no: 'LR-2026-001',
        e_way_bill: 'EWB-2026-001',
        place_of_supply: 'Maharashtra (27)',
      },
      items: Array.from({ length: 25 }, (_, i) => ({
        id: `item-${i + 1}`,
        description: `Industrial Product ${i + 1} - Premium Quality Item with Extended Description`,
        quantity: Math.floor(Math.random() * 200) + 10,
        unitPrice: Math.floor(Math.random() * 5000) + 500,
        taxPercent: i % 2 === 0 ? 18 : 28,
        hsnCode: '8482',
        customFields: { category: 'Industrial', remarks: `Batch ${i + 1}` },
      })),
    };
    setInvoiceData(sample);
    setItemCount(25);
    setStatus({ type: 'info', message: 'Sample 25-item invoice loaded' });
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Invoice PDF Generator</h1>
        <p className="text-gray-600 mb-8">Upload invoice JSON or generate sample stress test data</p>

        {status && (
          <div className={`p-4 rounded-lg mb-6 ${
            status.type === 'success' ? 'bg-green-100 text-green-800' :
            status.type === 'error' ? 'bg-red-100 text-red-800' :
            'bg-blue-100 text-blue-800'
          }`}>
            {status.message}
          </div>
        )}

        <div className="bg-white rounded-xl shadow-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Upload Invoice JSON</h2>
          
          <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-blue-500 transition-colors">
            <input
              type="file"
              accept=".json"
              onChange={handleFileUpload}
              className="hidden"
              id="file-upload"
            />
            <label htmlFor="file-upload" className="cursor-pointer">
              <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              <p className="mt-2 text-sm text-gray-600">
                {jsonFile ? jsonFile.name : 'Drop JSON file here or click to browse'}
              </p>
            </label>
          </div>

          <div className="mt-4 flex gap-3">
            <button
              onClick={handleLoadSample}
              className="px-4 py-2 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors"
            >
              Load Sample (25 Items)
            </button>
            {invoiceData && (
              <button
                onClick={handleGeneratePDF}
                disabled={generating}
                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
              >
                {generating ? 'Generating...' : 'Generate PDF'}
              </button>
            )}
          </div>
        </div>

        {invoiceData && (
          <div className="bg-white rounded-xl shadow-lg p-6">
            <h2 className="text-xl font-semibold mb-4">Invoice Preview</h2>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-gray-500">Invoice Number:</span>
                <p className="font-medium">{invoiceData.invoiceNumber}</p>
              </div>
              <div>
                <span className="text-gray-500">Customer:</span>
                <p className="font-medium">{invoiceData.customerName}</p>
              </div>
              <div>
                <span className="text-gray-500">Items:</span>
                <p className="font-medium">{itemCount}</p>
              </div>
              <div>
                <span className="text-gray-500">Currency:</span>
                <p className="font-medium">{invoiceData.currency}</p>
              </div>
              <div>
                <span className="text-gray-500">Status:</span>
                <p className="font-medium">
                  <span className={`px-2 py-1 rounded text-xs ${
                    invoiceData.status === 'PAID' ? 'bg-green-100 text-green-800' :
                    invoiceData.status === 'PENDING' ? 'bg-yellow-100 text-yellow-800' :
                    invoiceData.status === 'OVERDUE' ? 'bg-red-100 text-red-800' :
                    'bg-blue-100 text-blue-800'
                  }`}>
                    {invoiceData.status}
                  </span>
                </p>
              </div>
              <div>
                <span className="text-gray-500">Due Date:</span>
                <p className="font-medium">{new Date(invoiceData.dueDate).toLocaleDateString()}</p>
              </div>
            </div>

            {invoiceData.items && invoiceData.items.length > 0 && (
              <div className="mt-6">
                <h3 className="font-medium mb-2">Line Items ({invoiceData.items.length})</h3>
                <div className="max-h-64 overflow-y-auto border rounded-lg">
                  <table className="w-full text-sm">
                    <thead className="bg-gray-50 sticky top-0">
                      <tr>
                        <th className="px-3 py-2 text-left">#</th>
                        <th className="px-3 py-2 text-left">Description</th>
                        <th className="px-3 py-2 text-right">Qty</th>
                        <th className="px-3 py-2 text-right">Price</th>
                        <th className="px-3 py-2 text-right">Tax %</th>
                        <th className="px-3 py-2 text-right">Total</th>
                      </tr>
                    </thead>
                    <tbody>
                      {invoiceData.items.map((item, idx) => (
                        <tr key={item.id} className="border-t page-break-inside-avoid">
                          <td className="px-3 py-2">{idx + 1}</td>
                          <td className="px-3 py-2 truncate max-w-xs">{item.description}</td>
                          <td className="px-3 py-2 text-right">{item.quantity}</td>
                          <td className="px-3 py-2 text-right">₹{item.unitPrice.toFixed(2)}</td>
                          <td className="px-3 py-2 text-right">{item.taxPercent}%</td>
                          <td className="px-3 py-2 text-right font-medium">
                            ₹{(item.quantity * item.unitPrice * (1 + item.taxPercent / 100)).toFixed(2)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
