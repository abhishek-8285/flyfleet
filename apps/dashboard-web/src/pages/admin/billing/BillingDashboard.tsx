import {
    ArrowDownRightIcon,
    ArrowUpRightIcon,
    BanknotesIcon,
    CalendarDaysIcon,
    CreditCardIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import billingService from "../../../services/billingService";
import { invoicingV2Service } from "../../../services/invoicingV2Service";
import { BillingMetrics, Invoice } from "../../../types/billing";
import UsageOverview from "./UsageOverview";
import SubscriptionPlans from "./SubscriptionPlans";
import InvoiceForm from "./InvoiceForm";

const BillingDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [history, setHistory] = useState<any[]>([]);
  const [metrics, setMetrics] = useState<BillingMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("admin_billing");
  const [billingData, setBillingData] = useState<any>(null); // Assuming billingData is needed for new stat cards
  const [activeTab, setActiveTab] = useState("Overview"); // For new tab navigation

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [invoicesData, v2InvoicesData, metricsData, settingsData] = await Promise.all([
        billingService.getInvoices(1, 10),
        invoicingV2Service.listInvoices(10, 0),
        billingService.getMetrics(),
        billingService.getSettings(),
      ]);
      
      // Merge and sort by date descending
      const allInvoices = [
        ...(invoicesData?.data || []),
        ...(v2InvoicesData?.invoices || [])
      ].sort((a, b) => new Date(b.invoice_date).getTime() - new Date(a.invoice_date).getTime());

      setInvoices(allInvoices.slice(0, 10)); // Top 10
      setMetrics(metricsData);
      
      setBillingData({
        balance: metricsData?.pending_payment || 0,
        next_billing_date: settingsData?.auto_invoicing_enabled ? "Auto-generated" : "Manual",
        tier: settingsData?.base_currency || "USD",
      });
    } catch (error) {
      console.error("Failed to fetch billing data", error);
    } finally {
      setLoading(false);
    }
  };

  const statusColors: Record<string, string> = {
    PAID: "bg-green-500/10 text-green-500 border-green-500/20",
    PENDING: "bg-amber-500/10 text-amber-500 border-amber-500/20",
    OVERDUE: "bg-red-500/10 text-red-500 border-red-500/20",
    PARTIAL: "bg-blue-500/10 text-blue-500 border-blue-500/20",
    DRAFT: "bg-gray-500/10 text-gray-500 border-gray-500/20",
    SENT: "bg-purple-500/10 text-purple-500 border-purple-500/20",
    VIEWED: "bg-cyan-500/10 text-cyan-500 border-cyan-500/20",
  };

  if (loading)
    return (
      <div className="p-8 text-center text-text-muted">
        {uiStrings?.loading || "Loading billing data..."}
      </div>
    );

  return (
    <div className="p-8">
      {/* Header */}
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-white text-3xl font-bold mb-1">
            {uiStrings?.title || "Billing & Subscriptions"}
          </h1>
          <p className="text-text-muted">
            {uiStrings?.subtitle ||
              "Manage payment methods, invoices, and service tiers"}
          </p>
        </div>
        <div className="flex gap-4">
          <button 
            onClick={() => navigate("/admin/billing/new")}
            className="bg-white/10 text-white px-6 py-2 rounded-xl font-bold hover:bg-white/20 transition-colors"
          >
            {uiStrings?.btn_create_invoice || "Create Invoice"}
          </button>
          <button className="bg-primary text-white px-6 py-2 rounded-xl font-bold hover:bg-primary-hover transition-colors">
            {uiStrings?.btn_manage || "Manage Subscription"}
          </button>
        </div>
      </div>

      {/* New Stat Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-surface-dark border border-border-color rounded-2xl p-6">
          <div className="flex items-center gap-3 mb-4">
            <BanknotesIcon className="h-6 w-6 text-primary" />
            <span className="text-text-muted font-medium">
              {uiStrings?.stat_balance || "Current Balance"}
            </span>
          </div>
          <div className="text-3xl font-bold text-white">
            ₹{billingData?.balance?.toLocaleString('en-IN', { minimumFractionDigits: 2 })}
          </div>
        </div>

        <div className="bg-surface-dark border border-border-color rounded-2xl p-6">
          <div className="flex items-center gap-3 mb-4">
            <CalendarDaysIcon className="h-6 w-6 text-primary" />
            <span className="text-text-muted font-medium">
              {uiStrings?.stat_next_bill || "Next Bill Date"}
            </span>
          </div>
          <div className="text-3xl font-bold text-white">
            {billingData?.next_billing_date}
          </div>
        </div>

        <div className="bg-surface-dark border border-border-color rounded-2xl p-6">
          <div className="flex items-center gap-3 mb-4">
            <CreditCardIcon className="h-6 w-6 text-primary" />
            <span className="text-text-muted font-medium">
              {uiStrings?.stat_tier || "Active Tier"}
            </span>
          </div>
          <div className="text-3xl font-bold text-white capitalize">
            {billingData?.tier}
          </div>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="flex gap-8 mb-8 border-b border-border-color overflow-x-auto">
        {["Overview", "Usage", "Subscription Plans", "Invoice History", "Miscellaneous", "Payment Methods"].map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`pb-4 px-2 font-medium transition-all whitespace-nowrap ${
              activeTab === tab
                ? "text-primary border-b-2 border-primary"
                : "text-text-muted hover:text-white"
            }`}
          >
            {tab === "Overview"
              ? uiStrings?.tab_overview || tab
              : tab === "Usage"
                ? tab
                : tab === "Subscription Plans"
                  ? tab
                  : tab === "Invoice History"
                    ? uiStrings?.tab_history || tab
                    : tab === "Miscellaneous"
                      ? "Miscellaneous"
                      : tab === "Payment Methods"
                        ? uiStrings?.tab_methods || tab
                        : tab}
          </button>
        ))}
      </div>

      {activeTab === "Usage" && <UsageOverview />}
      {activeTab === "Subscription Plans" && <SubscriptionPlans />}

      {activeTab === "Overview" && (
        <>
          {/* New Stat Cards */}

      {/* Recent Invoices */}
      <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden">
        <div className="p-6 border-b border-white/10 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-white">
            {uiStrings?.title_recent_invoices || "Recent Invoices"}
          </h2>
          <button
            onClick={() => navigate("/admin/billing/invoices")}
            className="text-primary hover:underline text-sm font-medium"
          >
            {uiStrings?.btn_view_all || "View All"}
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead className="bg-surface-darker border-b border-border-color">
              <tr>
                <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                  {uiStrings?.th_invoice_id || "Invoice ID"}
                </th>
                <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                  Category
                </th>
                <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                  {uiStrings?.th_date || "Date"}
                </th>
                <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                  {uiStrings?.th_amount || "Amount"}
                </th>
                <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                  {uiStrings?.th_status || "Status"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {invoices.map((invoice) => (
                <tr
                  key={invoice.id}
                  onClick={() => navigate(`/admin/billing/invoices/${invoice.id}`)}
                  className="hover:bg-white/5 transition-colors cursor-pointer"
                >
                  <td className="px-6 py-4">
                    <span className="text-white font-medium">
                      #{invoice.id.substring(0, 8).toUpperCase()}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`text-xs px-2 py-0.5 rounded border ${
                      invoice.category === "Miscellaneous" 
                        ? "bg-purple-500/10 text-purple-400 border-purple-500/20" 
                        : "bg-white/5 text-gray-400 border-white/10"
                    }`}>
                      {invoice.category || "Transportation"}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-gray-400">
                      {new Date(invoice.invoice_date).toLocaleDateString()}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-white font-semibold">
                      ₹{invoice.total_amount.toLocaleString('en-IN')}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-2.5 py-1 rounded-full text-xs font-medium border ${statusColors[invoice.status]}`}
                    >
                      {invoice.status}
                    </span>
                  </td>
                </tr>
              ))}
              {invoices.length === 0 && (
                <tr>
                  <td
                    colSpan={4}
                    className="px-6 py-12 text-center text-gray-500"
                  >
                    {uiStrings?.msg_no_invoices || "No recent invoices found."}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
            </div>
        </>
      )}

      {activeTab === "Invoice History" && (
        <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden min-h-[400px]">
           <div className="p-12 text-center text-text-muted">
              History content will be paginated here.
           </div>
        </div>
      )}

      {activeTab === "Miscellaneous" && (
        <div className="space-y-6">
          <div className="bg-purple-500/10 border border-purple-500/20 p-6 rounded-2xl">
            <h2 className="text-lg font-bold text-purple-400 mb-2">Miscellaneous Billing</h2>
            <p className="text-sm text-purple-300/70">
              Track ad-hoc charges, administrative fees, and non-transportation items here.
            </p>
          </div>
          
          <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden">
            <div className="p-6 border-b border-white/10">
              <h3 className="text-white font-semibold">Ad-hoc Invoice List</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-left">
                <thead className="bg-surface-darker border-b border-border-color">
                  <tr>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">ID</th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">Customer</th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">Date</th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">Amount</th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {invoices.filter(i => i.category === "Miscellaneous").map((invoice) => (
                    <tr
                      key={invoice.id}
                      onClick={() => navigate(`/admin/billing/invoices/${invoice.id}`)}
                      className="hover:bg-white/5 transition-colors cursor-pointer"
                    >
                      <td className="px-6 py-4 text-white font-medium">#{invoice.id.substring(0, 8).toUpperCase()}</td>
                      <td className="px-6 py-4 text-gray-300">{invoice.customer_name}</td>
                      <td className="px-6 py-4 text-gray-400">{new Date(invoice.invoice_date).toLocaleDateString()}</td>
                      <td className="px-6 py-4 text-white font-semibold">₹{invoice.total_amount.toLocaleString('en-IN')}</td>
                      <td className="px-6 py-4">
                        <span className={`px-2.5 py-1 rounded-full text-xs font-medium border ${statusColors[invoice.status]}`}>
                          {invoice.status}
                        </span>
                      </td>
                    </tr>
                  ))}
                  {invoices.filter(i => i.category === "Miscellaneous").length === 0 && (
                    <tr>
                      <td colSpan={5} className="px-6 py-12 text-center text-gray-500">
                        No miscellaneous invoices found.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const StatCard: React.FC<{
  title: string;
  value: string;
  subtitle?: string;
  trend?: string;
  trendUp?: boolean;
  icon: React.ReactNode;
}> = ({ title, value, subtitle, trend, trendUp, icon }) => (
  <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl">
    <div className="flex items-center justify-between mb-4">
      <div className="p-2 bg-white/5 rounded-lg border border-white/10">
        {icon}
      </div>
      {trend && (
        <div
          className={`flex items-center gap-1 text-sm font-medium ${trendUp ? "text-green-500" : "text-red-500"}`}
        >
          {trendUp ? (
            <ArrowUpRightIcon className="w-4 h-4" />
          ) : (
            <ArrowDownRightIcon className="w-4 h-4" />
          )}
          {trend}
        </div>
      )}
    </div>
    <h3 className="text-gray-400 text-sm font-medium">{title}</h3>
    <div className="mt-1 flex items-baseline gap-2">
      <span className="text-2xl font-bold text-white tracking-tight">
        {value}
      </span>
      {subtitle && (
        <span className="text-xs text-gray-500 uppercase tracking-wider">
          {subtitle}
        </span>
      )}
    </div>
  </div>
);

export default BillingDashboard;
