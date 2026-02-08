import {
    AlertCircle,
    CheckCircle2,
    ChevronLeft,
    Clock,
    CreditCard,
    Download,
    IndianRupee,
    Printer
} from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import billingService from '../../../services/billingService';
import { Invoice } from '../../../types/billing';

const InvoiceDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [invoice, setInvoice] = useState<Invoice | null>(null);
  const [loading, setLoading] = useState(true);
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [paymentAmount, setPaymentAmount] = useState(0);

  useEffect(() => {
    fetchInvoice();
  }, [id]);

  const fetchInvoice = async () => {
    try {
      const data = await billingService.getInvoice(id!);
      setInvoice(data);
      setPaymentAmount(data.balance_due);
    } catch (error) {
      console.error('Failed to fetch invoice', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRecordPayment = async () => {
    try {
      await billingService.recordPayment(id!, paymentAmount);
      setShowPaymentModal(false);
      fetchInvoice();
    } catch (error) {
      console.error('Failed to record payment', error);
    }
  };

  const statusIcons = {
    PAID: <CheckCircle2 className="w-4 h-4 text-green-500" />,
    PENDING: <Clock className="w-4 h-4 text-amber-500" />,
    OVERDUE: <AlertCircle className="w-4 h-4 text-red-500" />,
    PARTIAL: <Clock className="w-4 h-4 text-blue-500" />,
  };

  const statusColors = {
    PAID: 'bg-green-500/10 text-green-500 border-green-500/20',
    PENDING: 'bg-amber-500/10 text-amber-500 border-amber-500/20',
    OVERDUE: 'bg-red-500/10 text-red-500 border-red-500/20',
    PARTIAL: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
  };

  if (loading || !invoice) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Header Actions */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <button 
          onClick={() => navigate('/admin/billing')}
          className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors"
        >
          <ChevronLeft className="w-5 h-5" />
          Back to Billing
        </button>
        <div className="flex items-center gap-2">
          <button className="flex items-center gap-2 px-4 py-2 bg-gray-800 text-gray-300 rounded-lg hover:bg-gray-700 transition-colors">
            <Printer className="w-4 h-4" /> Print
          </button>
          <button className="flex items-center gap-2 px-4 py-2 bg-gray-800 text-gray-300 rounded-lg hover:bg-gray-700 transition-colors">
            <Download className="w-4 h-4" /> Download PDF
          </button>
          {invoice.status !== 'PAID' && (
            <button 
              onClick={() => setShowPaymentModal(true)}
              className="flex items-center gap-2 px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors shadow-lg shadow-primary/20"
            >
              <CreditCard className="w-4 h-4" /> Record Payment
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Invoice Card */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden">
            <div className="p-8 border-b border-white/5 space-y-6">
              <div className="flex justify-between items-start">
                <div>
                  <h2 className="text-3xl font-bold text-white tracking-tight leading-none uppercase mb-2">Invoice</h2>
                  <p className="text-gray-500 font-mono text-sm tracking-widest">#{invoice.id.toUpperCase()}</p>
                </div>
                <div className={`flex items-center gap-2 px-4 py-1.5 rounded-full border ${statusColors[invoice.status]}`}>
                  {statusIcons[invoice.status]}
                  <span className="text-sm font-bold uppercase tracking-wider">{invoice.status}</span>
                </div>
              </div>

              <div className="grid grid-cols-2 lg:grid-cols-4 gap-8">
                <div className="space-y-1">
                  <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em]">Issued On</p>
                  <p className="text-white font-medium">{new Date(invoice.invoice_date).toLocaleDateString(undefined, { dateStyle: 'long' })}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em]">Due By</p>
                  <p className="text-white font-medium">{new Date(invoice.due_date).toLocaleDateString(undefined, { dateStyle: 'long' })}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em]">Customer</p>
                  <p className="text-white font-medium">{invoice.customer_name}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em]">Currency</p>
                  <p className="text-white font-medium">{invoice.currency}</p>
                </div>
              </div>
            </div>

            <div className="px-8 py-6 overflow-x-auto">
              <table className="w-full text-left">
                <thead>
                  <tr className="border-b border-white/5">
                    <th className="pb-4 text-[10px] font-bold text-gray-500 uppercase tracking-widest">Description</th>
                    <th className="pb-4 text-[10px] font-bold text-gray-500 uppercase tracking-widest text-center">Qty</th>
                    <th className="pb-4 text-[10px] font-bold text-gray-500 uppercase tracking-widest text-right">Unit Price</th>
                    <th className="pb-4 text-[10px] font-bold text-gray-500 uppercase tracking-widest text-right">Tax</th>
                    <th className="pb-4 text-[10px] font-bold text-gray-500 uppercase tracking-widest text-right">Total</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {invoice.items?.map((item, idx) => (
                    <tr key={idx}>
                      <td className="py-4 text-gray-300 font-medium">{item.description}</td>
                      <td className="py-4 text-gray-400 text-center">{item.quantity}</td>
                      <td className="py-4 text-gray-400 text-right">₹{item.unit_price.toLocaleString()}</td>
                      <td className="py-4 text-gray-400 text-right">{item.tax_percent}%</td>
                      <td className="py-4 text-white font-semibold text-right">₹{item.total.toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="p-8 bg-white/5 border-t border-white/5 flex flex-col md:flex-row justify-between gap-6">
              <div className="flex-1">
                <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em] mb-2">Notes & Terms</p>
                <p className="text-sm text-gray-400 max-w-sm">Please make the payment by the due date mentioned above. Interest of 2% may be applicable for late payments.</p>
              </div>
              <div className="w-full md:w-64 space-y-3">
                <div className="flex justify-between items-center text-gray-400 font-medium">
                  <span className="text-xs uppercase tracking-wider">Subtotal</span>
                  <span>₹{(invoice.total_amount / 1.18).toFixed(2)}</span>
                </div>
                <div className="flex justify-between items-center text-gray-400 font-medium">
                  <span className="text-xs uppercase tracking-wider">Tax (18%)</span>
                  <span>₹{(invoice.total_amount - (invoice.total_amount / 1.18)).toFixed(2)}</span>
                </div>
                <div className="pt-3 border-t border-white/10 flex justify-between items-center">
                  <span className="text-white font-bold uppercase tracking-widest">Total</span>
                  <span className="text-2xl font-black text-white">₹{invoice.total_amount.toLocaleString()}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Info Sidebar */}
        <div className="space-y-6">
          <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl space-y-6">
            <h3 className="text-lg font-semibold text-white">Payment Summary</h3>
            <div className="space-y-4">
              <div className="p-4 bg-white/5 rounded-xl border border-white/5 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-green-500/10 rounded-lg"><Clock className="w-4 h-4 text-green-500" /></div>
                  <span className="text-sm font-medium text-gray-300">Amount Paid</span>
                </div>
                <span className="text-white font-bold">₹{invoice.amount_paid.toLocaleString()}</span>
              </div>
              <div className="p-4 bg-white/5 rounded-xl border border-white/5 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-amber-500/10 rounded-lg"><IndianRupee className="w-4 h-4 text-amber-500" /></div>
                  <span className="text-sm font-medium text-gray-300">Balance Due</span>
                </div>
                <span className="text-white font-bold text-lg">₹{invoice.balance_due.toLocaleString()}</span>
              </div>
            </div>
          </div>

          <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl space-y-4">
            <h3 className="text-lg font-semibold text-white">History</h3>
            <div className="space-y-6">
              <HistoryItem title="Invoice Created" date={new Date(invoice.created_at).toLocaleString()} active={true} />
              {invoice.amount_paid > 0 && <HistoryItem title="Payment Recorded" date={new Date(invoice.updated_at).toLocaleString()} active={true} color="green" />}
              {invoice.status === 'OVERDUE' && <HistoryItem title="Marked Overdue" date="Automated System" color="red" active={true} />}
            </div>
          </div>
        </div>
      </div>

      {/* Payment Modal */}
      {showPaymentModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
          <div className="bg-gray-900 border border-white/10 w-full max-w-md rounded-2xl overflow-hidden shadow-2xl animate-in fade-in zoom-in duration-200">
            <div className="p-6 border-b border-white/5 flex items-center justify-between">
              <h3 className="text-xl font-bold text-white">Record Payment</h3>
              <button 
                onClick={() => setShowPaymentModal(false)}
                className="p-1 hover:bg-white/10 rounded-lg transition-colors"
                ><ChevronLeft className="w-6 h-6 rotate-90" />
              </button>
            </div>
            <div className="p-6 space-y-6">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-400">Payment Amount (₹)</label>
                <input 
                  type="number" 
                  value={paymentAmount}
                  onChange={(e) => setPaymentAmount(parseFloat(e.target.value) || 0)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-2xl font-bold text-white focus:outline-none focus:border-primary transition-colors"
                />
                <p className="text-xs text-gray-500 italic">Remaining balance: ₹{invoice.balance_due.toLocaleString()}</p>
              </div>
              
              <div className="grid grid-cols-2 gap-3">
                <button 
                  onClick={() => setPaymentAmount(invoice.balance_due)}
                  className="px-4 py-2 bg-white/5 text-gray-300 rounded-lg text-sm hover:bg-white/10 transition-colors"
                >
                  Full Amount
                </button>
                <button 
                  onClick={() => setPaymentAmount(invoice.balance_due / 2)}
                  className="px-4 py-2 bg-white/5 text-gray-300 rounded-lg text-sm hover:bg-white/10 transition-colors"
                >
                  Split Half
                </button>
              </div>
            </div>
            <div className="p-6 bg-white/5 flex items-center gap-3">
              <button 
                onClick={() => setShowPaymentModal(false)}
                className="flex-1 px-4 py-3 bg-gray-800 text-gray-300 rounded-xl hover:bg-gray-700 transition-colors font-semibold"
              >
                Cancel
              </button>
              <button 
                onClick={handleRecordPayment}
                className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary/90 transition-all font-bold shadow-lg shadow-primary/20"
              >
                Confirm Payment
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const HistoryItem: React.FC<{ title: string; date: string; active?: boolean; color?: 'green' | 'amber' | 'red' }> = ({ title, date, active, color }) => (
  <div className="flex gap-4 relative">
    <div className={`mt-1 z-10 w-2.5 h-2.5 rounded-full ring-4 ring-gray-900 ${
      color === 'green' ? 'bg-green-500' : 
      color === 'red' ? 'bg-red-500' : 
      color === 'amber' ? 'bg-amber-500' : 'bg-primary'
    }`}></div>
    <div className="space-y-1">
      <p className={`text-sm font-semibold ${active ? 'text-white' : 'text-gray-500'}`}>{title}</p>
      <p className="text-xs text-gray-500 font-mono italic">{date}</p>
    </div>
  </div>
);

export default InvoiceDetail;
