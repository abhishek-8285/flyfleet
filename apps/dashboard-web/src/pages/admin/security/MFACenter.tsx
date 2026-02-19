import {
  AdjustmentsVerticalIcon,
  ArrowPathIcon,
  DevicePhoneMobileIcon,
  ExclamationTriangleIcon,
  KeyIcon,
  MagnifyingGlassIcon,
  NoSymbolIcon,
  ShieldCheckIcon,
  UserGroupIcon
} from '@heroicons/react/24/outline';
import React, { useEffect, useMemo, useState } from 'react';
import SecurityFooter from '../../../components/SecurityFooter';
import SecuritySidebar from '../../../components/SecuritySidebar';
import api from '../../../services/api';

const toArray = (value: any) => {
  if (Array.isArray(value)) return value;
  if (Array.isArray(value?.data)) return value.data;
  if (Array.isArray(value?.users)) return value.users;
  if (Array.isArray(value?.events)) return value.events;
  return [];
};

const toNumber = (value: any) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
};

const toRelativeTime = (value: string) => {
  if (!value) return 'Never';
  const timestamp = new Date(value).getTime();
  if (!Number.isFinite(timestamp)) return value;

  const diffMinutes = Math.max(0, Math.floor((Date.now() - timestamp) / 60000));
  if (diffMinutes < 1) return 'Just now';
  if (diffMinutes < 60) return `${diffMinutes} min${diffMinutes === 1 ? '' : 's'} ago`;
  const diffHours = Math.floor(diffMinutes / 60);
  if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
};

