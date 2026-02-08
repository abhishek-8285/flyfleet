import { format } from 'date-fns';
import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import GovernanceService from '../../services/GovernanceService';
import { GovernanceRequest } from '../../types/governance';

const RequestDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [request, setRequest] = useState<GovernanceRequest | null>(null);
  const [loading, setLoading] = useState(true);
  const [notes, setNotes] = useState('');
  const [processing, setProcessing] = useState(false);

  useEffect(() => {
    const fetchRequest = async () => {
      if (!id) return;
      try {
        const resp = await GovernanceService.getRequest(id);
        setRequest(resp);
      } catch (error) {
        console.error('Failed to fetch request:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchRequest();
  }, [id]);

  const handleAction = async (action: 'approve' | 'reject') => {
    if (!id) return;
    setProcessing(true);
    try {
      if (action === 'approve') {
        await GovernanceService.approveRequest(id, notes);
      } else {
        await GovernanceService.rejectRequest(id, notes);
      }
      // Refresh
      const resp = await GovernanceService.getRequest(id);
      setRequest(resp);
    } catch (error) {
      console.error(`Failed to ${action} request:`, error);
    } finally {
      setProcessing(false);
    }
  };

  if (loading) return <div className="p-8 text-white">Loading...</div>;
  if (!request) return <div className="p-8 text-white">Request not found.</div>;

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex items-center justify-between gap-4 flex-shrink-0 z-10">
        <div className="flex items-center gap-3">
          <button 
            onClick={() => navigate('/admin/governance')}
            className="p-1.5 text-text-muted hover:text-white hover:bg-surface-dark rounded transition-colors"
          >
            <span className="material-symbols-outlined text-[20px]">arrow_back</span>
          </button>
          <div className="text-lg font-semibold text-white tracking-tight">Request Details</div>
        </div>
        <div className="text-xs text-text-muted font-mono bg-surface-dark px-3 py-1 rounded-full border border-border-color">
          ID: {request.id}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6 bg-[radial-gradient(circle_at_top_right,rgba(19,127,236,0.05),transparent)]">
        <div className="max-w-4xl mx-auto space-y-6">
          <div className="grid grid-cols-12 gap-6">
            <div className="col-span-8 space-y-6">
              <div className="bg-surface-dark/50 backdrop-blur-md border border-border-color rounded-2xl p-6 shadow-xl space-y-6">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-4">
                    <span className="p-4 rounded-2xl bg-primary/10 text-primary material-symbols-outlined text-[32px]">
                      {request.request_type === 'EXPORT' ? 'download' : request.request_type === 'DELETION' ? 'delete' : 'policy'}
                    </span>
                    <div>
                      <h2 className="text-xl font-bold text-white uppercase tracking-tight">{request.request_type.replace('_', ' ')}</h2>
                      <p className="text-sm text-text-muted">Requested by Admin on {format(new Date(request.created_at), 'MMMM dd, yyyy')}</p>
                    </div>
                  </div>
                  <span className={`px-4 py-1.5 rounded-full text-xs font-bold border ${
                    request.status === 'PENDING' ? 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20' :
                    request.status === 'APPROVED' ? 'bg-green-500/10 text-green-500 border-green-500/20' :
                    'bg-red-500/10 text-red-500 border-red-500/20'
                  }`}>
                    {request.status}
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-8 border-t border-border-color pt-6">
                  <div className="space-y-1">
                    <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest">Format</p>
                    <p className="text-sm text-white">{request.format}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest">Delivery Method</p>
                    <p className="text-sm text-white">{request.delivery_method}</p>
                  </div>
                </div>

                <div className="space-y-4 border-t border-border-color pt-6">
                  <div>
                    <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest mb-2">Detailed Scope</p>
                    <div className="bg-background-dark/50 rounded-xl p-4 text-sm text-text-muted leading-relaxed">
                      {request.data_scope}
                    </div>
                  </div>
                  <div>
                    <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest mb-2">Business Purpose</p>
                    <div className="bg-background-dark/50 rounded-xl p-4 text-sm text-text-muted leading-relaxed">
                      {request.purpose}
                    </div>
                  </div>
                </div>
              </div>

              {request.status === 'PENDING' && (
                <div className="bg-surface-dark/50 backdrop-blur-md border border-border-color rounded-2xl p-6 shadow-xl space-y-4">
                  <h3 className="text-sm font-bold text-white uppercase tracking-tight">Review Action</h3>
                  <textarea 
                    value={notes}
                    onChange={(e) => setNotes(e.target.value)}
                    placeholder="Add internal notes for approval or rejection reason..."
                    className="w-full bg-background-dark border border-border-color text-white rounded-xl px-4 py-3 min-h-[100px] outline-none focus:border-primary transition-colors text-sm"
                  />
                  <div className="flex gap-4">
                    <button 
                      onClick={() => handleAction('reject')}
                      disabled={processing}
                      className="flex-1 py-3 px-6 rounded-xl border border-red-500/50 text-red-500 font-medium hover:bg-red-500/10 transition-colors disabled:opacity-50"
                    >
                      Reject Request
                    </button>
                    <button 
                      onClick={() => handleAction('approve')}
                      disabled={processing}
                      className="flex-1 py-3 px-6 rounded-xl bg-green-600 text-white font-semibold hover:bg-green-500 shadow-[0_4px_14px_rgba(22,163,74,0.39)] transition-all disabled:opacity-50"
                    >
                      Approve & Execute
                    </button>
                  </div>
                </div>
              )}
            </div>

            <div className="col-span-4 space-y-6">
              <div className="bg-surface-dark/50 backdrop-blur-md border border-border-color rounded-2xl p-6 shadow-xl">
                <h3 className="text-xs font-bold text-white uppercase tracking-widest mb-4">Request Timeline</h3>
                <div className="space-y-6 relative before:absolute before:left-2 before:top-2 before:bottom-2 before:w-px before:bg-border-color">
                  <div className="relative pl-8">
                    <div className="absolute left-0 top-1 w-4 h-4 rounded-full bg-primary border-4 border-background-dark"></div>
                    <p className="text-xs font-bold text-white uppercase tracking-tight">Request Created</p>
                    <p className="text-[10px] text-text-muted">{format(new Date(request.created_at), 'MMM dd, HH:mm')}</p>
                  </div>
                  {request.status !== 'PENDING' && (
                    <div className="relative pl-8">
                      <div className={`absolute left-0 top-1 w-4 h-4 rounded-full border-4 border-background-dark ${request.status === 'APPROVED' ? 'bg-green-500' : 'bg-red-500'}`}></div>
                      <p className="text-xs font-bold text-white uppercase tracking-tight">
                        {request.status === 'APPROVED' ? 'Approved' : 'Rejected'}
                      </p>
                      <p className="text-[10px] text-text-muted">{format(new Date(request.updated_at), 'MMM dd, HH:mm')}</p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RequestDetail;
