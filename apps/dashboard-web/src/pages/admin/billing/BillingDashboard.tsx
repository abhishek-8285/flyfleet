import {
  ArrowDownRightIcon,
  ArrowUpRightIcon,
  BanknotesIcon,
  CalendarDaysIcon,
  CreditCardIcon,
} from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUIStrings } from '../../../hooks/useUIStrings';
import billingService from '../../../services/billingService';
import { invoicingV2Service } from '../../../services/invoicingV2Service';
import { BillingMetrics, Invoice } from '../../../types/billing';
import SubscriptionPlans from './SubscriptionPlans';
import UsageOverview from './UsageOverview';

const SaaSSalesLogs: React.FC = () => {
  const [gateLogs, setGateLogs] = useState<any[]>([]);
  const [fintechTx, setFintechTx] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [logType, setLogType] = useState<'gate' | 'fintech'>('gate');
  const [filterText, setFilterText] = useState('');

  useEffect(() => {
    fetchLogs();
  }, []);

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const [gData, fData] = await Promise.all([
        billingService.getGateLogs(50, 0),
        billingService.getFintechTransactions(50, 0),
      ]);
      setGateLogs(gData || []);
      setFintechTx(fData || []);
    } catch (err) {
      console.error('Failed to fetch SaaS logs', err);
    } finally {
      setLoading(false);
    }
  };

  const totalCommissions = fintechTx.reduce((acc, tx) => acc + (tx.commission_earned || 0), 0);
  const totalSales = fintechTx.reduce((acc, tx) => acc + (tx.transaction_amount || 0), 0);

  const filteredGateLogs = gateLogs.filter(
    (log) =>
      log.gate_name?.toLowerCase().includes(filterText.toLowerCase()) ||
      log.vehicle_registration?.toLowerCase().includes(filterText.toLowerCase())
  );

  const filteredFintechTx = fintechTx.filter(
    (tx) =>
      tx.service_type?.toLowerCase().includes(filterText.toLowerCase()) ||
      tx.vehicle_registration?.toLowerCase().includes(filterText.toLowerCase())
  );

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      {/* Visual Analytics Hub */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-[#0b1329] border border-white/5 rounded-2xl p-6 relative overflow-hidden group hover:border-blue-500/30 transition-all">
          <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
            <span className="material-symbols-outlined text-6xl text-blue-500">point_of_sale</span>
          </div>
          <p className="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">Total SaaS Transactions</p>
          <p className="text-3xl font-extrabold text-white">{fintechTx.length + gateLogs.length}</p>
          <div className="mt-2 text-xs text-blue-400 flex items-center gap-1 font-medium">
            <span className="w-1.5 h-1.5 rounded-full bg-blue-500 animate-ping"></span>
            Real-time Operational Stream
          </div>
        </div>

        <div className="bg-[#0b1329] border border-white/5 rounded-2xl p-6 relative overflow-hidden group hover:border-emerald-500/30 transition-all">
          <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
            <span className="material-symbols-outlined text-6xl text-emerald-500">payments</span>
          </div>
          <p className="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">Commission Earned</p>
          <p className="text-3xl font-extrabold text-emerald-400">₹{totalCommissions.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</p>
          <div className="mt-2 text-xs text-emerald-400/70 flex items-center gap-1 font-medium">
            +3.5% Platform Margin Active
          </div>
        </div>

        <div className="bg-[#0b1329] border border-white/5 rounded-2xl p-6 relative overflow-hidden group hover:border-purple-500/30 transition-all">
          <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
            <span className="material-symbols-outlined text-6xl text-purple-500">wallet</span>
          </div>
          <p className="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">Fintech Gross Volume</p>
          <p className="text-3xl font-extrabold text-purple-400">₹{totalSales.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</p>
          <div className="mt-2 text-xs text-purple-400/70 flex items-center gap-1 font-medium">
            FASTag, Insurance & Challan Swipes
          </div>
        </div>

        <div className="bg-[#0b1329] border border-white/5 rounded-2xl p-6 relative overflow-hidden group hover:border-amber-500/30 transition-all">
          <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
            <span className="material-symbols-outlined text-6xl text-amber-500">sensor_door</span>
          </div>
          <p className="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">Gate Transitions (IN / OUT)</p>
          <p className="text-3xl font-extrabold text-amber-400">{gateLogs.length} Events</p>
          <div className="mt-2 text-xs text-amber-400/70 flex items-center gap-1 font-medium">
            Logistics Yards Security Scans
          </div>
        </div>
      </div>

      {/* Control bar */}
      <div className="flex flex-col md:flex-row justify-between items-center gap-4 bg-[#0b1329]/40 border border-white/5 p-4 rounded-2xl">
        <div className="flex gap-2 p-1 bg-black/40 rounded-xl border border-white/5">
          <button
            onClick={() => setLogType('gate')}
            className={`px-4 py-2 rounded-lg text-sm font-semibold transition-all flex items-center gap-2 ${
              logType === 'gate'
                ? 'bg-blue-600 text-white shadow-lg shadow-blue-600/10'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            Yard Gate Logs
          </button>
          <button
            onClick={() => setLogType('fintech')}
            className={`px-4 py-2 rounded-lg text-sm font-semibold transition-all flex items-center gap-2 ${
              logType === 'fintech'
                ? 'bg-blue-600 text-white shadow-lg shadow-blue-600/10'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            Fintech Commissions
          </button>
        </div>

        <div className="flex items-center gap-3 w-full md:w-auto">
          <div className="relative w-full md:w-80">
            <input
              type="text"
              placeholder={logType === 'gate' ? "Search gate name, registration..." : "Search service, registration..."}
              value={filterText}
              onChange={(e) => setFilterText(e.target.value)}
              className="w-full pl-4 pr-4 py-2 bg-black/40 border border-white/5 rounded-xl text-white placeholder:text-slate-600 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all"
            />
          </div>
          <button
            onClick={fetchLogs}
            className="p-2 bg-black/40 border border-white/5 rounded-xl text-slate-400 hover:text-white transition-all hover:bg-slate-900"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Grid List View */}
      {loading ? (
        <div className="text-center py-20 text-slate-500 font-light">
          Loading SaaS transaction streams...
        </div>
      ) : logType === 'gate' ? (
        <div className="bg-[#0b1329]/20 border border-white/5 rounded-2xl overflow-hidden shadow-2xl">
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-white/5 bg-black/40">
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Timestamp</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Gate Name</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Vehicle</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Action</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Cargo Status</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Weight</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {filteredGateLogs.map((log) => (
                  <tr key={log.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="px-6 py-4 text-sm text-slate-400 font-mono">
                      {new Date(log.created_at).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-sm text-slate-200 font-semibold">{log.gate_name}</td>
                    <td className="px-6 py-4 text-sm text-blue-400 font-mono">{log.vehicle_registration}</td>
                    <td className="px-6 py-4 text-sm">
                      <span className={`px-2.5 py-1 rounded-full text-xs font-bold border ${
                        log.action_type === 'IN'
                          ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                          : 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                      }`}>
                        {log.action_type}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-slate-300">{log.cargo_status || 'EMPTY'}</td>
                    <td className="px-6 py-4 text-sm text-white font-medium">{log.weight_measured ? `${log.weight_measured} Tons` : '--'}</td>
                  </tr>
                ))}
                {filteredGateLogs.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-12 text-center text-slate-500 font-light">
                      No yard gate logs found.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : (
        <div className="bg-[#0b1329]/20 border border-white/5 rounded-2xl overflow-hidden shadow-2xl">
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-white/5 bg-black/40">
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Timestamp</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Service Type</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Vehicle</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Tx Amount</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Commission (3%)</th>
                  <th className="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {filteredFintechTx.map((tx) => (
                  <tr key={tx.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="px-6 py-4 text-sm text-slate-400 font-mono">
                      {new Date(tx.created_at).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-sm font-semibold text-slate-200">
                      <span className={`px-2 py-0.5 rounded text-xs border ${
                        tx.service_type === 'FASTAG'
                          ? 'bg-blue-500/10 text-blue-400 border-blue-500/20'
                          : tx.service_type === 'INSURANCE'
                          ? 'bg-purple-500/10 text-purple-400 border-purple-500/20'
                          : 'bg-red-500/10 text-red-400 border-red-500/20'
                      }`}>
                        {tx.service_type}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-blue-400 font-mono">{tx.vehicle_registration}</td>
                    <td className="px-6 py-4 text-sm text-white font-medium">₹{tx.transaction_amount?.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                    <td className="px-6 py-4 text-sm text-emerald-400 font-bold">+₹{tx.commission_earned?.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
                    <td className="px-6 py-4 text-sm">
                      <span className={`px-2 py-0.5 rounded-full text-xs font-semibold border ${
                        tx.payment_status === 'SUCCESS'
                          ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                          : tx.payment_status === 'PENDING'
                          ? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                          : 'bg-rose-500/10 text-rose-400 border-rose-500/20'
                      }`}>
                        {tx.payment_status}
                      </span>
                    </td>
                  </tr>
                ))}
                {filteredFintechTx.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-12 text-center text-slate-500 font-light">
                      No fintech transactions found.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};

const BillingDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [history, setHistory] = useState<any[]>([]);
  const [metrics, setMetrics] = useState<BillingMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings('admin_billing');
  const [billingData, setBillingData] = useState<any>(null); // Assuming billingData is needed for new stat cards
  const [activeTab, setActiveTab] = useState('Overview'); // For new tab navigation

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
      const allInvoices = [...(invoicesData?.data || []), ...(v2InvoicesData?.invoices || [])].sort(
        (a, b) => new Date(b.invoice_date).getTime() - new Date(a.invoice_date).getTime()
      );

      setInvoices(allInvoices.slice(0, 10)); // Top 10
      setMetrics(metricsData);

      setBillingData({
        balance: metricsData?.pending_payment || 0,
        next_billing_date: settingsData?.auto_invoicing_enabled ? 'Auto-generated' : 'Manual',
        tier: settingsData?.base_currency || 'USD',
      });
    } catch (error) {
      console.error('Failed to fetch billing data', error);
    } finally {
      setLoading(false);
    }
  };

  const statusColors: Record<string, string> = {
    PAID: 'bg-green-500/10 text-green-500 border-green-500/20',
    PENDING: 'bg-amber-500/10 text-amber-500 border-amber-500/20',
    OVERDUE: 'bg-red-500/10 text-red-500 border-red-500/20',
    PARTIAL: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
    DRAFT: 'bg-gray-500/10 text-gray-500 border-gray-500/20',
    SENT: 'bg-purple-500/10 text-purple-500 border-purple-500/20',
    VIEWED: 'bg-cyan-500/10 text-cyan-500 border-cyan-500/20',
  };

  if (loading)
    return (
      <div className="p-8 text-center text-text-muted">
        {uiStrings?.loading || 'Loading billing data...'}
      </div>
    );

  return (
    <div className="p-8">
      {/* Header */}
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-white text-3xl font-bold mb-1">
            {uiStrings?.title || 'Billing & invoicing'}
          </h1>
          <p className="text-text-muted">
            {uiStrings?.subtitle || 'Manage payment methods, invoices, and service tiers'}
          </p>
        </div>
        <div className="flex gap-4">
          <button
            onClick={() => navigate('/admin/billing/new')}
            className="bg-white/10 text-white px-6 py-2 rounded-xl font-bold hover:bg-white/20 transition-colors"
          >
            {uiStrings?.btn_create_invoice || 'Create Invoice'}
          </button>
          {/* <button className="bg-primary text-white px-6 py-2 rounded-xl font-bold hover:bg-primary-hover transition-colors">
            {uiStrings?.btn_manage || "Manage Subscription"}
          </button> */}
        </div>
      </div>

      {/* New Stat Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-surface-dark border border-border-color rounded-2xl p-6">
          <div className="flex items-center gap-3 mb-4">
            <BanknotesIcon className="h-6 w-6 text-primary" />
            <span className="text-text-muted font-medium">
              {uiStrings?.stat_balance || 'Current Balance'}
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
              {uiStrings?.stat_next_bill || 'Next Bill Date'}
            </span>
          </div>
          <div className="text-3xl font-bold text-white">{billingData?.next_billing_date}</div>
        </div>

        <div className="bg-surface-dark border border-border-color rounded-2xl p-6">
          <div className="flex items-center gap-3 mb-4">
            <CreditCardIcon className="h-6 w-6 text-primary" />
            <span className="text-text-muted font-medium">
              {uiStrings?.stat_tier || 'Active Tier'}
            </span>
          </div>
          <div className="text-3xl font-bold text-white capitalize">{billingData?.tier}</div>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="flex gap-8 mb-8 border-b border-border-color overflow-x-auto">
        {[
          'Overview',
          'Usage',
          'Subscription Plans',
          'Invoice History',
          'SaaS & Sales Logs',
          'Miscellaneous',
          'Payment Methods',
        ].map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`pb-4 px-2 font-medium transition-all whitespace-nowrap ${
              activeTab === tab
                ? 'text-primary border-b-2 border-primary'
                : 'text-text-muted hover:text-white'
            }`}
          >
            {tab === 'Overview'
              ? uiStrings?.tab_overview || tab
              : tab === 'Usage'
                ? tab
                : tab === 'Subscription Plans'
                  ? tab
                  : tab === 'Invoice History'
                    ? uiStrings?.tab_history || tab
                    : tab === 'SaaS & Sales Logs'
                      ? tab
                      : tab === 'Miscellaneous'
                        ? 'Miscellaneous'
                        : tab === 'Payment Methods'
                          ? uiStrings?.tab_methods || tab
                          : tab}
          </button>
        ))}
      </div>

      {activeTab === 'Usage' && <UsageOverview />}
      {activeTab === 'Subscription Plans' && <SubscriptionPlans />}
      {activeTab === 'SaaS & Sales Logs' && <SaaSSalesLogs />}

      {activeTab === 'Overview' && (
        <>
          {/* New Stat Cards */}

          {/* Recent Invoices */}
          <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden">
            <div className="p-6 border-b border-white/10 flex items-center justify-between">
              <h2 className="text-lg font-semibold text-white">
                {uiStrings?.title_recent_invoices || 'Recent Invoices'}
              </h2>
              <button
                onClick={() => navigate('/admin/billing/invoices')}
                className="text-primary hover:underline text-sm font-medium"
              >
                {uiStrings?.btn_view_all || 'View All'}
              </button>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left">
                <thead className="bg-surface-darker border-b border-border-color">
                  <tr>
                    <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                      {uiStrings?.th_invoice_id || 'Invoice ID'}
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                      Category
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                      {uiStrings?.th_date || 'Date'}
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                      {uiStrings?.th_amount || 'Amount'}
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-bold text-text-muted uppercase tracking-widest">
                      {uiStrings?.th_status || 'Status'}
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
                        <span
                          className={`text-xs px-2 py-0.5 rounded border ${
                            invoice.category === 'Miscellaneous'
                              ? 'bg-purple-500/10 text-purple-400 border-purple-500/20'
                              : 'bg-white/5 text-gray-400 border-white/10'
                          }`}
                        >
                          {invoice.category || 'Transportation'}
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
                      <td colSpan={4} className="px-6 py-12 text-center text-gray-500">
                        {uiStrings?.msg_no_invoices || 'No recent invoices found.'}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {activeTab === 'Invoice History' && (
        <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden min-h-[400px]">
          <div className="p-12 text-center text-text-muted">
            History content will be paginated here.
          </div>
        </div>
      )}

      {activeTab === 'Miscellaneous' && (
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
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">
                      ID
                    </th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">
                      Customer
                    </th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">
                      Date
                    </th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">
                      Amount
                    </th>
                    <th className="px-6 py-4 text-xs font-bold text-text-muted uppercase tracking-widest">
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {invoices
                    .filter((i) => i.category === 'Miscellaneous')
                    .map((invoice) => (
                      <tr
                        key={invoice.id}
                        onClick={() => navigate(`/admin/billing/invoices/${invoice.id}`)}
                        className="hover:bg-white/5 transition-colors cursor-pointer"
                      >
                        <td className="px-6 py-4 text-white font-medium">
                          #{invoice.id.substring(0, 8).toUpperCase()}
                        </td>
                        <td className="px-6 py-4 text-gray-300">{invoice.customer_name}</td>
                        <td className="px-6 py-4 text-gray-400">
                          {new Date(invoice.invoice_date).toLocaleDateString()}
                        </td>
                        <td className="px-6 py-4 text-white font-semibold">
                          ₹{invoice.total_amount.toLocaleString('en-IN')}
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
                  {invoices.filter((i) => i.category === 'Miscellaneous').length === 0 && (
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
      <div className="p-2 bg-white/5 rounded-lg border border-white/10">{icon}</div>
      {trend && (
        <div
          className={`flex items-center gap-1 text-sm font-medium ${trendUp ? 'text-green-500' : 'text-red-500'}`}
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
      <span className="text-2xl font-bold text-white tracking-tight">{value}</span>
      {subtitle && (
        <span className="text-xs text-gray-500 uppercase tracking-wider">{subtitle}</span>
      )}
    </div>
  </div>
);

export default BillingDashboard;