const MFACenter: React.FC = () => {
  const [users, setUsers] = useState<any[]>([]);
  const [summary, setSummary] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');

  const orgId = localStorage.getItem('org_id') || 'default';

  useEffect(() => {
    fetchMFAUsers();
  }, [orgId]);

  const fetchMFAUsers = async () => {
    setLoading(true);
    try {
      const [eventsRes, dashboardRes, usersRes] = await Promise.all([
        api.getMFAEvents(orgId, 1, 200),
        api.getMFADashboard(orgId).catch(() => null),
        api.getUsers(orgId).catch(() => []),
      ]);

      const userDirectory = toArray(usersRes).map((u: any, idx: number) => ({
        id: u.id || u.user_id || `user-${idx}`,
        email: u.email || '',
        name: u.name || u.username || u.email || 'Unknown User',
        status: 'DISABLED',
        method: 'NONE',
        devices: 0,
        lastUsed: 'Never',
      }));

      const userById = new Map<string, any>();
      const userByEmail = new Map<string, any>();
      userDirectory.forEach((u: any) => {
        userById.set(String(u.id), u);
        if (u.email) userByEmail.set(String(u.email).toLowerCase(), u);
      });

      const merged = new Map<string, any>();
      const ensureUser = (idKey: string, fallback: any) => {
        if (!merged.has(idKey)) {
          merged.set(idKey, {
            id: fallback.id || idKey,
            email: fallback.email || '',
            name: fallback.name || fallback.email || 'Unknown User',
            status: fallback.status || 'DISABLED',
            method: fallback.method || 'NONE',
            devices: toNumber(fallback.devices),
            lastUsed: fallback.lastUsed || 'Never',
            lastTimestamp: 0,
          });
        }
        return merged.get(idKey);
      };

      userDirectory.forEach((u: any) => {
        const key = String(u.id || u.email);
        ensureUser(key, u);
      });

      const eventRows = toArray(eventsRes);
      eventRows.forEach((event: any) => {
        const eventUserId = String(event.user_id || event.userId || '');
        const eventEmail = String(event.email || event.user_email || event.details?.email || '').toLowerCase();
        const directoryUser = userById.get(eventUserId) || userByEmail.get(eventEmail);
        const key = String(directoryUser?.id || eventUserId || eventEmail || event.id);

        const row = ensureUser(key, {
          id: directoryUser?.id || eventUserId || event.id,
          email: directoryUser?.email || eventEmail || '',
          name: directoryUser?.name || event.details?.name || event.user_name || eventEmail || 'Unknown User',
          status: 'DISABLED',
          method: 'NONE',
          devices: directoryUser?.devices || 0,
          lastUsed: 'Never',
        });

        const status = String(event.status || '').toUpperCase();
        if (status) {
          row.status = status === 'SUCCESS' ? 'ENABLED' : status === 'FAILED' ? 'ENFORCED' : status;
        }

        if (event.method) row.method = String(event.method).toUpperCase();
        row.devices = Math.max(row.devices, toNumber(event.details?.device_count || event.details?.trusted_devices));

        const createdAt = event.created_at || event.createdAt;
        if (createdAt) {
          const ts = new Date(createdAt).getTime();
          if (Number.isFinite(ts) && ts > row.lastTimestamp) {
            row.lastTimestamp = ts;
            row.lastUsed = toRelativeTime(createdAt);
          }
        }
      });

      setSummary(dashboardRes || null);
      setUsers(Array.from(merged.values()).map(({ lastTimestamp, ...rest }) => rest));
    } catch (error) {
      console.error('Failed to fetch MFA users', error);
      setSummary(null);
      setUsers([]);
    } finally {
      setLoading(false);
    }
  };

  const filteredUsers = useMemo(() => {
    const term = searchTerm.trim().toLowerCase();
    if (!term) return users;
    return users.filter((u) =>
      [u.name, u.email, u.status, u.method].some((value) =>
        String(value || '').toLowerCase().includes(term)
      )
    );
  }, [searchTerm, users]);

  const totalUsers = toNumber(summary?.total_users) || users.length;
  const enabledUsers = toNumber(summary?.mfa_enabled) || users.filter((u) => u.status === 'ENABLED' || u.status === 'ENFORCED').length;
  const enabledPct = totalUsers > 0 ? Math.round((enabledUsers / totalUsers) * 100) : 0;
  const trustedDevices = users.reduce((sum, u) => sum + toNumber(u.devices), 0);
  const policyViolations = toNumber(summary?.locked_accounts || summary?.failed_attempts_24h);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ENABLED': return <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold bg-emerald-500/10 text-vibrant-green border border-emerald-500/20"><ShieldCheckIcon className="w-3 h-3"/> ACTIVE</span>;
      case 'ENFORCED': return <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold bg-blue-500/10 text-vibrant-blue border border-blue-500/20"><KeyIcon className="w-3 h-3"/> ENFORCED</span>;
      case 'DISABLED': return <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold bg-red-500/10 text-vibrant-red border border-red-500/20"><NoSymbolIcon className="w-3 h-3"/> INACTIVE</span>;
      default: return <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold bg-slate-500/10 text-slate-400 border border-slate-500/20">BYPASSED</span>;
    }
  };

  return (
    <div className="flex h-screen bg-[#0b1117] text-slate-200 overflow-hidden font-inter">
      <SecuritySidebar />

      <main className="flex-1 flex flex-col min-w-0 bg-grid-pattern relative">
        <div className="glow-effect" />

        <header className="px-8 py-6 border-b border-white/5 bg-slate-900/40 backdrop-blur-md z-10 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-white flex items-center gap-3">
              <span className="material-symbols-outlined text-primary text-3xl">fingerprint</span>
              MFA Command Center
            </h2>
            <p className="text-slate-400 text-sm mt-1">Multi-factor identity directory and global enforcement policies.</p>
          </div>

          <div className="flex items-center gap-3">
            <button className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-hover text-white rounded-lg font-semibold transition-all shadow-lg shadow-primary/20">
              <AdjustmentsVerticalIcon className="w-5 h-5" />
              Global Policy
            </button>
          </div>
        </header>

        <div className="px-8 py-6 grid grid-cols-4 gap-6 z-10">
          {[
            { label: 'Total Users', value: totalUsers.toLocaleString(), icon: UserGroupIcon, color: 'text-blue-400' },
            { label: 'MFA Enabled', value: `${enabledPct}%`, icon: ShieldCheckIcon, color: 'text-emerald-400' },
            { label: 'Trusted Devices', value: trustedDevices.toLocaleString(), icon: DevicePhoneMobileIcon, color: 'text-purple-400' },
            { label: 'Policy Violations', value: policyViolations.toLocaleString(), icon: ExclamationTriangleIcon, color: 'text-amber-400' },
          ].map((stat, i) => (
            <div key={i} className="glass-card p-4 rounded-xl flex items-center gap-4">
              <div className={`p-3 rounded-lg bg-slate-800/50 ${stat.color}`}>
                <stat.icon className="w-6 h-6" />
              </div>
              <div>
                <p className="text-xs text-slate-500 font-medium uppercase tracking-wider">{stat.label}</p>
                <p className="text-2xl font-bold text-white">{stat.value}</p>
              </div>
            </div>
          ))}
        </div>

        <div className="flex-1 px-8 pb-8 flex flex-col min-h-0 z-10">
          <div className="glass-card rounded-2xl flex flex-col flex-1 overflow-hidden">
            <div className="p-4 border-b border-white/5 bg-white/5 flex items-center justify-between">
              <div className="relative w-96">
                <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                <input
                  type="text"
                  placeholder="Search by identity or email..."
                  className="w-full bg-slate-900/50 border border-white/10 rounded-lg py-2 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              <div className="flex items-center gap-2">
                <button
                  className="p-2 text-slate-400 hover:text-white transition-colors"
                  onClick={fetchMFAUsers}
                  aria-label="Refresh MFA users"
                >
                  <ArrowPathIcon className="w-5 h-5" />
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-auto custom-scrollbar">
              {loading ? (
                <div className="h-full min-h-[280px] flex items-center justify-center text-slate-400 text-sm">
                  Loading MFA users...
                </div>
              ) : (
                <table className="w-full text-left">
                  <thead>
                    <tr className="border-b border-white/5 bg-white/[0.02]">
                      <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Identity</th>
                      <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">MFA Status</th>
                      <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Primary Method</th>
                      <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Trusted Assets</th>
                      <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Last Activity</th>
                      <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/5">
                    {filteredUsers.map((u) => (
                      <tr key={u.id} className="hover:bg-white/[0.02] transition-colors group">
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <div className="w-10 h-10 rounded-lg bg-slate-800 flex items-center justify-center border border-white/5 shadow-inner">
                              <UserGroupIcon className="w-5 h-5 text-slate-400" />
                            </div>
                            <div className="flex flex-col">
                              <span className="text-sm font-semibold text-white">{u.name}</span>
                              <span className="text-xs text-slate-500 font-mono tracking-tighter">{u.email || '-'}</span>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          {getStatusBadge(u.status)}
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-2 text-slate-300">
                            {u.method === 'TOTP' ? <ShieldCheckIcon className="w-4 h-4 text-primary" /> : <DevicePhoneMobileIcon className="w-4 h-4 text-slate-500" />}
                            <span className="text-sm font-medium">{u.method}</span>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <span className="text-sm text-slate-300 font-bold">{toNumber(u.devices)} Devices</span>
                        </td>
                        <td className="px-6 py-4">
                          <span className="text-xs text-slate-400 italic">{u.lastUsed}</span>
                        </td>
                        <td className="px-6 py-4 text-right">
                          <button className="text-primary hover:text-white text-xs font-bold uppercase tracking-widest transition-colors opacity-0 group-hover:opacity-100 px-4 py-2 bg-primary/10 rounded-lg">
                            Manage
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default MFACenter;
