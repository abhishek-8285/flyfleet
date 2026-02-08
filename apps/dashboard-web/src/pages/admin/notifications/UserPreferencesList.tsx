import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { notificationService } from '../../../services/notificationService';
import { UserNotificationPreference } from '../../../types/notification';

const UserPreferencesList: React.FC = () => {
  const navigate = useNavigate();
  const [prefs, setPrefs] = useState<UserNotificationPreference[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState({
    page: 1,
    limit: 10,
    search: ''
  });

  const fetchPrefs = async () => {
    setLoading(true);
    try {
      const result = await notificationService.getPreferences(filters);
      setPrefs(result.items);
      setTotal(result.total);
    } catch (err) {
      console.error('Failed to fetch preferences', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPrefs();
  }, [filters]);

  const ChannelToggle = ({ enabled, channel }: { enabled: boolean, channel: string }) => {
    const icons = {
      EMAIL: 'mail',
      SMS: 'sms',
      PUSH: 'notifications_active',
      IN_APP: 'app_registration'
    };
    
    return (
      <div className={`p-1.5 rounded-lg border flex items-center gap-1.5 transition-all ${
        enabled 
          ? 'bg-blue-500/10 border-blue-500/50 text-blue-400' 
          : 'bg-slate-800/50 border-slate-700 text-slate-500 grayscale'
      }`}>
        <span className="material-symbols-outlined text-[16px]">{icons[channel as keyof typeof icons]}</span>
        <span className="text-[10px] font-bold uppercase">{channel}</span>
      </div>
    );
  };

  return (
    <div className="p-6 space-y-6 animate-in fade-in slide-in-from-bottom-2 duration-500">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-400 to-indigo-400 bg-clip-text text-transparent flex items-center gap-2">
            <span className="material-symbols-outlined text-blue-400 text-[32px]">manage_accounts</span>
            User Preferences
          </h1>
          <p className="text-slate-400 mt-1">Manage granular communication channels per user</p>
        </div>
        <div className="flex items-center gap-2">
          <button 
            onClick={fetchPrefs}
            className="p-2.5 rounded-xl bg-slate-800/50 border border-slate-700 hover:bg-slate-700/50 transition-all text-slate-300 flex items-center justify-center"
          >
            <span className="material-symbols-outlined">refresh</span>
          </button>
          <button 
            onClick={() => navigate('global')}
            className="px-4 py-2.5 rounded-xl bg-slate-800/50 border border-slate-700 hover:border-blue-500/50 text-slate-200 font-medium transition-all flex items-center gap-2"
          >
            <span className="material-symbols-outlined text-[20px]">settings</span>
            Global Defaults
          </button>
        </div>
      </div>

      {/* Search & Filters */}
      <div className="flex flex-col md:flex-row gap-4 p-5 rounded-2xl bg-slate-900/50 border border-slate-800 backdrop-blur-xl">
        <div className="relative flex-1">
          <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 text-[20px]">search</span>
          <input 
            type="text"
            placeholder="Search by User Name or ID..."
            className="w-full pl-10 pr-4 py-2.5 rounded-xl bg-slate-950/50 border border-slate-700 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all text-slate-200"
            value={filters.search}
            onChange={(e) => setFilters({ ...filters, search: e.target.value, page: 1 })}
          />
        </div>
        <button className="px-6 py-2.5 rounded-xl border border-slate-700 hover:bg-slate-800 text-slate-300 transition-all flex items-center justify-center gap-2">
          <span className="material-symbols-outlined text-[20px]">filter_list</span> Filters
        </button>
      </div>

      {/* Preferences Table */}
      <div className="relative overflow-hidden rounded-2xl border border-slate-800 bg-slate-900/30 backdrop-blur-sm shadow-2xl">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-slate-800 bg-slate-800/20">
              <th className="px-6 py-4 text-sm font-semibold text-slate-400 uppercase tracking-wider">User</th>
              <th className="px-6 py-4 text-sm font-semibold text-slate-400 uppercase tracking-wider">Enabled Channels</th>
              <th className="px-6 py-4 text-sm font-semibold text-slate-400 uppercase tracking-wider">Last Sync</th>
              <th className="px-6 py-4 text-right"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/50">
            {loading ? (
              [...Array(5)].map((_, i) => (
                <tr key={i} className="animate-pulse">
                  <td colSpan={4} className="px-6 py-4 h-16 bg-slate-800/5"></td>
                </tr>
              ))
            ) : prefs.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-6 py-12 text-center text-slate-500">
                  No user preferences found.
                </td>
              </tr>
            ) : prefs.map((pref) => (
              <tr 
                key={pref.user_id} 
                className="hover:bg-slate-400/5 transition-all group cursor-pointer border-l-2 border-transparent hover:border-blue-500"
                onClick={() => navigate(pref.user_id)}
              >
                <td className="px-6 py-5">
                  <div className="flex flex-col">
                    <span className="text-sm font-semibold text-slate-100">{pref.user_name || 'System User'}</span>
                    <span className="text-xs text-slate-500 mt-0.5">{pref.user_email || pref.user_id}</span>
                  </div>
                </td>
                <td className="px-6 py-5">
                  <div className="flex gap-2">
                    <ChannelToggle enabled={pref.email_enabled} channel="EMAIL" />
                    <ChannelToggle enabled={pref.sms_enabled} channel="SMS" />
                    <ChannelToggle enabled={pref.push_enabled} channel="PUSH" />
                    <ChannelToggle enabled={pref.in_app_enabled} channel="IN_APP" />
                  </div>
                </td>
                <td className="px-6 py-5 text-sm text-slate-400">
                  {pref.last_modified_at ? new Date(pref.last_modified_at).toLocaleDateString() : 'Initial State'}
                </td>
                <td className="px-6 py-5 text-right">
                  <span className="material-symbols-outlined text-slate-600 group-hover:text-blue-400 transition-all opacity-0 group-hover:opacity-100">visibility</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {/* Pagination placeholder */}
        <div className="px-6 py-4 border-t border-slate-800/50 bg-slate-800/10 flex items-center justify-between">
          <span className="text-xs text-slate-500 font-medium">
            DISPLAYING <span className="text-slate-300">{prefs.length}</span> OF <span className="text-slate-300">{total}</span> USERS
          </span>
          <div className="flex gap-2">
            <button 
              disabled={filters.page === 1}
              onClick={() => setFilters({ ...filters, page: filters.page - 1 })}
              className="px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-slate-400 disabled:opacity-30 hover:bg-slate-700 transition-all text-xs font-bold"
            >
              PREV
            </button>
            <button 
              disabled={filters.page * filters.limit >= total}
              onClick={() => setFilters({ ...filters, page: filters.page + 1 })}
              className="px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-slate-400 disabled:opacity-30 hover:bg-slate-700 transition-all text-xs font-bold"
            >
              NEXT
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default UserPreferencesList;
