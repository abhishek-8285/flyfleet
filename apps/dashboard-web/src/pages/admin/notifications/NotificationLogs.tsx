import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { notificationService } from '../../../services/notificationService';
import { NotificationChannel, NotificationLog, NotificationStatus } from '../../../types/notification';

const NotificationLogs: React.FC = () => {
  const navigate = useNavigate();
  const [logs, setLogs] = useState<NotificationLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState({
    page: 1,
    limit: 10,
    status: '',
    channel: '',
    search: ''
  });

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const result = await notificationService.getLogs(filters);
      setLogs(result.data);
      setTotal(result.total);
    } catch (err) {
      console.error('Failed to fetch logs', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
  }, [filters]);

  const getChannelIcon = (channel: NotificationChannel) => {
    switch (channel) {
      case 'EMAIL': return <span className="material-symbols-outlined text-[#60a5fa]">mail</span>;
      case 'SMS': return <span className="material-symbols-outlined text-[#facc15]">sms</span>;
      case 'PUSH': return <span className="material-symbols-outlined text-[#f87171]">notifications_active</span>;
      case 'IN_APP': return <span className="material-symbols-outlined text-[#a78bfa]">app_registration</span>;
      default: return <span className="material-symbols-outlined">notifications_active</span>;
    }
  };

  const getStatusBadge = (status: NotificationStatus) => {
    const styles: Record<string, string> = {
      SENT: 'bg-green-900/40 text-green-400 border-green-500/50',
      FAILED: 'bg-red-900/40 text-red-400 border-red-500/50',
      QUEUED: 'bg-blue-900/40 text-blue-400 border-blue-500/50',
      READ: 'bg-purple-900/40 text-purple-400 border-purple-500/50',
      BOUNCED: 'bg-orange-900/40 text-orange-400 border-orange-500/50'
    };
    return (
      <span className={`px-2.5 py-1 rounded-full text-xs font-medium border ${styles[status] || styles['QUEUED']}`}>
        {status}
      </span>
    );
  };

  return (
    <div className="p-6 space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-400 to-indigo-400 bg-clip-text text-transparent flex items-center gap-2">
            <span className="material-symbols-outlined text-blue-400 text-[32px]">history</span>
            Notification Audit
          </h1>
          <p className="text-slate-400 mt-1">Monitor and govern security & system communications</p>
        </div>
        <div className="flex items-center gap-2">
          <button 
            onClick={fetchLogs}
            className="p-2.5 rounded-xl bg-slate-800/50 border border-slate-700 hover:bg-slate-700/50 transition-all flex items-center justify-center"
          >
            <span className="material-symbols-outlined text-slate-300">refresh</span>
          </button>
          <button 
            onClick={() => navigate('settings')}
            className="px-4 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white font-medium transition-all shadow-lg shadow-blue-900/20"
          >
            Manage Channels
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 p-5 rounded-2xl bg-slate-900/50 border border-slate-800 backdrop-blur-xl">
        <div className="relative">
          <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 text-[20px]">search</span>
          <input 
            type="text"
            placeholder="Search email, ID..."
            className="w-full pl-10 pr-4 py-2.5 rounded-xl bg-slate-950/50 border border-slate-700 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all text-slate-200"
            value={filters.search}
            onChange={(e) => setFilters({ ...filters, search: e.target.value, page: 1 })}
          />
        </div>
        <select 
          className="px-4 py-2.5 rounded-xl bg-slate-950/50 border border-slate-700 focus:border-blue-500 outline-none text-slate-200 appearance-none"
          value={filters.channel}
          onChange={(e) => setFilters({ ...filters, channel: e.target.value, page: 1 })}
        >
          <option value="">All Channels</option>
          <option value="EMAIL">Email</option>
          <option value="SMS">SMS</option>
          <option value="PUSH">Push</option>
          <option value="IN_APP">In-App</option>
        </select>
        <select 
          className="px-4 py-2.5 rounded-xl bg-slate-950/50 border border-slate-700 focus:border-blue-500 outline-none text-slate-200"
          value={filters.status}
          onChange={(e) => setFilters({ ...filters, status: e.target.value, page: 1 })}
        >
          <option value="">All Statuses</option>
          <option value="SENT">Sent</option>
          <option value="FAILED">Failed</option>
          <option value="QUEUED">Queued</option>
          <option value="BOUNCED">Bounced</option>
        </select>
        <div className="flex items-center gap-2">
          <button className="flex-1 px-4 py-2.5 rounded-xl border border-slate-700 hover:bg-slate-800 text-slate-300 transition-all flex items-center justify-center gap-2">
            <span className="material-symbols-outlined text-[20px]">filter_list</span> Filters
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="relative overflow-hidden rounded-2xl border border-slate-800 bg-slate-900/30 backdrop-blur-sm">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-slate-800 bg-slate-800/20">
              <th className="px-6 py-4 text-sm font-semibold text-slate-400 uppercase tracking-wider">Channel</th>
              <th className="px-6 py-4 text-sm font-semibold text-slate-400 uppercase tracking-wider">Recipient</th>
              <th className="px-6 py-4 text-sm font-semibold text-slate-400 uppercase tracking-wider">Subject</th>
              <th className="px-6 py-4 text-sm font-semibold text-slate-400 uppercase tracking-wider">Status</th>
              <th className="px-6 py-4 text-sm font-semibold text-slate-400 uppercase tracking-wider">Sent At</th>
              <th className="px-6 py-4 text-right"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/50">
            {loading ? (
              [...Array(5)].map((_, i) => (
                <tr key={i} className="animate-pulse">
                  <td colSpan={6} className="px-6 py-4 h-16 bg-slate-800/10"></td>
                </tr>
              ))
            ) : logs.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-12 text-center text-slate-500">
                  No notification logs found.
                </td>
              </tr>
            ) : logs.map((log) => (
              <tr 
                key={log.id} 
                className="hover:bg-slate-800/40 transition-colors group cursor-pointer"
                onClick={() => navigate(log.id)}
              >
                <td className="px-6 py-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded-lg bg-slate-800 border border-slate-700 group-hover:border-blue-500/50 transition-all">
                      {getChannelIcon(log.channel)}
                    </div>
                    <span className="text-sm font-medium text-slate-200">{log.channel}</span>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <div className="flex flex-col">
                    <span className="text-sm text-slate-200">{log.user_email}</span>
                    <span className="text-[10px] text-slate-500 font-mono mt-0.5">{log.id.slice(0, 8)}...</span>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <span className="text-sm text-slate-300 line-clamp-1">{log.subject}</span>
                </td>
                <td className="px-6 py-4">
                  {getStatusBadge(log.status)}
                </td>
                <td className="px-6 py-4 text-sm text-slate-400 font-mono">
                  {new Date(log.created_at).toLocaleString()}
                </td>
                <td className="px-6 py-4 text-right">
                  <span className="material-symbols-outlined text-slate-500 group-hover:text-blue-400 transition-all opacity-0 group-hover:opacity-100">visibility</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {/* Pagination */}
        <div className="px-6 py-4 border-t border-slate-800 bg-slate-800/20 flex items-center justify-between">
          <span className="text-sm text-slate-500">
            Showing <span className="text-slate-300">{logs.length}</span> of <span className="text-slate-300">{total}</span> logs
          </span>
          <div className="flex gap-2">
            <button 
              disabled={filters.page === 1}
              onClick={() => setFilters({ ...filters, page: filters.page - 1 })}
              className="px-4 py-2 rounded-lg bg-slate-800 border border-slate-700 text-slate-300 disabled:opacity-50 hover:bg-slate-700 transition-all text-sm"
            >
              Previous
            </button>
            <button 
              disabled={filters.page * filters.limit >= total}
              onClick={() => setFilters({ ...filters, page: filters.page + 1 })}
              className="px-4 py-2 rounded-lg bg-slate-800 border border-slate-700 text-slate-300 disabled:opacity-50 hover:bg-slate-700 transition-all text-sm"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default NotificationLogs;
