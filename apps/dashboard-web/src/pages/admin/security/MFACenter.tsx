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
import React, { useEffect, useState } from 'react';
import SecurityFooter from '../../../components/SecurityFooter';
import SecuritySidebar from '../../../components/SecuritySidebar';
import api from '../../../services/api';

const MFACenter: React.FC = () => {
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    fetchMFAUsers();
  }, []);

  const fetchMFAUsers = async () => {
    try {
      const response = await api.getMFAEvents('default'); // Placeholder for MFACenter needs a "list users" endpoint
      // Mocking for high-fidelity demonstration since real endpoint is still being mapped
      setUsers([
        { id: '1', email: 'admin@flyfleet.com', name: 'Abhishek', status: 'ENABLED', method: 'TOTP', devices: 2, lastUsed: '2 mins ago' },
        { id: '2', email: 'driver_alpha@flyfleet.com', name: 'John Doe', status: 'ENFORCED', method: 'SMS', devices: 1, lastUsed: '1 hour ago' },
        { id: '3', email: 'manager@flyfleet.com', name: 'Sarah Connor', status: 'DISABLED', method: 'NONE', devices: 0, lastUsed: 'Never' },
        { id: '4', email: 'api_key_system@flyfleet.com', name: 'System Service', status: 'BYPASSED', method: 'Cert', devices: 1, lastUsed: 'Just now' },
      ]);
    } catch (error) {
       setUsers([]);
    } finally {
      setLoading(false);
    }
  };

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
        
        {/* Header */}
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

        {/* Stats Strip */}
        <div className="px-8 py-6 grid grid-cols-4 gap-6 z-10">
          {[
            { label: 'Total Users', value: '1,284', icon: UserGroupIcon, color: 'text-blue-400' },
            { label: 'MFA Enabled', value: '94%', icon: ShieldCheckIcon, color: 'text-emerald-400' },
            { label: 'Trusted Devices', value: '2,103', icon: DevicePhoneMobileIcon, color: 'text-purple-400' },
            { label: 'Policy Violations', value: '0', icon: ExclamationTriangleIcon, color: 'text-amber-400' },
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

        {/* User Directory */}
        <div className="flex-1 px-8 pb-8 flex flex-col min-h-0 z-10">
           <div className="glass-card rounded-2xl flex flex-col flex-1 overflow-hidden">
             
             {/* Toolbar */}
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
                   <button className="p-2 text-slate-400 hover:text-white transition-colors">
                      <ArrowPathIcon className="w-5 h-5" />
                   </button>
                </div>
             </div>

             <div className="flex-1 overflow-auto custom-scrollbar">
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
                    {users.map((u) => (
                      <tr key={u.id} className="hover:bg-white/[0.02] transition-colors group">
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <div className="w-10 h-10 rounded-lg bg-slate-800 flex items-center justify-center border border-white/5 shadow-inner">
                              <UserGroupIcon className="w-5 h-5 text-slate-400" />
                            </div>
                            <div className="flex flex-col">
                              <span className="text-sm font-semibold text-white">{u.name}</span>
                              <span className="text-xs text-slate-500 font-mono tracking-tighter">{u.email}</span>
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
                          <span className="text-sm text-slate-300 font-bold">{u.devices} Devices</span>
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
             </div>
           </div>
        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default MFACenter;
