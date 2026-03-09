import axios from "axios";
import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";

const InvoicesList: React.FC = () => {
  const [invoices, setInvoices] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("portal_invoices");

  useEffect(() => {
    const fetchInvoices = async () => {
      try {
        const token = localStorage.getItem("customer_token");
        if (!token) return;

        const config = { headers: { Authorization: `Bearer ${token}` } };
        const res = await axios.get("/api/v1/portal/invoices", config);
        setInvoices(res.data || []);
      } catch (err) {
        console.error("Failed to fetch invoices", err);
      } finally {
        setLoading(false);
      }
    };

    fetchInvoices();
  }, []);

  const handleDownload = (invoice: any) => {
    // If we have a direct PDF URL in the model (from billing_v2 integration)
    if (invoice.pdf_url) {
      window.open(invoice.pdf_url, "_blank");
      return;
    }
    // Fallback: This would link to the billing_v2 export endpoint
    window.open(`/api/v1/billing/invoices/${invoice.id}/pdf`, "_blank");
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-white">
        {uiStrings?.title || "Billing & Invoices"}
      </h1>

      <div className="bg-surface-dark border border-border-color rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-text-muted">
            <thead className="bg-[#111a22] text-xs uppercase font-semibold text-text-muted/70">
              <tr>
                <th className="px-5 py-3">Invoice #</th>
                <th className="px-5 py-3">Amount</th>
                <th className="px-5 py-3">Status</th>
                <th className="px-5 py-3">Issue Date</th>
                <th className="px-5 py-3">Due Date</th>
                <th className="px-5 py-3 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
              {loading ? (
                <tr>
                  <td colSpan={6} className="p-10 text-center">
                    {uiStrings?.loading || "Loading invoices..."}
                  </td>
                </tr>
              ) : invoices.length === 0 ? (
                <tr>
                  <td colSpan={6} className="p-10 text-center">
                    {uiStrings?.empty || "No invoices found."}
                  </td>
                </tr>
              ) : (
                invoices.map((inv) => (
                  <tr
                    key={inv.id}
                    className="hover:bg-surface-light/5 transition-colors"
                  >
                    <td className="px-5 py-4 font-bold text-white uppercase">
                      {inv.invoice_number}
                    </td>
                    <td className="px-5 py-4 text-white font-medium">
                      {inv.currency} {inv.amount.toLocaleString()}
                    </td>
                    <td className="px-5 py-4">
                      <span
                        className={`px-2 py-1 rounded text-[10px] font-bold uppercase ${
                          inv.status === "PAID"
                            ? "bg-green-500/10 text-green-500"
                            : inv.status === "OVERDUE"
                              ? "bg-red-500/10 text-red-500"
                              : "bg-yellow-500/10 text-yellow-500"
                        }`}
                      >
                        {inv.status}
                      </span>
                    </td>
                    <td className="px-5 py-4">
                      {new Date(inv.issue_date).toLocaleDateString()}
                    </td>
                    <td className="px-5 py-4">
                      {new Date(inv.due_date).toLocaleDateString()}
                    </td>
                    <td className="px-5 py-4 text-right">
                      <button
                        onClick={() => handleDownload(inv)}
                        className="p-2 hover:bg-primary/20 text-primary rounded-lg transition-colors"
                        title="Download PDF"
                      >
                        <span className="material-symbols-outlined">
                          download
                        </span>
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default InvoicesList;
