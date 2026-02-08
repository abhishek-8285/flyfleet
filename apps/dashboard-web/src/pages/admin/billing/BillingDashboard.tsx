import {
    AlertCircle,
    ArrowDownRight,
    ArrowUpRight,
    CheckCircle2,
    Clock,
    IndianRupee,
    MoreVertical,
    Plus
} from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import billingService from '../../../services/billingService';
import { BillingMetrics, Invoice } from '../../../types/billing';

const BillingDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [metrics, setMetrics] = useState<BillingMetrics | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [invoicesRes, metricsRes] = await Promise.all([
        billingService.getInvoices(1, 5),
        billingService.getMetrics()
      ]);
      setInvoices(invoicesRes.data);
      setMetrics(metricsRes);
    } catch (error) {
      console.error('Failed to fetch billing data', error);
    } finally {
      setLoading(status => false);
    }
  };

  const statusColors = {
    PAID: 'bg-green-500/10 text-green-500 border-green-500/20',
    PENDING: 'bg-amber-500/10 text-amber-500 border-amber-500/20',
    OVERDUE: 'bg-red-500/10 text-red-500 border-red-500/20',
    PARTIAL: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Billing & Invoices</h1>
          <p className="text-gray-400">Manage your organization's financials and billing history.</p>
        </div>
        <div className="flex items-center gap-3">
          <button 
            onClick={() => navigate('/admin/billing/settings')}
            className="px-4 py-2 bg-gray-800 text-gray-300 rounded-lg hover:bg-gray-700 transition-colors"
          >
            Settings
          </button>
          <button 
            onClick={() => navigate('/admin/billing/new')}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors"
          >
            <Plus className="w-4 h-4" />
            Create Invoice
          </button>
        </div>
      </div>

      {/* Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard 
          title="Total Invoiced (MTD)" 
          value={`₹${metrics?.total_invoiced_mtd.toLocaleString()}`}
          trend="+12.5%"
          trendUp={true}
          icon={<IndianRupee className="w-5 h-5" />}
        />
        <MetricCard 
          title="Pending Payment" 
          value={`₹${metrics?.pending_payment.toLocaleString()}`}
          subtitle={`${metrics?.open_count} Open Invoices`}
          icon={<Clock className="w-5 h-5 text-amber-500" />}
        />
        <MetricCard 
          title="Overdue Amount" 
          value={`₹${metrics?.overdue_amount.toLocaleString()}`}
          subtitle="Requires attention"
          icon={<AlertCircle className="w-5 h-5 text-red-500" />}
        />
        <MetricCard 
          title="Paid (Last 30 Days)" 
          value={`${metrics?.paid_count}`}
          subtitle="Invoices settled"
          icon={<CheckCircle2 className="w-5 h-5 text-green-500" />}
        />
      </div>

      {/* Recent Invoices */}
      <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden">
        <div className="p-6 border-b border-white/10 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-white">Recent Invoices</h2>
          <button 
            onClick={() => navigate('/admin/billing/invoices')}
            className="text-primary hover:underline text-sm font-medium"
          >
            View All
          </button>
        </div>
        
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="bg-white/5">
                <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Invoice ID</th>
                <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Customer</th>
                <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Date</th>
                <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Amount</th>
                <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {invoices.map((invoice) => (
                <tr key={invoice.id} className="hover:bg-white/5 transition-colors cursor-pointer" onClick={() => navigate(`/admin/billing/invoices/${invoice.id}`)}>
                  <td className="px-6 py-4">
                    <span className="text-white font-medium">#{invoice.id.substring(0, 8).toUpperCase()}</span>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-gray-300">{invoice.customer_name}</span>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-gray-400">{new Date(invoice.invoice_date).toLocaleDateString()}</span>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-white font-semibold">₹{invoice.total_amount.toLocaleString()}</span>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2.5 py-1 rounded-full text-xs font-medium border ${statusColors[invoice.status]}`}>
                      {invoice.status}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <button className="p-1 hover:bg-gray-700 rounded transition-colors">
                      <MoreVertical className="w-4 h-4 text-gray-400" />
                    </button>
                  </td>
                </tr>
              ))}
              {invoices.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center text-gray-500">
                    No recent invoices found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

const MetricCard: React.FC<{ 
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
        <div className={`flex items-center gap-1 text-sm font-medium ${trendUp ? 'text-green-500' : 'text-red-500'}`}>
          {trendUp ? <ArrowUpRight className="w-4 h-4" /> : <ArrowDownRight className="w-4 h-4" />}
          {trend}
        </div>
      )}
    </div>
    <h3 className="text-gray-400 text-sm font-medium">{title}</h3>
    <div className="mt-1 flex items-baseline gap-2">
      <span className="text-2xl font-bold text-white tracking-tight">{value}</span>
      {subtitle && <span className="text-xs text-gray-500 uppercase tracking-wider">{subtitle}</span>}
    </div>
  </div>
);

export default BillingDashboard;
