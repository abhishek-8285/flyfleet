import { ArrowDownTrayIcon, ArrowLeftIcon } from '@heroicons/react/24/outline';
import {
  AlertCircle,
  CheckCircle2,
  ChevronLeft,
  Clock,
  CreditCard,
  IndianRupee,
} from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useUIStrings } from '../../../hooks/useUIStrings';
import billingService from '../../../services/billingService';
import { invoicingV2Service } from '../../../services/invoicingV2Service';

const InvoiceDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [invoice, setInvoice] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings('admin_billing');
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [paymentAmount, setPaymentAmount] = useState(0);
  const [billingMeta, setBillingMeta] = useState<any>(null);

  useEffect(() => {
    fetchInvoice();
  }, [id]);

  const fetchInvoice = async () => {
    try {
      // Try invoicingV2 first as it's the newer module
      try {
        const data = await invoicingV2Service.getInvoice(id!);
        setInvoice(data);
        setPaymentAmount(data?.balance_due || 0);
      } catch (e) {
        // Fallback to legacy billing service
        const data = await billingService.getInvoice(id!);
        setInvoice(data);
        setPaymentAmount(data?.balance_due || 0);
      }
    } catch (error) {
      console.error('Failed to fetch invoice from both services', error);
    }

    try {
      const res = await fetch('/api/v1/ui/global_config');
      const data = await res.json();
      setBillingMeta(data.billing_metadata);
    } catch (error) {
      console.error('Failed to fetch billing metadata', error);
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

  const handleDownloadPDF = async () => {
    try {
      let blob: Blob | null = null;
      try {
        // Try V2 service first
        blob = await invoicingV2Service.downloadPDF(id!);
      } catch (e) {
        console.warn('V2 PDF download failed, falling back to legacy', e);
        // Fallback to legacy service
        blob = await billingService.downloadPDF(id!);
      }

      if (!blob || !(blob instanceof Blob)) {
        throw new Error('Invalid PDF data received from server');
      }

      if (!blob || (blob as any).size === 0) {
        throw new Error('Invalid PDF data received from server');
      }

      const url = window.URL.createObjectURL(blob as Blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `invoice-${invoice?.invoice_number || id}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('Failed to download PDF from both services', error);
      alert('Failed to download PDF. Please try again or contact support.');
    }
  };

  const statusIcons = {
    PAID: <CheckCircle2 className="w-4 h-4 text-green-500" />,
    PENDING: <Clock className="w-4 h-4 text-amber-500" />,
    OVERDUE: <AlertCircle className="w-4 h-4 text-red-500" />,
    PARTIAL: <Clock className="w-4 h-4 text-blue-500" />,
    DRAFT: <Clock className="w-4 h-4 text-gray-500" />,
  };

  const statusColors = {
    PAID: 'bg-green-500/10 text-green-500 border-green-500/20',
    PENDING: 'bg-amber-500/10 text-amber-500 border-amber-500/20',
    OVERDUE: 'bg-red-500/10 text-red-500 border-red-500/20',
    PARTIAL: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
    DRAFT: 'bg-gray-500/10 text-gray-400 border-gray-500/20',
  };

  if (loading)
    return (
      <div className="p-8 text-center text-text-muted">
        {uiStrings?.loading_invoice || 'Loading invoice details...'}
      </div>
    );

  const formatDate = (dateStr: string) => {
    if (!dateStr) return 'N/A';
    const d = new Date(dateStr);
    return isNaN(d.getTime()) ? 'N/A' : d.toLocaleDateString(undefined, { dateStyle: 'long' });
  };

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Header Actions */}
      <div className="p-8 max-w-4xl mx-auto">
        <button
          onClick={() => navigate('/admin/billing')}
          className="flex items-center text-gray-400 hover:text-white transition-colors mb-8 group"
        >
          <ArrowLeftIcon className="h-4 w-4 mr-2 group-hover:-translate-x-1 transition-transform" />
          {uiStrings?.btn_back_billing || 'Back to Billing'}
        </button>

        <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl overflow-hidden">
          <div className="bg-gray-900/50 px-8 py-8 border-b border-white/5 flex justify-between items-start">
            <div>
              <div className="flex items-center gap-3 mb-2">
                <h2 className="text-white text-3xl font-black">
                   {uiStrings?.invoice_title || 'Invoice'} #{invoice?.invoice_number || invoice?.id?.substring(0, 8)}
                </h2>
                <span className={`px-3 py-1 rounded-full text-[10px] font-black border ${invoice?.template_type === 'TRIP_NOTE' ? 'bg-blue-500/10 text-blue-400 border-blue-400/20' : 'bg-purple-500/10 text-purple-400 border-purple-400/20'}`}>
                   {invoice?.template_type?.replace('_', ' ') || 'TAX INVOICE'}
                </span>
              </div>
              <p className="text-gray-500 font-medium">
                {uiStrings?.th_date || 'Date'}:{' '}
                {formatDate(invoice?.invoice_date)}
              </p>
            </div>
            <button
              onClick={handleDownloadPDF}
              className="bg-primary/10 text-primary px-6 py-2 rounded-xl font-bold hover:bg-primary/20 transition-colors flex items-center gap-2"
            >
              <ArrowDownTrayIcon className="h-5 w-5" />
              {uiStrings?.btn_download || 'Download PDF'}
            </button>
          </div>
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            {invoice?.status === 'DRAFT' && (
              <button
                onClick={async () => {
                  try {
                    await billingService.updateInvoice(id!, { ...invoice, status: 'PENDING' });
                    fetchInvoice();
                  } catch (err) {
                    console.error('Failed to send invoice', err);
                  }
                }}
                className="flex items-center gap-2 px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-500 transition-colors shadow-lg shadow-green-900/20"
              >
                <CheckCircle2 className="w-4 h-4" /> {uiStrings?.btn_send_invoice || 'Send Invoice'}
              </button>
            )}
            {invoice?.status !== 'PAID' && invoice?.status !== 'DRAFT' && (
              <button
                onClick={() => setShowPaymentModal(true)}
                className="flex items-center gap-2 px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors shadow-lg shadow-primary/20"
              >
                <CreditCard className="w-4 h-4" />{' '}
                {uiStrings?.btn_record_payment || 'Record Payment'}
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
                    <h2 className="text-3xl font-bold text-white tracking-tight leading-none uppercase mb-2">
                      {uiStrings?.invoice_title || 'Invoice'}
                    </h2>
                    <p className="text-gray-500 font-mono text-sm tracking-widest">
                      #{invoice?.id?.toUpperCase()}
                    </p>
                  </div>
                  <div
                    className={`flex items-center gap-2 px-4 py-1.5 rounded-full border ${statusColors[invoice?.status as keyof typeof statusColors] || ''}`}
                  >
                    {statusIcons[invoice?.status as keyof typeof statusIcons]}
                    <span className="text-sm font-bold uppercase tracking-wider">
                      {invoice?.status}
                    </span>
                  </div>
                </div>

                <div className="grid grid-cols-2 lg:grid-cols-4 gap-8">
                  <div className="space-y-1">
                    <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em]">
                      {uiStrings?.label_issued_on || 'Issued On'}
                    </p>
                    <p className="text-white font-medium">
                      {formatDate(invoice?.invoice_date)}
                    </p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em]">
                      {uiStrings?.label_due_by || 'Due By'}
                    </p>
                    <p className="text-white font-medium">
                      {formatDate(invoice?.due_date)}
                    </p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em]">
                      {uiStrings?.label_customer || 'Customer'}
                    </p>
                    <p className="text-white font-medium">{invoice?.customer_name}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em]">
                      {uiStrings?.label_currency || 'Currency'}
                    </p>
                    <p className="text-white font-medium">{invoice?.currency}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em]">
                      Category
                    </p>
                    <p
                      className={`text-sm font-bold uppercase tracking-wider ${
                        invoice?.category === 'Miscellaneous' ? 'text-purple-400' : 'text-white'
                      }`}
                    >
                      {invoice?.category || 'Transportation'}
                    </p>
                  </div>
                </div>

                {/* Logistics Highlight */}
                {invoice?.vehicle_no && (
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-6 p-6 bg-white/5 rounded-2xl border border-white/5">
                    <div>
                       <p className="text-[9px] font-black text-gray-500 uppercase tracking-widest mb-1">Vehicle No</p>
                       <p className="text-white font-bold">{invoice.vehicle_no}</p>
                    </div>
                    <div>
                       <p className="text-[9px] font-black text-gray-500 uppercase tracking-widest mb-1">Driver</p>
                       <p className="text-white font-bold">{invoice.driver_name || 'N/A'}</p>
                    </div>
                    <div className="col-span-2">
                       <p className="text-[9px] font-black text-gray-500 uppercase tracking-widest mb-1">Route</p>
                       <p className="text-white font-bold text-sm">
                          {invoice.pickup_location || 'N/A'} <span className="text-primary mx-1">→</span> {invoice.drop_location || 'N/A'}
                       </p>
                    </div>
                  </div>
                )}
              </div>

              <div className="px-8 py-6 overflow-x-auto">
                <table className="w-full text-left">
                  <thead className="bg-white/5 border-b border-white/5">
                    <tr>
                      <th className="px-6 py-4 text-left text-xs font-bold text-gray-500 uppercase tracking-widest">
                        {uiStrings?.th_desc || 'Description'}
                      </th>
                      <th className="pb-4 text-[10px] font-bold text-gray-500 uppercase tracking-widest text-center">
                        {uiStrings?.th_qty || 'Qty'}
                      </th>
                      <th className="pb-4 text-[10px] font-bold text-gray-500 uppercase tracking-widest text-right">
                        {uiStrings?.th_unit_price || 'Unit Price'}
                      </th>
                      <th className="pb-4 text-[10px] font-bold text-gray-500 uppercase tracking-widest text-right">
                        {uiStrings?.th_tax || 'Tax'}
                      </th>
                      <th className="px-6 py-4 text-right text-xs font-bold text-gray-500 uppercase tracking-widest">
                        {uiStrings?.th_amount || 'Amount'}
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/5">
                    {invoice?.items?.map((item: any, idx: number) => (
                      <tr key={idx}>
                        <td className="py-4 text-gray-300 font-medium">{item?.description}</td>
                        <td className="py-4 text-gray-400 text-center">{item?.quantity}</td>
                        <td className="py-4 text-gray-400 text-right">
                          ₹{item?.unit_price?.toLocaleString('en-IN')}
                        </td>
                        <td className="py-4 text-gray-400 text-right">{item?.tax_percent}%</td>
                        <td className="py-4 text-white font-semibold text-right">
                          ₹{item?.total?.toLocaleString('en-IN')}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div className="p-8 bg-white/5 border-t border-white/5 flex flex-col md:flex-row justify-between gap-6">
                <div className="flex-1">
                  <p className="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em] mb-2">
                    {uiStrings?.label_notes_terms || 'Notes & Terms'}
                  </p>
                  <p className="text-sm text-gray-400 max-w-sm">
                    {uiStrings?.desc_notes_terms?.replace(
                      '{late_fee}',
                      billingMeta?.late_fee_percent || '--'
                    ) ||
                      `Please make the payment by the due date mentioned above. Interest of ${billingMeta?.late_fee_percent || '--'}% may be applicable for late payments.`}
                  </p>
                </div>
                <div className="p-8 bg-white/5 flex justify-end">
                  <div className="w-64 space-y-4">
                    <div className="flex justify-between items-center text-gray-400">
                      <span>{uiStrings?.invoice_subtotal || 'Subtotal'}</span>
                      <span className="font-mono">
                        ₹{invoice?.subtotal?.toLocaleString('en-IN')}
                      </span>
                    </div>
                    <div className="flex justify-between items-center text-gray-400">
                      <span>{uiStrings?.invoice_tax || 'Tax'}</span>
                      <span className="font-mono">
                        ₹{invoice?.tax_amount?.toLocaleString('en-IN') || '0.00'}
                      </span>
                    </div>
                    <div className="pt-4 border-t border-white/10 flex justify-between items-center">
                      <span className="text-white font-bold">
                        {uiStrings?.invoice_total || 'Total'}
                      </span>
                      <span className="text-white text-2xl font-black font-mono">
                        ₹{invoice?.total_amount?.toLocaleString('en-IN')}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Info Sidebar */}
          <div className="space-y-6">
            <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl space-y-6">
              <h3 className="text-lg font-semibold text-white">
                {uiStrings?.title_payment_summary || 'Payment Summary'}
              </h3>
              <div className="space-y-4">
                <div className="p-4 bg-white/5 rounded-xl border border-white/5 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-green-500/10 rounded-lg">
                      <Clock className="w-4 h-4 text-green-500" />
                    </div>
                    <span className="text-sm font-medium text-gray-300">
                      {uiStrings?.label_amount_paid || 'Amount Paid'}
                    </span>
                  </div>
                  <span className="text-white font-bold">
                    ₹{invoice?.amount_paid?.toLocaleString('en-IN')}
                  </span>
                </div>
                <div className="p-4 bg-white/5 rounded-xl border border-white/5 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-amber-500/10 rounded-lg">
                      <IndianRupee className="w-4 h-4 text-amber-500" />
                    </div>
                    <span className="text-sm font-medium text-gray-300">
                      {uiStrings?.label_balance_due || 'Balance Due'}
                    </span>
                  </div>
                  <span className="text-white font-bold text-lg">
                    ₹{invoice?.balance_due?.toLocaleString('en-IN')}
                  </span>
                </div>
              </div>
            </div>

            <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl space-y-4">
              <h3 className="text-lg font-semibold text-white">
                {uiStrings?.title_history || 'History'}
              </h3>
              <div className="space-y-6">
                <HistoryItem
                  title={uiStrings?.hist_created || 'Invoice Created'}
                  date={invoice?.created_at ? new Date(invoice.created_at).toLocaleString() : 'N/A'}
                  active={true}
                />
                {invoice?.amount_paid > 0 && (
                  <HistoryItem
                    title={uiStrings?.hist_payment || 'Payment Recorded'}
                    date={invoice?.updated_at ? new Date(invoice.updated_at).toLocaleString() : 'N/A'}
                    active={true}
                    color="green"
                  />
                )}
                {invoice?.status === 'OVERDUE' && (
                  <HistoryItem
                    title={uiStrings?.hist_overdue || 'Marked Overdue'}
                    date={uiStrings?.hist_automated || 'Automated System'}
                    color="red"
                    active={true}
                  />
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Payment Modal */}
        {showPaymentModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
            <div className="bg-gray-900 border border-white/10 w-full max-w-md rounded-2xl overflow-hidden shadow-2xl animate-in fade-in zoom-in duration-200">
              <div className="p-6 border-b border-white/5 flex items-center justify-between">
                <h3 className="text-xl font-bold text-white">
                  {uiStrings?.title_record_payment || 'Record Payment'}
                </h3>
                <button
                  onClick={() => setShowPaymentModal(false)}
                  className="p-1 hover:bg-white/10 rounded-lg transition-colors"
                >
                  <ChevronLeft className="w-6 h-6 rotate-90" />
                </button>
              </div>
              <div className="p-6 space-y-6">
                <div className="space-y-2">
                  <label className="text-sm font-medium text-gray-400">
                    {uiStrings?.label_payment_amount || 'Payment Amount (₹)'}
                  </label>
                  <input
                    type="number"
                    value={paymentAmount}
                    onChange={(e) => setPaymentAmount(parseFloat(e.target.value) || 0)}
                    className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-2xl font-bold text-white focus:outline-none focus:border-primary transition-colors"
                  />
                  <p className="text-xs text-gray-500 italic">
                    {uiStrings?.msg_remaining_balance || 'Remaining balance:'} ₹
                    {invoice?.balance_due?.toLocaleString() || '0'}
                  </p>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <button
                    onClick={() => setPaymentAmount(invoice?.balance_due || 0)}
                    className="px-4 py-2 bg-white/5 text-gray-300 rounded-lg text-sm hover:bg-white/10 transition-colors"
                  >
                    {uiStrings?.btn_full_amount || 'Full Amount'}
                  </button>
                  <button
                    onClick={() => setPaymentAmount((invoice?.balance_due || 0) / 2)}
                    className="px-4 py-2 bg-white/5 text-gray-300 rounded-lg text-sm hover:bg-white/10 transition-colors"
                  >
                    {uiStrings?.btn_split_half || 'Split Half'}
                  </button>
                </div>
              </div>
              <div className="p-6 bg-white/5 flex items-center gap-3">
                <button
                  onClick={() => setShowPaymentModal(false)}
                  className="flex-1 px-4 py-3 bg-gray-800 text-gray-300 rounded-xl hover:bg-gray-700 transition-colors font-semibold"
                >
                  {uiStrings?.btn_cancel || 'Cancel'}
                </button>
                <button
                  onClick={handleRecordPayment}
                  className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary/90 transition-all font-bold shadow-lg shadow-primary/20"
                >
                  {uiStrings?.btn_confirm_payment || 'Confirm Payment'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

const HistoryItem: React.FC<{
  title: string;
  date: string;
  active?: boolean;
  color?: 'green' | 'amber' | 'red';
}> = ({ title, date, active, color }) => (
  <div className="flex gap-4 relative">
    <div
      className={`mt-1 z-10 w-2.5 h-2.5 rounded-full ring-4 ring-gray-900 ${
        color === 'green'
          ? 'bg-green-500'
          : color === 'red'
            ? 'bg-red-500'
            : color === 'amber'
              ? 'bg-amber-500'
              : 'bg-primary'
      }`}
    ></div>
    <div className="space-y-1">
      <p className={`text-sm font-semibold ${active ? 'text-white' : 'text-gray-500'}`}>{title}</p>
      <p className="text-xs text-gray-500 font-mono italic">{date}</p>
    </div>
  </div>
);

export default InvoiceDetail;
