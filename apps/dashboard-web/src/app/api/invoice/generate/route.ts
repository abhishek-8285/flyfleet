import { NextRequest, NextResponse } from 'next/server';

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
  status: string;
  currency: string;
  items: InvoiceItem[];
  notes?: string;
  metadata?: Record<string, string>;
}

export async function POST(request: NextRequest) {
  try {
    const invoiceData: InvoiceData = await request.json();

    // Call the Go backend to generate PDF
    const goBackendUrl = process.env.FLEET_CORE_URL || 'http://localhost:8080';
    
    const response = await fetch(`${goBackendUrl}/api/v1/invoice/generate-pdf`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(invoiceData),
    });

    if (!response.ok) {
      throw new Error(`Backend error: ${response.status}`);
    }

    const pdfBuffer = await response.arrayBuffer();
    
    return new NextResponse(pdfBuffer, {
      headers: {
        'Content-Type': 'application/pdf',
        'Content-Disposition': `attachment; filename="${invoiceData.invoiceNumber || 'invoice'}.pdf"`,
      },
    });
  } catch (error) {
    console.error('PDF generation error:', error);
    
    // Fallback: Generate simple HTML and convert client-side
    return NextResponse.json(
      { error: 'PDF generation temporarily unavailable' },
      { status: 500 }
    );
  }
}
