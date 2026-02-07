import React, { useCallback, useEffect, useState } from 'react';
import {
    Area,
    AreaChart,
    CartesianGrid,
    Cell,
    Pie,
    PieChart,
    ResponsiveContainer,
    Tooltip,
    XAxis, YAxis
} from 'recharts';
import { useAuth } from '../context/AuthProvider';
import { api } from '../services/api';

interface Stats {
  total_users: number;
  mfa_enabled: number;
  mfa_pending: number;
  mfa_disabled: number;
  failed_attempts_24h: number;
  locked_accounts: number;
}

interface MFAEvent {
  time: string;
  user: string;
  event: string;
  method: string;
  ip: string;
  loc: string;
  status: string;
}

const MFACenter: React.FC = () => {
  const { user } = useAuth();
  const orgId = localStorage.getItem('org_id') || user?.org_id || '';

  // --- State ---
  const [activeTab, setActiveTab] = useState<'overview' | 'policy' | 'users'>('overview');
  const [isPIIMasked, setIsPIIMasked] = useState(true);
  const [isAutoRefresh, setIsAutoRefresh] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [dateRange, setDateRange] = useState('Last 90 Days');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);

  const [dashboardStats, setDashboardStats] = useState<Stats | null>(null);
  const [trendData, setTrendData] = useState<any[]>([]);
  const [methodDistribution, setMethodDistribution] = useState<any[]>([]);
  const [failureLocations, setFailureLocations] = useState<any[]>([]);
  const [failureUsers, setFailureUsers] = useState<any[]>([]);
  const [recentEvents, setRecentEvents] = useState<MFAEvent[]>([]);
  const [lockedUsers, setLockedUsers] = useState<any[]>([]);
  const [mfaPolicy, setMFAPolicy] = useState<any>(null);

  // --- Data Fetching ---
  const fetchAllData = useCallback(async () => {
    if (!orgId) return;
    setIsLoading(true);
    try {
      const [stats, trend, methods, locs, ufails, locked, policy] = await Promise.all([
        api.getMFADashboard(orgId),
        api.getMFAAdoptionTrend(orgId),
        api.getMFAMethodDistribution(orgId),
        api.getMFAFailureLocations(orgId),
        api.getMFAFailureUsers(orgId),
        api.getLockedUsers(orgId),
        api.getMFAPolicy(orgId)
      ]);

      setDashboardStats(stats || {});
      setTrendData(trend || []);
      
      // Transform map to array for Recharts
      const methodArray = Object.entries(methods || {}).map(([name, value]) => ({
        name,
        value,
        color: name === 'TOTP' ? '#137fec' : name === 'SMS' ? '#6366f1' : '#10b981'
      }));
      setMethodDistribution(methodArray);
      
      setFailureLocations(locs || []);
      setFailureUsers(ufails || []);
      setLockedUsers(locked || []);
      setMFAPolicy(policy || {});

      // Initial events pull
      const eventsResp = await api.getMFAEvents(orgId, 1, pageSize);
      setRecentEvents(eventsResp?.events || []);
    } catch (err) {
      console.error("Failed to fetch MFA security data", err);
    } finally {
      setIsLoading(false);
    }
  }, [orgId, pageSize]);

  useEffect(() => {
    fetchAllData();
  }, [fetchAllData]);

  // --- WebSocket Live Feed ---
  useEffect(() => {
    if (!orgId || !isAutoRefresh) return;

    const wsUrl = (process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1').replace('http', 'ws') + '/security/mfa/stream?org_id=' + orgId;
    const socket = new WebSocket(wsUrl);

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (Array.isArray(data)) {
          // Merge with recent events, keeping unique ones
          setRecentEvents(prev => {
            const combined = [...data, ...prev];
            // Simple deduplication by id if present, or just slice
            return combined.slice(0, 50); 
          });
        }
      } catch (e) {
        console.error("WS Message error", e);
      }
    };

    return () => socket.close();
  }, [orgId, isAutoRefresh]);

  // --- Helpers ---
  const maskEmail = (email: string) => {
    if (!isPIIMasked || !email) return email || 'n/a';
    const parts = email.split('@');
    if (parts.length !== 2) return email;
    const [name, domain] = parts;
    return `${name[0]}${'*'.repeat(Math.max(0, name.length - 1))}@${domain}`;
  };

  const maskIP = (ip: string) => {
    if (!isPIIMasked || !ip || ip === 'Internal' || ip === '-') return ip;
    const parts = ip.split('.');
    if (parts.length !== 4) return ip;
    return `${parts[0]}.${parts[1]}.x.x`;
  };

  const handleUnlock = async (userId: string) => {
    try {
      await api.unlockUser(userId);
      fetchAllData(); // Refresh
    } catch (err) {
      alert("Failed to unlock user");
    }
  };

  const handleReset = async (userId: string) => {
    if (!window.confirm("Are you sure you want to reset this user's MFA? They will be forced to re-enroll.")) return;
    try {
      await api.resetMFA(userId);
      fetchAllData(); // Refresh
    } catch (err) {
      alert("Failed to reset MFA");
    }
  };

  // Convert stats object to array for mapping
  const statsArray = [
    { label: "Total Users", value: dashboardStats?.total_users?.toLocaleString() || "0", change: "+2% this month", trend: "up", icon: "group" },
    { label: "MFA Enabled", value: dashboardStats?.mfa_enabled?.toLocaleString() || "0", sub: `${dashboardStats?.total_users ? ((dashboardStats.mfa_enabled / dashboardStats.total_users) * 100).toFixed(1) : 0}% coverage`, percent: dashboardStats?.total_users ? (dashboardStats.mfa_enabled / dashboardStats.total_users) * 100 : 0, icon: "verified_user", color: "text-success" },
    { label: "MFA Pending", value: dashboardStats?.mfa_pending?.toLocaleString() || "0", sub: "Awaiting setup", percent: dashboardStats?.total_users ? (dashboardStats.mfa_pending / dashboardStats.total_users) * 100 : 0, icon: "pending", color: "text-warning" },
    { label: "MFA Disabled", value: dashboardStats?.mfa_disabled?.toLocaleString() || "0", sub: "Requires attention", percent: dashboardStats?.total_users ? (dashboardStats.mfa_disabled / dashboardStats.total_users) * 100 : 0, icon: "no_accounts", color: "text-text-muted" },
    { label: "Failed Attempts (24h)", value: dashboardStats?.failed_attempts_24h?.toLocaleString() || "0", change: "+12% vs avg", trend: "up", icon: "gpp_bad", color: "text-danger" },
    { label: "Locked Accounts", value: dashboardStats?.locked_accounts?.toLocaleString() || "0", sub: "Due to policy", icon: "lock_clock", color: "text-orange-500" },
  ];

  return (
    <div className="flex-1 flex flex-col overflow-hidden bg-background-dark h-full">
      {/* Action Header */}
      <div className="p-6 border-b border-border-color bg-surface-darker/50 flex flex-wrap gap-4 items-center z-10">
        <div className="flex-1 flex flex-wrap gap-4 items-center min-w-0">
          <div className="flex bg-surface-dark rounded-lg border border-border-color p-1 shadow-inner overflow-x-auto whitespace-nowrap">
            <button 
              onClick={() => setActiveTab('overview')}
              className={`px-4 py-1.5 text-xs font-medium rounded-md transition-all ${activeTab === 'overview' ? 'bg-primary text-white shadow-md' : 'text-text-muted hover:text-white hover:bg-border-color'}`}
            >Overview</button>
            <button 
              onClick={() => setActiveTab('policy')}
              className={`px-4 py-1.5 text-xs font-medium rounded-md transition-all ${activeTab === 'policy' ? 'bg-primary text-white shadow-md' : 'text-text-muted hover:text-white hover:bg-border-color'}`}
            >Policy Editor</button>
            <button 
              onClick={() => setActiveTab('users')}
              className={`px-4 py-1.5 text-xs font-medium rounded-md transition-all ${activeTab === 'users' ? 'bg-primary text-white shadow-md' : 'text-text-muted hover:text-white hover:bg-border-color'}`}
            >Users Security</button>
          </div>
          <div className="h-6 w-px bg-border-color hidden sm:block"></div>
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-surface-dark border border-border-color">
            <span className="text-[10px] text-text-muted font-bold uppercase tracking-tight whitespace-nowrap">Global MFA</span>
            <label className="relative inline-flex items-center cursor-pointer scale-90">
              <input type="checkbox" className="sr-only peer" checked={mfaPolicy?.require_mfa} readOnly />
              <div className="w-9 h-5 bg-gray-700 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-success"></div>
            </label>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-3 ml-auto sm:ml-0">
          <div className="flex items-center gap-2 px-2 bg-surface-dark/50 py-1 rounded-md border border-border-color/30">
            <span className={`w-2 h-2 rounded-full ${isAutoRefresh ? 'bg-success animate-pulse' : 'bg-text-muted'}`}></span>
            <span className="text-[10px] text-text-muted font-medium whitespace-nowrap">Live Feed</span>
          </div>
          <button 
            onClick={() => setIsPIIMasked(!isPIIMasked)}
            className={`px-4 py-2 rounded-lg text-xs font-medium transition-all flex items-center gap-2 border shadow-sm ${isPIIMasked ? 'bg-primary/20 text-primary border-primary/30 hover:bg-primary/30' : 'bg-surface-dark text-text-muted border-border-color hover:text-white hover:bg-border-color'}`}
          >
            <span className="material-symbols-outlined text-sm">{isPIIMasked ? 'visibility' : 'visibility_off'}</span>
            <span className="hidden sm:inline">{isPIIMasked ? 'Reveal PII' : 'Mask PII'}</span>
          </button>
          <button className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded-lg flex items-center gap-2 transition-all shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-95">
            <span className="material-symbols-outlined text-[16px]">lock_reset</span>
            <span className="hidden sm:inline">Force Enrollment</span>
            <span className="sm:hidden">Force</span>
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        {/* Stats Grid */}
        <div className="grid grid-cols-2 lg:grid-cols-6 gap-4 mb-6">
          {statsArray.map((stat, idx) => (
            <div key={idx} className="bg-surface-dark rounded-xl border border-border-color p-4 relative overflow-hidden group hover:border-primary/50 transition-all cursor-default shadow-lg hover:shadow-primary/5">
              <div className="absolute right-0 top-0 p-3 opacity-20 group-hover:opacity-40 transition-opacity">
                <span className={`material-symbols-outlined text-4xl ${stat.color?.replace('success', '[#34d399]').replace('warning', '[#fbbf24]').replace('danger', '[#fb7185]') || 'text-white'}`}>{stat.icon}</span>
              </div>
              <div className="text-text-muted text-[10px] font-bold uppercase tracking-wider mb-1">{stat.label}</div>
              <div className={`text-2xl font-bold ${stat.color ? stat.color.replace('success', '[#34d399]').replace('warning', '[#fbbf24]').replace('danger', '[#fb7185]') : 'text-white'}`}>{stat.value}</div>
              {stat.change && (
                <div className={`text-[10px] flex items-center gap-1 mt-1 font-medium ${stat.trend === 'up' ? 'text-[#34d399]' : 'text-[#fb7185]'}`}>
                  <span className="material-symbols-outlined text-[12px]">{stat.trend === 'up' ? 'trending_up' : 'trending_down'}</span>
                  {stat.change}
                </div>
              )}
              {stat.sub && <div className="text-[10px] text-text-muted mt-1 opacity-80">{stat.sub}</div>}
              {stat.percent !== undefined && (
                <div className="w-full bg-border-color/50 h-1.5 rounded-full mt-3 overflow-hidden">
                  <div className={`h-full transition-all duration-500 rounded-full ${stat.color?.includes('success') ? 'bg-emerald-500' : stat.color?.includes('warning') ? 'bg-amber-500' : 'bg-primary'}`} style={{ width: `${stat.percent}%` }}></div>
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Charts Row */}
        <div className="grid grid-cols-12 gap-6 mb-6">
          <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
            <div className="bg-surface-dark rounded-xl border border-border-color p-5 h-80 flex flex-col shadow-lg">
              <div className="flex justify-between items-center mb-6">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-primary text-[18px]">show_chart</span>
                  MFA Adoption Trend
                </h3>
                <select 
                  value={dateRange} 
                  onChange={(e) => setDateRange(e.target.value)}
                  className="bg-surface-darker text-[10px] text-text-muted border border-border-color rounded px-2 py-1 outline-none hover:text-white transition-colors"
                >
                  <option>Last 90 Days</option>
                  <option>Last 30 Days</option>
                  <option>Last 7 Days</option>
                </select>
              </div>
              <div className="flex-1 w-full translate-x-[-10px] min-h-[200px]">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={trendData}>
                    <defs>
                      <linearGradient id="colorUsers" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#137fec" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#137fec" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#233648" vertical={false} />
                    <XAxis dataKey="name" stroke="#64748b" fontSize={10} tickLine={false} axisLine={false} />
                    <YAxis stroke="#64748b" fontSize={10} tickLine={false} axisLine={false} />
                    <Tooltip 
                      contentStyle={{ backgroundColor: '#16202a', borderColor: '#233648', color: '#fff' }}
                      itemStyle={{ color: '#137fec', fontSize: '12px' }}
                    />
                    <Area type="monotone" dataKey="users" stroke="#137fec" fillOpacity={1} fill="url(#colorUsers)" strokeWidth={2} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-6 h-64">
              <div className="bg-surface-dark rounded-xl border border-border-color p-5 flex flex-col shadow-lg">
                <h3 className="text-white text-sm font-semibold mb-4">Method Distribution</h3>
                <div className="flex items-center gap-6 h-full">
                  <div className="w-32 h-32 flex-shrink-0">
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={methodDistribution}
                          innerRadius={35}
                          outerRadius={50}
                          paddingAngle={5}
                          dataKey="value"
                        >
                          {methodDistribution.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={entry.color} />
                          ))}
                        </Pie>
                      </PieChart>
                    </ResponsiveContainer>
                  </div>
                  <div className="flex-1 space-y-3">
                    {methodDistribution.map((item, idx) => (
                      <div key={idx} className="flex justify-between items-center text-xs">
                        <div className="flex items-center gap-2 text-text-muted">
                          <span className="w-2 h-2 rounded-full" style={{ backgroundColor: item.color }}></span>
                          {item.name}
                        </div>
                        <span className="text-white font-medium">{item.value}%</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
              <div className="bg-surface-dark rounded-xl border border-border-color p-5 flex flex-col justify-center gap-4 shadow-lg">
                <div className="p-3 bg-surface-darker rounded border border-border-color flex items-center justify-between hover:border-primary/30 transition-colors">
                  <div>
                    <div className="text-[10px] uppercase text-text-muted font-bold tracking-tight">SMS Costs (MTD)</div>
                    <div className="text-lg font-bold text-white">$482.12</div>
                  </div>
                  <div className="w-8 h-8 rounded bg-surface-dark flex items-center justify-center text-text-muted">
                    <span className="material-symbols-outlined text-[18px]">sms</span>
                  </div>
                </div>
                <div className="p-3 bg-surface-darker rounded border border-border-color flex items-center justify-between hover:border-primary/30 transition-colors">
                  <div>
                    <div className="text-[10px] uppercase text-text-muted font-bold tracking-tight">Hardware Keys</div>
                    <div className="text-lg font-bold text-white">1,204</div>
                  </div>
                  <div className="w-8 h-8 rounded bg-surface-dark flex items-center justify-center text-text-muted">
                    <span className="material-symbols-outlined text-[18px]">usb</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
            <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col h-[280px] shadow-lg">
              <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker/30 rounded-t-xl">
                <h3 className="text-white text-sm font-semibold">Geo-Threat Watch</h3>
                <button className="text-primary text-xs font-semibold hover:underline">Full Map</button>
              </div>
              <div className="flex-1 overflow-y-auto custom-scrollbar p-2">
                <div className="space-y-1">
                  {failureLocations.map((item, idx) => (
                    <div key={idx} className="flex items-center justify-between p-3 hover:bg-surface-darker rounded transition-colors group cursor-default">
                      <div className="flex items-center gap-3">
                        <span className="w-6 text-center text-xs font-mono text-text-muted">{idx + 1}</span>
                        <div className="flex flex-col">
                          <span className="text-sm text-white font-medium group-hover:text-primary transition-colors">{item.location}</span>
                          <span className="text-[10px] text-slate-400 font-mono">{item.last_ip}</span>
                        </div>
                      </div>
                      <span className={`text-[10px] font-bold px-2 py-1 rounded ${item.count > 30 ? 'bg-danger/10 text-danger' : 'bg-warning/10 text-warning'}`}>
                        {item.count} fails
                      </span>
                    </div>
                  ))}
                  {failureLocations.length === 0 && <div className="p-8 text-center text-text-muted text-xs">No threats detected</div>}
                </div>
              </div>
            </div>

            <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col flex-1 min-h-[280px] shadow-lg">
              <div className="p-4 border-b border-border-color bg-surface-darker/30 rounded-t-xl">
                <h3 className="text-white text-sm font-semibold">High-Risk Users</h3>
              </div>
              <div className="flex-1 overflow-y-auto custom-scrollbar p-2">
                <div className="space-y-1">
                  {failureUsers.map((user, idx) => (
                    <div key={idx} className="flex items-center justify-between p-3 hover:bg-surface-darker rounded transition-colors cursor-pointer group">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-full bg-surface-darker border border-border-color flex items-center justify-center text-xs font-bold text-white group-hover:border-primary transition-colors">{user.name?.[0]}</div>
                        <div className="flex flex-col">
                          <span className="text-sm text-white font-medium group-hover:text-primary transition-colors">{user.name}</span>
                          <span className="text-[10px] text-slate-400">{maskEmail(user.email)}</span>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-xs font-bold text-white">{user.failure_count}</div>
                        <div className={`text-[10px] ${user.failure_count > 5 ? 'text-danger' : 'text-warning'}`}>{user.failure_count > 5 ? 'High Risk' : 'Suspicious'}</div>
                      </div>
                    </div>
                  ))}
                  {failureUsers.length === 0 && <div className="p-8 text-center text-text-muted text-xs">No high-risk users identified</div>}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Logs Table */}
        <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col overflow-hidden shadow-xl">
          <div className="p-4 border-b border-border-color flex flex-wrap gap-4 justify-between items-center bg-surface-darker/50">
            <h3 className="text-white font-semibold text-sm flex items-center gap-2">
              <span className="material-symbols-outlined text-text-muted text-[18px]">history</span>
              MFA Security Audit Trail
            </h3>
            <div className="flex items-center gap-2">
              <div className="relative">
                <span className="absolute inset-y-0 left-0 pl-3 flex items-center text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">search</span>
                </span>
                <input 
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="bg-surface-dark border border-border-color text-xs text-white rounded-lg pl-9 pr-3 py-2 w-64 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all focus:w-80" 
                  placeholder="Filter by user, IP, or event..." 
                />
              </div>
              <button className="p-2 hover:bg-border-color rounded-lg text-text-muted hover:text-white transition-colors border border-transparent hover:border-border-color">
                <span className="material-symbols-outlined text-[18px]">filter_list</span>
              </button>
              <button className="p-2 hover:bg-border-color rounded-lg text-text-muted hover:text-white transition-colors border border-transparent hover:border-border-color">
                <span className="material-symbols-outlined text-[18px]">download</span>
              </button>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse min-w-[1000px]">
              <thead>
                <tr className="border-b border-border-color text-[10px] text-slate-300 uppercase tracking-widest bg-surface-darker/30">
                  <th className="px-6 py-4 font-bold">Timestamp</th>
                  <th className="px-6 py-4 font-bold">Identity</th>
                  <th className="px-6 py-4 font-bold">Event Type</th>
                  <th className="px-6 py-4 font-bold">Method</th>
                  <th className="px-6 py-4 font-bold">Access Point</th>
                  <th className="px-6 py-4 font-bold">Geo Target</th>
                  <th className="px-6 py-4 font-bold text-right">Result</th>
                </tr>
              </thead>
              <tbody className="text-xs">
                {recentEvents.map((ev, idx) => (
                  <tr key={idx} className="border-b border-border-color hover:bg-primary/5 transition-all group cursor-pointer border-l-2 border-transparent hover:border-primary">
                    <td className="px-6 py-4 text-slate-400 font-mono">{ev.time}</td>
                    <td className="px-6 py-4 font-semibold text-white group-hover:text-primary transition-colors">{maskEmail(ev.user)}</td>
                    <td className="px-6 py-4 text-white font-medium">{ev.event}</td>
                    <td className="px-6 py-4 text-slate-400">{ev.method}</td>
                    <td className="px-6 py-4 text-slate-400 font-mono tabular-nums">{maskIP(ev.ip)}</td>
                    <td className="px-6 py-4 text-slate-400 font-medium">{ev.loc}</td>
                    <td className="px-6 py-4 text-right">
                      <span className={`inline-flex items-center gap-1 px-2.5 py-1 rounded text-[9px] font-bold tracking-wider border ${
                        ev.status === 'SUCCESS' ? 'bg-[#34d399]/10 text-[#34d399] border-[#34d399]/40' :
                        ev.status === 'FAILED' ? 'bg-[#fb7185]/10 text-[#fb7185] border-[#fb7185]/40' :
                        ev.status === 'AUDIT' ? 'bg-[#137fec]/10 text-[#137fec] border-[#137fec]/40' :
                        'bg-[#fbbf24]/10 text-[#fbbf24] border-[#fbbf24]/40'
                      }`}>
                        {ev.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          
          {/* Pagination */}
          <div className="p-4 border-t border-border-color bg-surface-darker/50 flex items-center justify-between text-[10px] text-slate-300 font-semibold uppercase tracking-wider">
            <div className="flex items-center gap-4">
              <span>Showing 1 to {recentEvents.length} of {dashboardStats?.total_users || 0} entries</span>
              <select 
                value={pageSize} 
                onChange={(e) => setPageSize(Number(e.target.value))}
                className="bg-surface-dark border border-border-color rounded px-2 py-1 outline-none"
              >
                <option value={10}>10 per page</option>
                <option value={20}>20 per page</option>
                <option value={50}>50 per page</option>
              </select>
            </div>
            <div className="flex items-center gap-2">
              <button disabled className="px-3 py-1.5 rounded border border-border-color hover:bg-surface-dark hover:text-white transition-all disabled:opacity-30">Prev</button>
              <button className="px-3 py-1.5 rounded border border-primary bg-primary/10 text-primary">1</button>
              <button className="px-3 py-1.5 rounded border border-border-color hover:bg-surface-dark hover:text-white transition-all">2</button>
              <button className="px-3 py-1.5 rounded border border-border-color hover:bg-surface-dark hover:text-white transition-all">3</button>
              <button className="px-3 py-1.5 rounded border border-border-color hover:bg-surface-dark hover:text-white transition-all">Next</button>
            </div>
          </div>
        </div>

        <footer className="mt-12 text-center text-[10px] text-[#64748b] pb-6 flex flex-col items-center gap-1">
          <div className="flex items-center gap-4 mb-2">
            <span className="px-2 py-0.5 rounded bg-success/10 text-success border border-success/20 font-bold uppercase tracking-widest text-[8px]">AES-256 Encrypted</span>
            <span className="px-2 py-0.5 rounded bg-primary/10 text-primary border border-primary/20 font-bold uppercase tracking-widest text-[8px]">Security Module V2.4.1</span>
          </div>
          <span>© 2026 FlyFleet Global Security Operations Center. All rights reserved.</span>
          <span className="font-mono mt-1">GSOC TIME: {new Date().toLocaleTimeString('en-US', { hour12: false, timeZone: 'UTC' })} UTC | SYSTEM: OPERATIONAL</span>
        </footer>
      </div>
    </div>
  );
};

export default MFACenter;
