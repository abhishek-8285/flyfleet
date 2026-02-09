import {
    ArrowDownTrayIcon,
    ArrowPathIcon,
    ComputerDesktopIcon,
    DevicePhoneMobileIcon,
    MapPinIcon
} from '@heroicons/react/24/outline';
import { format } from 'date-fns';
import React, { useEffect, useState } from 'react';
import SecurityFooter from '../../../components/SecurityFooter';
import SecuritySidebar from '../../../components/SecuritySidebar';
import api from '../../../services/api';

const LoginActivityLogs: React.FC = () => {
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isLive, setIsLive] = useState(true);
  const [filter, setFilter] = useState('ALL');

  useEffect(() => {
    fetchLogs();
    let interval: any;
    if (isLive) {
      interval = setInterval(fetchLogs, 5000);
    }
    return () => clearInterval(interval);
  }, [isLive, filter]);

  const fetchLogs = async () => {
    try {
      const response = await api.getLoginActivityLogs({ 
        status: filter === 'ALL' ? undefined : filter,
        size: 50 
      });
      setLogs(response.events || []);
    } catch (error) {
      console.error('Failed to fetch logs', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'SUCCESS': return 'text-vibrant-green bg-emerald-500/10';
      case 'FAILED': return 'text-vibrant-red bg-red-500/10';
      case 'CHALLENGED': return 'text-vibrant-blue bg-blue-500/10';
      default: return 'text-slate-400 bg-slate-500/10';
    }
  };

  const getRiskBadge = (risk: string) => {
    switch (risk) {
      case 'high': return <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-red-500 text-white animate-pulse">High Risk</span>;
      case 'medium': return <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-amber-500 text-white">Medium Risk</span>;
      default: return <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-slate-500/20 text-slate-400">Low Risk</span>;
    }
  };

  return (
    <div className="flex h-screen bg-[#0b1117] text-slate-200 overflow-hidden font-inter">
      <SecuritySidebar />
      
      <main className="flex-1 flex flex-col min-w-0 bg-grid-pattern relative">
        <div className="glow-effect" />
        
        {/* Header */}
        <header className="px-8 py-6 flex items-center justify-between border-b border-white/5 bg-slate-900/40 backdrop-blur-md z-10">
          <div>
            <h2 className="text-2xl font-bold text-white flex items-center gap-3">
              <span className="material-symbols-outlined text-primary text-3xl">history</span>
              Login Activity Logs
            </h2>
            <p className="text-slate-400 text-sm mt-1">Real-time authentication monitoring across the enterprise fleet.</p>
          </div>
          
          <div className="flex items-center gap-4">
            <button 
              onClick={() => setIsLive(!isLive)}
              className={`flex items-center gap-2 px-4 py-2 rounded-full border transition-all ${
                isLive 
                ? 'bg-emerald-500/10 border-emerald-500/50 text-emerald-400 pulse-live' 
                : 'bg-slate-800 border-white/10 text-slate-400'
              }`}
            >
              <div className={`w-2 h-2 rounded-full ${isLive ? 'bg-emerald-400' : 'bg-slate-500'}`} />
              <span className="text-sm font-semibold">{isLive ? 'LIVE STREAMING' : 'PAUSED'}</span>
            </button>
            
            <div className="h-8 w-[1px] bg-white/10" />
            
            <button className="p-2 rounded-lg bg-slate-800 border border-white/10 text-slate-300 hover:bg-slate-700 transition-colors">
              <ArrowDownTrayIcon className="w-5 h-5" />
            </button>
          </div>
        </header>

        {/* Filters & Summary */}
        <div className="px-8 py-4 bg-slate-900/20 border-b border-white/5 flex items-center justify-between z-10">
          <div className="flex items-center gap-2">
            {['ALL', 'SUCCESS', 'FAILED', 'CHALLENGED'].map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${
                  filter === f 
                  ? 'bg-primary text-white shadow-lg shadow-primary/20' 
                  : 'text-slate-400 hover:text-white hover:bg-white/5'
                }`}
              >
                {f}
              </button>
            ))}
          </div>
          
          <div className="text-xs text-slate-500 flex items-center gap-6">
            <span>Showing last 50 attempts</span>
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-emerald-500" />
              <span>98.2% Success Rate</span>
            </div>
          </div>
        </div>

        {/* Logs Table */}
        <div className="flex-1 overflow-auto custom-scrollbar p-8 z-10">
          <div className="glass-card rounded-2xl overflow-hidden">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-white/5 bg-white/5">
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-400">Timestamp</th>
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-400">Identity</th>
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-400">Context</th>
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-400">Location & IP</th>
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-400">Status</th>
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-slate-400">Risk</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-20 text-center text-slate-500">
                      <ArrowPathIcon className="w-8 h-8 animate-spin mx-auto mb-4 opacity-20" />
                      Decrypting security logs...
                    </td>
                  </tr>
                ) : logs.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-6 py-20 text-center text-slate-500">
                      No matching authentication events found.
                    </td>
                  </tr>
                ) : (
                  logs.map((log) => (
                    <tr key={log.id} className="border-b border-white/5 hover:bg-white/[0.02] transition-colors group">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-white">{format(new Date(log.created_at), 'MMM dd, HH:mm:ss')}</div>
                        <div className="text-[10px] text-slate-500 font-mono mt-0.5">{log.id.slice(0, 8)}</div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary/40 to-blue-600/40 flex items-center justify-center text-xs font-bold text-white border border-white/10 uppercase">
                            {log.email.charAt(0)}
                          </div>
                          <div className="flex flex-col">
                            <span className="text-sm font-medium text-white">{log.email}</span>
                            <span className="text-[10px] text-slate-500 uppercase tracking-tighter">Enterprise User</span>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2 text-slate-300">
                           {log.device?.toLowerCase().includes('mobile') ? <DevicePhoneMobileIcon className="w-4 h-4 text-slate-500" /> : <ComputerDesktopIcon className="w-4 h-4 text-slate-500" />}
                           <span className="text-sm">{log.browser}</span>
                        </div>
                        <div className="text-[10px] text-slate-500 mt-0.5 truncate max-w-[150px]">{log.device}</div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2 text-slate-300">
                          <MapPinIcon className="w-4 h-4 text-primary" />
                          <span className="text-sm">{log.location}</span>
                        </div>
                        <div className="text-[10px] text-slate-500 font-mono mt-0.5">{log.ip_address}</div>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`px-3 py-1 rounded-full text-[10px] font-bold tracking-wider uppercase ${getStatusColor(log.status)}`}>
                          {log.status}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        {getRiskBadge(log.risk_level)}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default LoginActivityLogs;
