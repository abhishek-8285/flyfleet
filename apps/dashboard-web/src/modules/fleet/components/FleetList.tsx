import React, { useEffect, useState, useCallback, useMemo } from 'react';
import {
  Search,
  Plus,
  Filter,
  MoreVertical,
  ChevronRight,
  Truck,
  Signal,
  RefreshCw,
  AlertTriangle,
  Clock,
  ShieldAlert,
  Calendar,
  Toolbox,
  Lock,
  FileText,
  UserCheck,
  Zap,
  Activity,
  AlertCircle,
  Fuel,
  Settings,
  UserPlus,
  Ban,
  Unlock,
  Info
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { api } from '../../../services/api';

type SortKey = 'priority' | 'expiry' | 'alerts' | 'last_seen';

export const FleetList: React.FC = () => {
  const navigate = useNavigate();
  const [vehicles, setVehicles] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [sortKey, setSortKey] = useState<SortKey>('priority');
  const [filter, setFilter] = useState<'ALL' | 'CRITICAL' | 'ACTIVE'>('ALL');

  const fetchVehicles = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getVehicles({ limit: 100 });
      const data = res?.data || res;
      if (Array.isArray(data)) {
        setVehicles(data);
      } else if (data && typeof data === 'object' && Array.isArray((data as any).vehicles)) {
        setVehicles((data as any).vehicles);
      } else {
        setVehicles([]);
      }
    } catch (error) {
      console.error('Failed to fetch vehicles:', error);
      setVehicles([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchVehicles();
  }, [fetchVehicles]);

  const getDriverLicenseData = (v: any) => {
    const summary = v.summary || {};
    const status = summary.driver_license_status;
    if (!v.current_driver_id) return { status: 'GRAY', text: 'Unassigned', isCritical: false };
    if (!status || status === 'UNKNOWN' || status === 'EXPIRED') return { status: 'RED', text: 'Invalid License', isCritical: true };
    return { status: 'GREEN', text: 'Valid', isCritical: false };
  };

  const getComplianceData = (v: any) => {
    const summary = v.summary || {};
    const expiryDate = summary.next_expiry_date;
    const lic = getDriverLicenseData(v);
    
    // Rule: Driver license invalid -> Compliance = BLOCKED
    if (lic.isCritical) {
      return { 
        status: 'RED', 
        text: 'BLOCKED', 
        detail: 'Driver License Expired/Invalid',
        risk: 'HIGH RISK'
      };
    }

    // Rule: No documents -> Compliance = BLOCKED
    if (!expiryDate) {
      return { 
        status: 'RED', 
        text: 'BLOCKED', 
        detail: 'No documents uploaded',
        risk: 'HIGH RISK'
      };
    }
    
    const now = new Date();
    const expiry = new Date(expiryDate);
    const diffDays = Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    
    // Rule: Expired doc -> Compliance = BLOCKED
    if (diffDays < 0) {
      return { 
        status: 'RED', 
        text: 'BLOCKED', 
        detail: `${summary.next_expiry_type || 'Insurance'} expired ${Math.abs(diffDays)}d ago`,
        risk: 'HIGH RISK'
      };
    }
    if (diffDays <= 30) {
      return { 
        status: 'YELLOW', 
        text: 'EXPIRING SOON', 
        detail: `${summary.next_expiry_type || 'Insurance'} in ${diffDays}d`,
        risk: 'MED RISK'
      };
    }
    return { 
      status: 'GREEN', 
      text: 'VALID', 
      detail: `${summary.next_expiry_type || 'Insurance'} valid (${diffDays}d)`,
      risk: (summary.risk_score || 0) > 30 ? 'MED RISK' : 'LOW RISK'
    };
  };

  const getTelemetryData = (v: any) => {
    const summary = v.summary || {};
    const lastSeen = summary.last_seen;
    const deviceId = v.gps_device_id;

    if (!deviceId) return { status: 'RED', text: 'No device', sub: '' };
    if (!lastSeen) return { status: 'RED', text: 'Offline', sub: 'Never seen' };
    
    const now = new Date();
    const then = new Date(lastSeen);
    const diffMins = Math.floor((now.getTime() - then.getTime()) / 1000 / 60);
    
    if (diffMins < 5) return { status: 'GREEN', text: 'Live', sub: 'Active' };
    if (diffMins < 60) return { status: 'YELLOW', text: 'Stale', sub: `${diffMins}m ago` };
    return { status: 'RED', text: 'Offline', sub: `${Math.floor(diffMins/60)}h ago` };
  };

  const getPriority = (v: any) => {
    const comp = getComplianceData(v);
    if (comp.text === 'BLOCKED') return 'CRITICAL';
    if (comp.status === 'YELLOW') return 'WARNING';
    return 'HEALTHY';
  };

  const sortedAndFilteredVehicles = useMemo(() => {
    if (!Array.isArray(vehicles)) return [];
    
    let filtered = vehicles.filter(v => {
      const matchesSearch = 
        v.registration_number?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        v.make?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (v.current_driver?.name || v.current_driver_name)?.toLowerCase().includes(searchQuery.toLowerCase());
      
      if (!matchesSearch) return false;
      if (filter === 'CRITICAL') return getPriority(v) === 'CRITICAL';
      if (filter === 'ACTIVE') return v.status === 'ACTIVE';
      return true;
    });

    return filtered.sort((a, b) => {
      const pMap: Record<string, number> = { 'CRITICAL': 0, 'WARNING': 1, 'HEALTHY': 2 };
      const priorityA = pMap[getPriority(a)];
      const priorityB = pMap[getPriority(b)];

      if (priorityA !== priorityB) return priorityA - priorityB;
      // Secondary: Alert count
      return (b.summary?.alert_count || 0) - (a.summary?.alert_count || 0);
    });
  }, [vehicles, searchQuery, filter]);

  const stats = useMemo(() => {
    const total = vehicles.length;
    const active = vehicles.filter(v => v.status === 'ACTIVE').length;
    const critical = vehicles.filter(v => getPriority(v) === 'CRITICAL').length;
    const offline = vehicles.filter(v => getTelemetryData(v).status === 'RED').length;
    return { total, active, critical, offline };
  }, [vehicles]);

  return (
    <div className="flex-1 flex flex-col bg-background-dark overflow-hidden font-sans relative">
      {/* --- Standardized Header --- */}
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-30">
        <div className="flex items-center gap-4">
          <div className="p-2 bg-primary/10 rounded-lg">
             <Zap className="w-5 h-5 text-primary fill-primary" />
          </div>
          <div>
            <div className="text-[10px] text-text-muted uppercase tracking-[0.2em] font-black mb-0.5">
              Operations Center
            </div>
            <h1 className="text-white text-lg font-bold uppercase tracking-tight leading-none">
              Fleet Operations
            </h1>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button 
            onClick={fetchVehicles}
            className="px-4 py-2 bg-surface-dark border border-border-color text-text-muted hover:text-white text-xs font-semibold rounded-lg transition-all flex items-center gap-2"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            Sync
          </button>
          <button
            onClick={() => navigate('/admin/vehicles/new')}
            className="px-6 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded-lg flex items-center gap-2 shadow-lg shadow-primary/20 transition-all"
          >
            <Plus className="w-3.5 h-3.5" />
            Add Vehicle
          </button>
        </div>
      </header>

      {/* --- Main Area --- */}
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="max-w-[1600px] mx-auto space-y-6 pb-32">
          
          {/* Status Summary Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
              <div className="p-3 rounded-xl bg-primary/10 text-primary">
                <Truck className="w-6 h-6" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Total Assets</div>
                <div className="text-xl font-bold text-white">{stats.total}</div>
              </div>
            </div>

            <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
              <div className="p-3 rounded-xl bg-emerald-500/10 text-emerald-500">
                <Zap className="w-6 h-6" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Active Now</div>
                <div className="text-xl font-bold text-emerald-500">{stats.active}</div>
              </div>
            </div>

            <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
              <div className="p-3 rounded-xl bg-rose-500/10 text-rose-500">
                <ShieldAlert className="w-6 h-6" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Critical Priority</div>
                <div className="text-xl font-bold text-rose-500">{stats.critical}</div>
              </div>
            </div>

            <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
              <div className="p-3 rounded-xl bg-amber-500/10 text-amber-500">
                <Signal className="w-6 h-6" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Signal Loss</div>
                <div className="text-xl font-bold text-amber-500">{stats.offline}</div>
              </div>
            </div>
          </div>

          {/* Triage Search Bar */}
          <div className="bg-surface-dark border border-border-color rounded-xl p-3 shadow-xl flex items-center gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-700" />
              <input
                type="text"
                placeholder="Search registration, driver, or model..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full bg-black/20 border border-white/5 rounded-lg pl-12 pr-4 py-2.5 text-white focus:outline-none focus:border-primary transition-all text-sm font-semibold placeholder:text-gray-800"
              />
            </div>
            <div className="flex items-center gap-2 px-4 border-l border-white/5">
               <span className="text-[10px] font-black text-gray-600 uppercase tracking-widest">Filter:</span>
               <div className="flex gap-1">
                  {['ALL', 'CRITICAL', 'ACTIVE'].map(f => (
                    <button 
                      key={f} 
                      onClick={() => setFilter(f as any)}
                      className={`px-3 py-1 rounded text-[9px] font-black uppercase border transition-all ${
                        filter === f 
                          ? 'bg-primary border-primary text-white' 
                          : 'border-white/5 hover:bg-white/5 text-gray-500 hover:text-white'
                      }`}
                    >
                      {f}
                    </button>
                  ))}
               </div>
            </div>
          </div>

      {/* Vehicle Fleet Overview */}
          <div className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-2xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="text-[9px] font-black text-gray-700 uppercase tracking-[0.45em] bg-white/[0.01] border-b border-border-color">
                <th className="px-6 py-7 w-2"></th>
                <th className="px-6 py-7">Vehicle Information</th>
                <th className="px-6 py-7">Driver</th>
                <th className="px-6 py-7">Status</th>
                <th className="px-6 py-7 min-w-[340px]">Documents & Compliance</th>
                <th className="px-6 py-7">Alerts</th>
                <th className="px-6 py-7 text-right pr-12">Actions</th>
              </tr>
            </thead>
            <tbody className="">
              {loading ? (
                Array(5).fill(0).map((_, i) => (
                  <tr key={i} className="animate-pulse">
                    <td className="px-6 py-16" colSpan={7}><div className="h-14 bg-white/[0.01] rounded-2xl w-full" /></td>
                  </tr>
                ))
              ) : (
                sortedAndFilteredVehicles.map((v) => {
                  const summary = v.summary || {};
                  const priority = getPriority(v);
                  const tel = getTelemetryData(v);
                  const comp = getComplianceData(v);
                  const driverName = v.current_driver?.name || v.current_driver_name || 'Unassigned';
                  
                  return (
                    <tr 
                      key={v.id} 
                      className="hover:bg-white/[0.02] transition-all cursor-pointer group relative border-b border-white/[0.05]"
                      onClick={() => navigate(`/admin/vehicles/${v.id}`)}
                    >
                      {/* Priority Signal */}
                      <td className="w-1 p-0">
                        <div className={`h-[64px] w-1 transition-all ${
                          priority === 'CRITICAL' ? 'bg-rose-600 shadow-[2px_0_10px_rgba(225,29,72,0.4)]' : 
                          priority === 'WARNING' ? 'bg-amber-500' : 'bg-emerald-500'
                        }`} />
                      </td>

                      {/* 1. Identity */}
                      <td className="px-5 py-2.5">
                        <div className="flex items-center gap-3">
                          <div className={`w-9 h-9 rounded-lg flex flex-col items-center justify-center font-black text-[8px] border transition-transform group-hover:scale-105 ${
                            priority === 'CRITICAL' ? 'bg-rose-500/10 text-rose-500 border-rose-500/25' : 
                            priority === 'WARNING' ? 'bg-amber-500/10 text-amber-500 border-amber-500/25' : 
                            'bg-primary/10 text-primary border-primary/25'
                          }`}>
                            <Truck className="w-4 h-4 mb-0 opacity-80" />
                          </div>
                          <div className="space-y-0">
                            <div className="flex items-center gap-1.5">
                               <div className="font-black text-white text-base leading-none tracking-tight group-hover:text-primary transition-colors">{v.registration_number}</div>
                               {priority === 'CRITICAL' && <Lock className="w-3 h-3 text-rose-500" />}
                            </div>
                            <div className="text-[9px] text-gray-500 uppercase font-black tracking-widest">{v.make} {v.model}</div>
                          </div>
                        </div>
                      </td>

                      {/* 2. Driver */}
                      <td className="px-5 py-2.5">
                        <div className="space-y-0">
                          <div className={`font-black text-xs tracking-tight transition-colors uppercase ${driverName === 'Unassigned' ? 'text-gray-700' : 'text-gray-200 group-hover:text-white'}`}>
                            {driverName}
                          </div>
                          <div className="text-[8px] font-black text-gray-700 uppercase tracking-widest">Driver</div>
                        </div>
                      </td>

                      {/* 3. Operational Health */}
                      <td className="px-5 py-2.5">
                        <div className="flex items-center gap-6">
                          <div className="flex flex-col">
                             <div className="flex items-center gap-2">
                                <Signal className={`w-3.5 h-3.5 ${tel.status === 'GREEN' ? 'text-emerald-500' : tel.status === 'YELLOW' ? 'text-amber-500' : 'text-rose-600'}`} />
                                <span className={`text-[10px] font-black uppercase tracking-widest ${tel.status === 'GREEN' ? 'text-emerald-500' : tel.status === 'YELLOW' ? 'text-amber-500' : 'text-rose-600'}`}>
                                  {tel.text}
                                </span>
                             </div>
                             {tel.sub && <span className="text-[8px] text-gray-700 font-black ml-5.5 uppercase tracking-tighter">{tel.sub}</span>}
                          </div>
                          <div className="flex items-center gap-2">
                             <Toolbox className={`w-3 h-3 opacity-30 ${summary.maintenance_status === 'HEALTHY' ? 'text-emerald-500' : 'text-amber-500'}`} />
                             <div className={`text-[8px] font-black uppercase tracking-widest px-1 py-0.5 rounded border ${
                                summary.maintenance_status === 'HEALTHY' ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-500/60' : 'bg-amber-500/10 border-amber-500/20 text-amber-500/60'
                             }`}>
                                {summary.maintenance_status === 'HEALTHY' ? 'OK' : 'DUE'}
                             </div>
                          </div>
                        </div>
                      </td>

                      {/* 4. Compliance (Integrity) */}
                      <td className="px-5 py-2.5">
                        <div className="flex items-center gap-6 group/compliance relative">
                          <div className="flex flex-col">
                             <span className={`text-base font-black tracking-tighter leading-none mb-1 ${
                                comp.status === 'GREEN' ? 'text-emerald-500' : 
                                comp.status === 'YELLOW' ? 'text-amber-500' : 
                                'text-rose-600 animate-pulse'
                             }`}>
                               {comp.text}
                             </span>
                             <div className="flex items-center gap-3">
                               <div className={`px-1.5 py-0.5 rounded-[3px] text-[8px] font-black border ${
                                 comp.status === 'RED' ? 'bg-rose-500/10 text-rose-500 border-rose-500/25' : 
                                 comp.status === 'YELLOW' ? 'bg-amber-500/10 text-amber-500 border-amber-500/25' : 
                                 'bg-emerald-500/10 text-emerald-500 border-emerald-500/25'
                               }`}>
                                 {comp.status === 'RED' ? 'RISK' : 'SAFE'}
                               </div>
                               <div className="flex items-center gap-1.5 text-gray-600">
                                  <Calendar className="w-3.5 h-3.5" />
                                  <span className="text-[9px] font-black uppercase tracking-tight leading-none">
                                    {comp.detail}
                                  </span>
                               </div>
                             </div>
                          </div>

                          {/* Document Preview Tooltip remains useful but smaller */}
                          <div className="absolute left-0 top-[100%] w-64 bg-[#0b1219] border border-border-color p-4 rounded-xl shadow-[0_20px_40px_rgba(0,0,0,0.8)] opacity-0 group-hover/compliance:opacity-100 transition-all z-50 pointer-events-none translate-y-1 group-hover/compliance:translate-y-0 backdrop-blur-xl">
                             <div className="space-y-2">
                                <div className="text-[8px] font-black text-gray-500 uppercase tracking-widest mb-2">Required Documents</div>
                                {[
                                  { label: 'Insurance', status: comp.status === 'GREEN' ? 'VALID' : 'REVIEW', color: comp.status === 'GREEN' ? 'text-emerald-500' : 'text-rose-500' },
                                  { label: 'Permit', status: 'VERIFIED', color: 'text-emerald-500' },
                                  { label: 'Fitness', status: 'SECURED', color: 'text-emerald-500' }
                                ].map((doc, idx) => (
                                  <div key={idx} className="flex justify-between items-center text-[9px] font-bold">
                                    <span className="text-gray-500">{doc.label}</span>
                                    <span className={doc.color}>{doc.status}</span>
                                  </div>
                                ))}
                             </div>
                          </div>
                        </div>
                      </td>

                      {/* 5. Alerts */}
                      <td className="px-5 py-2.5">
                         {(summary.alert_count || 0) > 0 ? (
                           <div className="flex gap-2">
                              {summary.critical_alert_count > 0 && (
                                <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-lg bg-rose-500/10 border border-rose-500/25">
                                   <ShieldAlert className="w-3 h-3 text-rose-500" />
                                   <span className="text-[8px] font-black text-rose-500">{summary.critical_alert_count}!</span>
                                </div>
                              )}
                              {summary.alert_count - summary.critical_alert_count > 0 && (
                                <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-lg bg-amber-500/10 border border-amber-500/25">
                                   <AlertTriangle className="w-3 h-3 text-amber-500" />
                                   <span className="text-[8px] font-black text-amber-500">{summary.alert_count - summary.critical_alert_count}</span>
                                </div>
                              )}
                           </div>
                         ) : (
                           <div className="text-[8px] font-black text-gray-800 uppercase tracking-widest opacity-20">Clean</div>
                         )}
                      </td>

                      {/* 6. Actions */}
                      <td className="px-5 py-2.5 text-right">
                        <div className="flex items-center justify-end gap-1.5 opacity-0 group-hover:opacity-100 transition-all">
                          <button 
                            className="p-1.5 bg-white/5 rounded-lg text-gray-600 hover:text-primary hover:bg-primary/10 transition-all"
                          >
                            <UserPlus className="w-3.5 h-3.5" />
                          </button>
                          <button 
                            className="p-1.5 bg-white/5 rounded-lg text-gray-600 hover:text-white hover:bg-white/10 transition-all"
                          >
                            <FileText className="w-3.5 h-3.5" />
                          </button>
                          <ChevronRight className="w-4 h-4 text-gray-800 group-hover:text-primary transition-all" />
                        </div>
                      </td>
                    </tr>
                  );
                })
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
