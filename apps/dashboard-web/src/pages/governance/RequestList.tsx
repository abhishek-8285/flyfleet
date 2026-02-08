import { format } from 'date-fns';
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import GovernanceService from '../../services/GovernanceService';
import { GovernanceRequestListResponse } from '../../types/governance';

const RequestList: React.FC = () => {
  const navigate = useNavigate();
  const [data, setData] = useState<GovernanceRequestListResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    const fetchRequests = async () => {
      try {
        const resp = await GovernanceService.listRequests({ page: 1, page_size: 20 });
        setData(resp);
      } catch (error) {
        console.error('Failed to fetch governance requests:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchRequests();
  }, []);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'PENDING': return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20';
      case 'APPROVED': return 'bg-green-500/10 text-green-500 border-green-500/20';
      case 'REJECTED': return 'bg-red-500/10 text-red-500 border-red-500/20';
      case 'COMPLETED': return 'bg-blue-500/10 text-blue-500 border-blue-500/20';
      default: return 'bg-gray-500/10 text-gray-500 border-gray-500/20';
    }
  };

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 flex-shrink-0 z-10">
        <div className="flex flex-wrap items-center gap-3">
          <div className="text-xl font-semibold text-white tracking-tight flex items-center gap-2">
            <span className="material-symbols-outlined text-primary">gavel</span>
            Data Governance
          </div>
          <div className="h-6 w-px bg-border-color mx-2 hidden md:block"></div>
        </div>
        
        <div className="flex items-center gap-3 ml-auto">
          <button 
            onClick={() => navigate('/admin/governance/settings')}
            className="p-1.5 text-text-muted hover:text-white hover:bg-surface-dark rounded transition-colors"
            title="Governance Settings"
          >
            <span className="material-symbols-outlined text-[20px]">settings</span>
          </button>
          <div className="h-4 w-px bg-border-color"></div>
          <button 
            onClick={() => navigate('/admin/governance/requests/new')}
            className="bg-primary hover:bg-primary/90 text-white text-xs font-medium py-1.5 px-4 rounded-full flex items-center gap-2 shadow-[0_4px_14px_rgba(19,127,236,0.39)] transition-all duration-200"
          >
            <span className="material-symbols-outlined text-[18px]">add</span>
            New Request
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6 bg-[radial-gradient(circle_at_top_right,rgba(19,127,236,0.05),transparent)]">
        <div className="max-w-[1600px] mx-auto pb-6 space-y-6">
          <div className="bg-surface-dark/50 backdrop-blur-md border border-border-color rounded-2xl overflow-hidden shadow-xl">
            <div className="px-6 py-4 border-b border-border-color flex items-center justify-between bg-surface-darker/30">
              <h3 className="text-sm font-medium text-white">Access & Governance Requests</h3>
              <div className="relative">
                <input 
                  type="text" 
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Search requests..."
                  className="bg-background-dark border border-border-color text-white text-xs rounded-full pl-9 pr-4 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary w-64 outline-none transition-all focus:w-80"
                />
                <span className="material-symbols-outlined absolute left-3 top-1.5 text-text-muted text-sm">search</span>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="bg-surface-darker/50">
                    <th className="px-6 py-4 text-[10px] font-semibold text-text-muted uppercase tracking-wider">Type</th>
                    <th className="px-6 py-4 text-[10px] font-semibold text-text-muted uppercase tracking-wider">Purpose</th>
                    <th className="px-6 py-4 text-[10px] font-semibold text-text-muted uppercase tracking-wider">Status</th>
                    <th className="px-6 py-4 text-[10px] font-semibold text-text-muted uppercase tracking-wider">Requested At</th>
                    <th className="px-6 py-4 text-[10px] font-semibold text-text-muted uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border-color">
                  {loading ? (
                    <tr>
                      <td colSpan={5} className="px-6 py-12 text-center text-text-muted italic">Loading requests...</td>
                    </tr>
                  ) : data?.requests.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="px-6 py-12 text-center text-text-muted italic">No governance requests found.</td>
                    </tr>
                  ) : (
                    data?.requests.map((req) => (
                      <tr 
                        key={req.id} 
                        className="hover:bg-white/[0.02] transition-colors cursor-pointer"
                        onClick={() => navigate(`/admin/governance/requests/${req.id}`)}
                      >
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <span className="p-2 rounded-lg bg-primary/10 text-primary material-symbols-outlined text-[20px]">
                              {req.request_type === 'EXPORT' ? 'download' : req.request_type === 'DELETION' ? 'delete' : 'policy'}
                            </span>
                            <span className="text-sm text-white font-medium">{req.request_type.replace('_', ' ')}</span>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="text-sm text-text-muted line-clamp-1 max-w-md">{req.purpose}</div>
                        </td>
                        <td className="px-6 py-4">
                          <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold border ${getStatusColor(req.status)}`}>
                            {req.status}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <div className="text-xs text-text-muted">
                            {format(new Date(req.created_at), 'MMM dd, yyyy HH:mm')}
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <button className="text-primary hover:text-primary-light text-xs font-medium">View Details</button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RequestList;
