import {
    ArrowPathIcon,
    CheckCircleIcon,
    EyeIcon,
    GlobeAltIcon,
    InformationCircleIcon,
    NoSymbolIcon,
    ShieldExclamationIcon
} from '@heroicons/react/24/outline';
import { format } from 'date-fns';
import React, { useEffect, useState } from 'react';
import SecurityFooter from '../../../components/SecurityFooter';
import SecuritySidebar from '../../../components/SecuritySidebar';
import api from '../../../services/api';

const ThreatAlerts: React.FC = () => {
  const [alerts, setAlerts] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatingId, setUpdatingId] = useState<string | null>(null);

  useEffect(() => {
    fetchAlerts();
  }, []);

  const fetchAlerts = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.getThreatAlerts({ size: 50 });
      const data = (response as any)?.data ?? response;
      const items = data?.items || data?.alerts || data?.data || data || [];
      setAlerts(Array.isArray(items) ? items : []);
    } catch (error) {
       setError('Failed to load threat alerts.');
       setAlerts([]);
    } finally {
      setLoading(false);
    }
  };

  const getSeverityStyles = (severity: string) => {
    switch (severity) {
      case 'critical': return 'bg-red-500/10 text-vibrant-red border-red-500/30';
      case 'warning': return 'bg-amber-500/10 text-amber-400 border-amber-500/30';
      default: return 'bg-blue-500/10 text-vibrant-blue border-blue-500/30';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'BRUTE_FORCE': return <ShieldExclamationIcon className="w-6 h-6" />;
      case 'GEO_ANOMALY': return <GlobeAltIcon className="w-6 h-6" />;
      default: return <InformationCircleIcon className="w-6 h-6" />;
    }
  };

  const handleUpdateStatus = async (alertId: string, status: 'acknowledged' | 'blocked') => {
    try {
      setUpdatingId(alertId);
      await api.updateThreatAlertStatus(alertId, status);
      setAlerts((prev) =>
        prev.map((a) => (a.id === alertId ? { ...a, status } : a))
      );
    } catch (err) {
      alert('Failed to update alert status');
    } finally {
      setUpdatingId(null);
    }
  };

  return (
    <div className="flex h-screen bg-[#0b1117] text-slate-200 overflow-hidden font-inter">
      <SecuritySidebar />
      
      <main className="flex-1 flex flex-col min-w-0 bg-grid-pattern relative">
        <div className="glow-effect" />
        
        {/* Header */}
        <header className="px-8 py-6 border-b border-white/5 bg-slate-900/40 backdrop-blur-md z-10 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-white flex items-center gap-3">
              <span className="material-symbols-outlined text-vibrant-red text-3xl">warning</span>
              Active Threat Engine
            </h2>
            <p className="text-slate-400 text-sm mt-1">Heuristic analysis of authentication patterns and network anomalies.</p>
          </div>
          
          <div className="flex items-center gap-4">
             <div className="flex -space-x-2">
                {[1, 2, 3].map(i => (
                  <div key={i} className="w-8 h-8 rounded-full border-2 border-[#0b1117] bg-slate-800 flex items-center justify-center text-[10px] font-bold text-slate-400 overflow-hidden">
                    <img src={`https://i.pravatar.cc/150?u=${i}`} alt="user" />
                  </div>
                ))}
                <div className="w-8 h-8 rounded-full border-2 border-[#0b1117] bg-slate-700 flex items-center justify-center text-[10px] font-bold text-white">
                  +2
                </div>
             </div>
             <span className="text-xs text-slate-500 font-medium">Analysts Online</span>
          </div>
        </header>

        {/* Content */}
        <div className="flex-1 overflow-auto custom-scrollbar p-8 z-10 flex flex-col gap-8">
           
           {/* Summary Cards */}
           <div className="grid grid-cols-3 gap-6">
              {[
                { title: 'Critical Threats', value: alerts.filter(a => a.severity === 'critical').length.toString(), color: 'text-red-500', trend: 'Live signal' },
                { title: 'Open Alerts', value: alerts.filter(a => (a.status || 'open') === 'open').length.toString(), color: 'text-blue-400', trend: 'Awaiting action' },
                { title: 'Blocked Today', value: alerts.filter(a => a.status === 'blocked').length.toString(), color: 'text-emerald-400', trend: 'Auto shield active' },
              ].map((c, i) => (
                <div key={i} className="glass-card p-6 rounded-2xl relative overflow-hidden group">
                   <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
                      <ShieldExclamationIcon className="w-20 h-20" />
                   </div>
                   <h3 className="text-slate-500 text-sm font-bold uppercase tracking-widest">{c.title}</h3>
                   <p className={`text-4xl font-black mt-2 ${c.color}`}>{c.value}</p>
                   <p className="text-xs text-slate-400 mt-2 font-medium">{c.trend}</p>
                </div>
              ))}
           </div>

           {/* Alert Feed */}
           <div className="flex flex-col gap-4">
              <h3 className="text-lg font-bold text-white px-2">Live Detection Feed</h3>
              
              {loading ? (
                <div className="glass-card rounded-2xl p-6 text-slate-500">Loading alerts...</div>
              ) : error ? (
                <div className="glass-card rounded-2xl p-6 text-red-400">{error}</div>
              ) : alerts.length === 0 ? (
                <div className="glass-card rounded-2xl p-6 text-slate-500">No threat alerts found.</div>
              ) : alerts.map((alert) => (
                <div key={alert.id} className={`glass-card-hover glass-card rounded-2xl p-6 border-l-4 ${getSeverityStyles(alert.severity)}`}>
                   <div className="flex items-start justify-between">
                      <div className="flex gap-4">
                         <div className={`p-3 rounded-xl ${getSeverityStyles(alert.severity)} bg-opacity-20`}>
                            {getTypeIcon(alert.type)}
                         </div>
                         <div>
                            <div className="flex items-center gap-3">
                               <h4 className="text-lg font-bold text-white">{(alert.type || 'UNKNOWN').replace('_', ' ')}</h4>
                               <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase border ${getSeverityStyles(alert.severity)}`}>
                                 {alert.severity}
                               </span>
                            </div>
                            <p className="text-slate-300 mt-1 max-w-2xl">{alert.description || alert.message || 'No details provided.'}</p>
                            <div className="flex items-center gap-4 mt-4 text-xs text-slate-500 font-medium uppercase tracking-tighter">
                               <span className="flex items-center gap-1.5"><GlobeAltIcon className="w-4 h-4" /> {alert.ip || alert.ip_address || '—'}</span>
                               <span className="flex items-center gap-1.5"><ArrowPathIcon className="w-4 h-4" /> {format(new Date(alert.created_at || alert.created_at_ms || Date.now()), 'HH:mm:ss')}</span>
                               <span className="flex items-center gap-1.5"><EyeIcon className="w-4 h-4" /> Observed: 124 instances</span>
                            </div>
                         </div>
                      </div>

                      <div className="flex flex-col gap-2 min-w-[140px]">
                         <button
                            onClick={() => handleUpdateStatus(alert.id, 'blocked')}
                            disabled={updatingId === alert.id}
                            className="flex items-center justify-center gap-2 px-4 py-2 bg-vibrant-red hover:bg-red-700 text-white rounded-lg text-sm font-bold transition-all shadow-lg shadow-red-900/20 disabled:opacity-60"
                         >
                            <NoSymbolIcon className="w-4 h-4" />
                            {updatingId === alert.id ? 'BLOCKING...' : 'BLOCK IP'}
                         </button>
                         <button
                            onClick={() => handleUpdateStatus(alert.id, 'acknowledged')}
                            disabled={updatingId === alert.id}
                            className="flex items-center justify-center gap-2 px-4 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-lg text-sm font-bold transition-all border border-white/5 disabled:opacity-60"
                         >
                            <CheckCircleIcon className="w-4 h-4" />
                            {updatingId === alert.id ? 'UPDATING...' : 'ACKNOWLEDGE'}
                         </button>
                      </div>
                   </div>
                </div>
              ))}
           </div>

        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default ThreatAlerts;
