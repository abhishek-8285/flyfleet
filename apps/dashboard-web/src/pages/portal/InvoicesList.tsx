import axios from "axios";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
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
    <div className="space-y-8 bg-surface-dim min-h-screen p-8 -mx-6 -mt-6">
      <div className="flex justify-between items-end bg-surface-container border border-outline rounded-3xl p-8 shadow-2xl">
        <div>
          <nav className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-on-surface-variant mb-4">
            <Link to="/portal" className="hover:text-primary transition-colors">Command Overview</Link>
            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
            <span className="text-white">Financial Ledger</span>
          </nav>
          <h1 className="text-4xl font-bold text-white tracking-tighter uppercase">
            {uiStrings?.title || "Accounts & Financial Ledger"}
          </h1>
        </div>
        <div className="flex gap-4">
          <button className="bg-surface-dim border border-outline text-white px-6 py-3.5 rounded-2xl text-[10px] font-black uppercase tracking-widest hover:bg-surface-container-highest transition-all flex items-center gap-3 active:scale-95 shadow-lg">
            <span className="material-symbols-outlined text-base">download_for_offline</span> Export Statement
          </button>
        </div>
      </div>

      <div className="bg-surface-container border border-outline rounded-[2.5rem] overflow-hidden shadow-2xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead className="bg-surface-dim border-b border-outline text-[9px] uppercase font-black text-on-surface-variant tracking-[0.2em]">
              <tr>
                <th className="px-8 py-5">Reference ID</th>
                <th className="px-8 py-5">Value (Base Currency)</th>
                <th className="px-8 py-5">Settlement Status</th>
                <th className="px-8 py-5">Issuance Date</th>
                <th className="px-8 py-5">Maturity Date</th>
                <th className="px-8 py-5 text-right">Operational Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-outline/30">
              {loading ? (
                <tr>
                  <td colSpan={6} className="p-20 text-center">
                    <div className="flex flex-col items-center gap-4">
                      <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary" />
                      <span className="text-[10px] font-black uppercase tracking-widest text-on-surface-variant">
                        {uiStrings?.loading || "Synchronising Financial Records..."}
                      </span>
                    </div>
                  </td>
                </tr>
              ) : invoices.length === 0 ? (
                <tr>
                  <td colSpan={6} className="p-20 text-center text-on-surface-variant">
                    <span className="material-symbols-outlined text-6xl opacity-20 mb-4 block">receipt_long</span>
                    <p className="text-[11px] font-black uppercase tracking-widest">
                      {uiStrings?.empty || "No financial records detected."}
                    </p>
                  </td>
                </tr>
              ) : (
                invoices.map((inv) => (
                  <tr
                    key={inv.id}
                    className="hover:bg-surface-light/5 transition-colors group"
                  >
                    <td className="px-8 py-5 font-bold text-white font-mono text-xs">
                      #{inv.invoice_number}
                    </td>
                    <td className="px-8 py-5 text-white font-bold tracking-tighter">
                      {inv.currency || "INR"} {inv.amount.toLocaleString(inv.currency === 'INR' ? 'en-IN' : 'en-US')}
                    </td>
                    <td className="px-8 py-5">
                      <span
                        className={`px-3 py-1 rounded-lg text-[9px] font-black uppercase tracking-[0.15em] 
                        ${
                          inv.status === "PAID"
                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                            : inv.status === "OVERDUE"
                              ? "bg-rose-500/10 text-rose-400 border border-rose-500/20"
                              : "bg-amber-500/10 text-amber-500 border border-amber-500/20"
                        }`}
                      >
                        {inv.status}
                      </span>
                    </td>
                    <td className="px-8 py-5 text-[11px] font-mono font-bold text-on-surface-variant">
                      {new Date(inv.issue_date).toLocaleDateString('en-GB')}
                    </td>
                    <td className="px-8 py-5 text-[11px] font-mono font-bold text-on-surface-variant">
                      {new Date(inv.due_date).toLocaleDateString('en-GB')}
                    </td>
                    <td className="px-8 py-5 text-right">
                      <button
                        onClick={() => handleDownload(inv)}
                        className="h-10 w-10 hover:bg-primary/10 text-on-surface-variant hover:text-primary border border-outline hover:border-primary/50 rounded-xl transition-all flex items-center justify-center active:scale-95"
                        title="Download PDF Manifest"
                      >
                        <span className="material-symbols-outlined text-lg">
                          file_download
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
