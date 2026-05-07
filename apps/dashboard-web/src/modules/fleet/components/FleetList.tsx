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
    
    let filtered = vehicles.filter(v => 
      v.registration_number?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      v.make?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (v.current_driver?.name || v.current_driver_name)?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return filtered.sort((a, b) => {
      const pMap: Record<string, number> = { 'CRITICAL': 0, 'WARNING': 1, 'HEALTHY': 2 };
      const priorityA = pMap[getPriority(a)];
      const priorityB = pMap[getPriority(b)];

      if (priorityA !== priorityB) return priorityA - priorityB;
      // Secondary: Alert count
      return (b.summary?.alert_count || 0) - (a.summary?.alert_count || 0);
    });
  }, [vehicles, searchQuery]);

  return (
    <div className="p-6 space-y-6 max-w-[1600px] mx-auto min-h-screen bg-[#0a0f14]">
      {/* Control Center Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="p-3.5 bg-primary/10 rounded-2xl border border-primary/20 shadow-lg shadow-primary/5">
             <Zap className="w-8 h-8 text-primary fill-primary" />
          </div>
          <div>
            <h1 className="text-3xl font-black text-white tracking-tighter italic uppercase leading-none">Fleet Operations Command</h1>
            <p className="text-gray-600 text-[10px] font-black uppercase tracking-[0.4em] mt-2 ml-0.5">Autonomous Asset Triage & Decision Engine</p>
          </div>
        </div>
        <div className="flex gap-4">
          <button 
            onClick={fetchVehicles}
            className="flex items-center gap-2 px-5 py-2.5 bg-white/5 border border-white/10 text-white rounded-xl hover:bg-white/10 transition-all font-bold text-[10px] uppercase tracking-widest"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh Signals
          </button>
          <button
            onClick={() => navigate('/admin/fleet/new')}
            className="flex items-center gap-2 px-7 py-2.5 bg-primary text-white rounded-xl hover:bg-primary/90 transition-all font-black text-[10px] uppercase tracking-[0.2em] shadow-2xl shadow-primary/40 border-b-2 border-primary-dark"
          >
            <Plus className="w-4 h-4" />
            Provision Asset
          </button>
        </div>
      </div>

      {/* Triage Search */}
      <div className="bg-surface-dark border border-border-color rounded-2xl p-4 shadow-2xl bg-white/[0.01]">
        <div className="relative max-w-2xl">
          <Search className="absolute left-5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-700" />
          <input
            type="text"
            placeholder="Identity registration, pilot, or serial..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-white/[0.03] border border-white/10 rounded-xl pl-14 pr-4 py-4 text-white focus:outline-none focus:border-primary transition-all text-sm font-semibold placeholder:text-gray-800 shadow-inner"
          />
        </div>
      </div>

      {/* Main Command Registry */}
      <div className="bg-[#10161d] border border-border-color rounded-2xl overflow-hidden shadow-2xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="text-[9px] font-black text-gray-700 uppercase tracking-[0.45em] bg-white/[0.01] border-b border-border-color">
                <th className="px-6 py-7 w-2"></th>
                <th className="px-6 py-7">Identity</th>
                <th className="px-6 py-7">Driver</th>
                <th className="px-6 py-7">Operational Health</th>
                <th className="px-6 py-7 min-w-[340px]">Compliance (Integrity)</th>
                <th className="px-6 py-7">Alerts</th>
                <th className="px-6 py-7 text-right pr-12">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
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
                      className="hover:bg-white/[0.02] transition-all cursor-pointer group relative"
                      onClick={() => navigate(`/admin/vehicles/${v.id}`)}
                    >
                      {/* Priority Signal */}
                      <td className="w-2 p-0">
                        <div className={`h-[140px] w-2 transition-all ${
                          priority === 'CRITICAL' ? 'bg-rose-600 shadow-[8px_0_30px_rgba(225,29,72,0.65)]' : 
                          priority === 'WARNING' ? 'bg-amber-500' : 'bg-emerald-500'
                        }`} />
                      </td>

                      {/* 1. Identity */}
                      <td className="px-6 py-6">
                        <div className="flex items-center gap-6">
                          <div className={`w-18 h-18 rounded-2xl flex flex-col items-center justify-center font-black text-[10px] border-2 transition-transform group-hover:scale-105 ${
                            priority === 'CRITICAL' ? 'bg-rose-500/10 text-rose-500 border-rose-500/25' : 
                            priority === 'WARNING' ? 'bg-amber-500/10 text-amber-500 border-amber-500/25' : 
                            'bg-primary/10 text-primary border-primary/25'
                          }`}>
                            <Truck className="w-8 h-8 mb-1.5 opacity-80" />
                            <span className="text-[9px] uppercase opacity-60 tracking-widest leading-none">{v.vehicle_type || 'VEH'}</span>
                          </div>
                          <div className="space-y-1.5">
                            <div className="flex items-center gap-3">
                               <div className="font-black text-white text-2xl leading-none tracking-tighter group-hover:text-primary transition-colors italic">{v.registration_number}</div>
                               {priority === 'CRITICAL' && <Lock className="w-4.5 h-4.5 text-rose-500 animate-pulse" />}
                            </div>
                            <div className="text-[11px] text-gray-500 uppercase font-black tracking-widest">{v.make} {v.model}</div>
                            <div className="flex items-center gap-2.5 mt-2.5">
                               <Fuel className="w-3.5 h-3.5 text-gray-800" />
                               <span className="text-[10px] font-black text-gray-700 uppercase tracking-widest">{v.fuel_type || 'Diesel'}</span>
                            </div>
                          </div>
                        </div>
                      </td>

                      {/* 2. Driver */}
                      <td className="px-6 py-6">
                        <div className="space-y-1">
                          <div className={`font-black text-base tracking-tight transition-colors uppercase italic ${driverName === 'Unassigned' ? 'text-gray-700 font-bold' : 'text-gray-200 group-hover:text-white'}`}>
                            {driverName}
                          </div>
                          <div className="text-[9px] font-black text-gray-700 uppercase tracking-widest">Active Personnel</div>
                        </div>
                      </td>

                      {/* 3. Operational Health */}
                      <td className="px-6 py-6">
                        <div className="space-y-6">
                          <div className="flex flex-col">
                             <div className="flex items-center gap-3">
                                <Signal className={`w-4.5 h-4.5 ${tel.status === 'GREEN' ? 'text-emerald-500' : tel.status === 'YELLOW' ? 'text-amber-500' : 'text-rose-600'}`} />
                                <span className={`text-[12px] font-black uppercase tracking-widest ${tel.status === 'GREEN' ? 'text-emerald-500' : tel.status === 'YELLOW' ? 'text-amber-500' : 'text-rose-600'}`}>
                                  {tel.text}
                                </span>
                             </div>
                             {tel.sub && <span className="text-[10px] text-gray-700 font-black ml-7.5 uppercase tracking-[0.1em] mt-1.5">{tel.sub}</span>}
                          </div>
                          <div className="flex items-center gap-3">
                             <Toolbox className={`w-4 h-4 opacity-40 ${summary.maintenance_status === 'HEALTHY' ? 'text-emerald-500' : 'text-amber-500'}`} />
                             <div className={`text-[10px] font-black uppercase tracking-widest px-2 py-0.5 rounded border ${
                               summary.maintenance_status === 'HEALTHY' ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-500/60' : 'bg-amber-500/10 border-amber-500/20 text-amber-500/60'
                             }`}>
                               Maintenance {summary.maintenance_status === 'HEALTHY' ? 'Healthy' : 'Due'}
                             </div>
                          </div>
                        </div>
                      </td>

                      {/* 4. Compliance (Integrity) - PRIMARY ANCHOR */}
                      <td className="px-6 py-6">
                        <div className="space-y-4 group/compliance relative">
                          <div className="flex flex-col">
                             <span className={`text-4xl font-black italic tracking-tighter leading-none mb-4 drop-shadow-xl ${
                               comp.status === 'GREEN' ? 'text-emerald-500' : 
                               comp.status === 'YELLOW' ? 'text-amber-500' : 
                               'text-rose-600 animate-pulse'
                             }`}>
                               {comp.text}
                             </span>
                             <div className="flex items-center gap-4">
                               <div className={`px-3 py-1.5 rounded text-[10px] font-black border transition-colors ${
                                 comp.status === 'RED' ? 'bg-rose-500/10 text-rose-500 border-rose-500/25' : 
                                 comp.status === 'YELLOW' ? 'bg-amber-500/10 text-amber-500 border-amber-500/25' : 
                                 'bg-emerald-500/10 text-emerald-500 border-emerald-500/25'
                               }`}>
                                 {comp.status === 'RED' ? 'HIGH RISK' : (summary.risk_score > 30 ? 'MED RISK' : 'LOW RISK')}
                               </div>
                               <div className="flex items-center gap-2.5 text-gray-500">
                                  <Calendar className={`w-5 h-5 ${comp.status === 'RED' ? 'text-rose-500' : 'text-gray-800'}`} />
                                  <span className={`text-[12px] font-black uppercase tracking-widest leading-none ${comp.status === 'RED' ? 'text-rose-500' : 'text-gray-600'}`}>
                                    {comp.detail}
                                  </span>
                               </div>
                             </div>
                          </div>

                          {/* Hover Tooltip - Registry Vault */}
                          <div className="absolute left-0 top-[120%] w-80 bg-[#0b1219] border-2 border-border-color p-6 rounded-3xl shadow-[0_40px_80px_rgba(0,0,0,0.9)] opacity-0 group-hover/compliance:opacity-100 transition-all z-50 pointer-events-none translate-y-4 group-hover/compliance:translate-y-0 backdrop-blur-3xl border-white/5">
                             <div className="flex items-center justify-between mb-6 pb-4 border-b border-white/10">
                                <div className="flex items-center gap-3">
                                  <FileText className="w-6 h-6 text-primary" />
                                  <span className="text-[12px] font-black text-white uppercase tracking-[0.35em]">Digital Vault Integrity</span>
                                </div>
                                <Activity className="w-5 h-5 text-gray-700" />
                             </div>
                             <div className="space-y-5">
                                {[
                                  { label: 'Insurance Policy', status: comp.status === 'GREEN' ? 'VALID' : 'REVIEW', color: comp.status === 'GREEN' ? 'text-emerald-500' : 'text-rose-500' },
                                  { label: 'Carrier Permit', status: 'VERIFIED', color: 'text-emerald-500' },
                                  { label: 'Pollution (PUC)', status: 'EXPIRING', color: 'text-amber-500' },
                                  { label: 'Fitness Registry', status: 'SECURED', color: 'text-emerald-500' }
                                ].map((doc, idx) => (
                                  <div key={idx} className="flex justify-between items-center group/item hover:bg-white/[0.04] p-2 rounded-xl transition-colors">
                                    <span className="text-[11px] font-bold text-gray-500">{doc.label}</span>
                                    <span className={`text-[11px] font-black ${doc.color} tracking-[0.1em]`}>{doc.status}</span>
                                  </div>
                                ))}
                             </div>
                             <div className="mt-6 pt-4 border-t border-white/5 flex items-center justify-between">
                                <span className="text-[9px] font-black text-gray-700 uppercase">Registry Updated 2h ago</span>
                                <Info className="w-3.5 h-3.5 text-gray-800" />
                             </div>
                          </div>
                        </div>
                      </td>

                      {/* 5. Alerts - CONDITIONAL */}
                      <td className="px-6 py-6">
                         {(summary.alert_count || 0) > 0 ? (
                           <div className="flex flex-col gap-4">
                              {summary.critical_alert_count > 0 && (
                                <div className="flex items-center gap-3 px-4 py-2 rounded-2xl bg-rose-500/10 border border-rose-500/25 shadow-[0_0_10px_rgba(225,29,72,0.1)]">
                                   <ShieldAlert className="w-5 h-5 text-rose-500" />
                                   <span className="text-xs font-black text-rose-500">{summary.critical_alert_count} CRITICAL</span>
                                </div>
                              )}
                              {summary.alert_count - summary.critical_alert_count > 0 && (
                                <div className="flex items-center gap-3 px-4 py-2 rounded-2xl bg-amber-500/10 border border-amber-500/25">
                                   <AlertTriangle className="w-5 h-5 text-amber-500" />
                                   <span className="text-xs font-black text-amber-500">{summary.alert_count - summary.critical_alert_count} WARNING</span>
                                </div>
                              )}
                           </div>
                         ) : (
                           <div className="text-[10px] font-black text-gray-800 uppercase tracking-widest opacity-30 italic">No Operational Alerts</div>
                         )}
                      </td>

                      {/* 6. Actions */}
                      <td className="px-6 py-6 text-right">
                        <div className="flex items-center justify-end gap-3 opacity-0 group-hover:opacity-100 transition-all translate-x-8 group-hover:translate-x-0">
                          <button 
                            title="Assign Pilot"
                            onClick={(e) => { e.stopPropagation(); }}
                            className="p-3.5 bg-white/5 rounded-2xl text-gray-600 hover:text-primary hover:bg-primary/10 transition-all border border-transparent hover:border-primary/25 shadow-lg"
                          >
                            <UserPlus className="w-5.5 h-5.5" />
                          </button>
                          <button 
                            title="Registry Vault"
                            onClick={(e) => { e.stopPropagation(); }}
                            className="p-3.5 bg-white/5 rounded-2xl text-gray-600 hover:text-white hover:bg-white/10 transition-all border border-transparent hover:border-white/15"
                          >
                            <FileText className="w-5.5 h-5.5" />
                          </button>
                          <button 
                            title={comp.text === 'BLOCKED' ? 'Unblock Asset' : 'Emergency Block'}
                            onClick={(e) => { e.stopPropagation(); }}
                            className={`p-3.5 bg-white/5 rounded-2xl text-gray-600 transition-all border border-transparent ${
                               comp.text === 'BLOCKED' ? 'hover:text-emerald-500 hover:bg-emerald-500/10 hover:border-emerald-500/25' : 'hover:text-rose-500 hover:bg-rose-500/10 hover:border-rose-500/25'
                            }`}
                          >
                            {comp.text === 'BLOCKED' ? <Unlock className="w-5.5 h-5.5" /> : <Ban className="w-5.5 h-5.5" />}
                          </button>
                          <div className="w-px h-10 bg-white/10 mx-3" />
                          <ChevronRight className="w-7 h-7 text-gray-800 group-hover:text-primary transition-all translate-x-0 group-hover:translate-x-1" />
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
  );
};
