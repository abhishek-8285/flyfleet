import {
    ChevronRightIcon,
    GlobeAltIcon,
    MapIcon,
    NoSymbolIcon,
    PencilSquareIcon,
    PlusIcon,
    ShieldCheckIcon,
    TrashIcon
} from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import SecurityFooter from '../../../components/SecurityFooter';
import SecuritySidebar from '../../../components/SecuritySidebar';
import api from '../../../services/api';

const IPControls: React.FC = () => {
  const [rules, setRules] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchRules();
  }, []);

  const fetchRules = async () => {
    try {
      const response = await api.getIPRules('default');
      // Mocking for high-fidelity demonstration
      setRules([
        { id: '1', name: 'HQ Office Network', cidr: '203.0.113.0/24', action: 'ALLOW', type: 'STATIC', description: 'Main office building branch' },
        { id: '2', name: 'Untrusted VPS Network', cidr: '45.1.0.0/16', action: 'BLOCK', type: 'RANGE', description: 'Common source of scraping attempts' },
        { id: '3', name: 'Remote Fleet Admin', cidr: '1.2.3.4/32', action: 'ALLOW', type: 'STICKY', description: 'Lead analyst temporary home IP' },
      ]);
    } catch (error) {
       setRules([]);
    } finally {
      setLoading(false);
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
              <span className="material-symbols-outlined text-primary text-3xl">public</span>
              IP Access Governance
            </h2>
            <p className="text-slate-400 text-sm mt-1">Granular network control list and geofencing configuration.</p>
          </div>
          
          <button className="flex items-center gap-2 px-6 py-2.5 bg-primary hover:bg-primary-hover text-white rounded-xl font-bold transition-all shadow-lg shadow-primary/30">
             <PlusIcon className="w-5 h-5" />
             NEW SECURITY RULE
          </button>
        </header>

        {/* Content */}
        <div className="flex-1 overflow-auto custom-scrollbar p-8 z-10 flex flex-col gap-8">
           
           {/* Geofencing Preview */}
           <div className="glass-card rounded-2xl p-6 flex items-center justify-between border-l-4 border-blue-500">
              <div className="flex items-center gap-6">
                 <div className="p-4 rounded-2xl bg-blue-500/10 text-blue-400">
                    <MapIcon className="w-10 h-10" />
                 </div>
                 <div>
                    <h3 className="text-xl font-bold text-white">Geofencing Active</h3>
                    <p className="text-slate-400 text-sm mt-1">Primary traffic region: <span className="text-white font-semibold">South Asia (IN, SG)</span>. Requests from outside 12 restricted zones are being challenged.</p>
                 </div>
              </div>
              <button className="flex items-center gap-2 text-primary font-bold text-sm uppercase tracking-widest hover:text-white transition-colors">
                 ADVANCED GEOMAP <ChevronRightIcon className="w-4 h-4" />
              </button>
           </div>

           {/* Rules Table */}
           <div className="flex flex-col gap-4">
              <div className="flex items-center justify-between px-2">
                 <h3 className="text-lg font-bold text-white uppercase tracking-tighter">Active Network Rules</h3>
                 <span className="text-xs text-slate-500">Auto-synced with Fleet Firewall nodes</span>
              </div>

              <div className="glass-card rounded-2xl overflow-hidden">
                 <table className="w-full text-left">
                    <thead>
                       <tr className="border-b border-white/5 bg-white/[0.02]">
                          <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Rule Identification</th>
                          <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Network / CIDR</th>
                          <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Action</th>
                          <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Context</th>
                          <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500 text-right">Management</th>
                       </tr>
                    </thead>
                    <tbody className="divide-y divide-white/5">
                       {rules.map((rule) => (
                          <tr key={rule.id} className="hover:bg-white/[0.02] transition-colors group">
                             <td className="px-6 py-4">
                                <div className="flex flex-col">
                                   <span className="text-sm font-bold text-white">{rule.name}</span>
                                   <span className="text-[10px] text-slate-500 uppercase tracking-widest font-mono">ID: {rule.id}</span>
                                </div>
                             </td>
                             <td className="px-6 py-4">
                                <span className="bg-slate-800 px-3 py-1.5 rounded-lg font-mono text-sm text-primary border border-white/5">
                                   {rule.cidr}
                                </span>
                             </td>
                             <td className="px-6 py-4">
                                {rule.action === 'ALLOW' ? (
                                   <span className="flex items-center gap-1.5 text-vibrant-green font-bold text-xs">
                                      <ShieldCheckIcon className="w-4 h-4" /> ALLOW
                                   </span>
                                ) : (
                                   <span className="flex items-center gap-1.5 text-vibrant-red font-bold text-xs">
                                      <NoSymbolIcon className="w-4 h-4" /> BLOCK
                                   </span>
                                )}
                             </td>
                             <td className="px-6 py-4">
                                <p className="text-sm text-slate-400 max-w-[200px] truncate">{rule.description}</p>
                             </td>
                             <td className="px-6 py-4 text-right">
                                <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                   <button className="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 border border-white/10 transition-colors">
                                      <PencilSquareIcon className="w-4 h-4" />
                                   </button>
                                   <button className="p-2 rounded-lg bg-red-500/10 hover:bg-red-500 text-vibrant-red hover:text-white border border-red-500/20 transition-all">
                                      <TrashIcon className="w-4 h-4" />
                                   </button>
                                </div>
                             </td>
                          </tr>
                       ))}
                    </tbody>
                 </table>
              </div>
           </div>

           {/* Security Advice */}
           <div className="grid grid-cols-2 gap-6">
              <div className="glass-card p-6 rounded-2xl bg-blue-500/5">
                 <h4 className="flex items-center gap-2 text-white font-bold">
                    <GlobeAltIcon className="w-5 h-5 text-blue-400" />
                    CIDR Best Practices
                 </h4>
                 <p className="text-sm text-slate-400 mt-2">
                    Avoid using broad /16 ranges unless necessary. Use /32 for single hosts and /24 for local networks to minimize the attack surface.
                 </p>
              </div>
              <div className="glass-card p-6 rounded-2xl bg-emerald-500/5">
                 <h4 className="flex items-center gap-2 text-white font-bold">
                    <ShieldCheckIcon className="w-5 h-5 text-emerald-400" />
                    Auto-Propagation
                 </h4>
                 <p className="text-sm text-slate-400 mt-2">
                    Rules registered here are distributed to all edge proxies (BFF Layer) within 30 seconds for immediate enforcement.
                 </p>
              </div>
           </div>

        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default IPControls;
