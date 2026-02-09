import {
    ArrowPathIcon,
    ClockIcon,
    FingerPrintIcon,
    KeyIcon,
    PlusIcon,
    TrashIcon
} from '@heroicons/react/24/outline';
import React, { useState } from 'react';

const ApiClientHub: React.FC = () => {
  const [showKeyModal, setShowKeyModal] = useState(false);
  
  const clients = [
    {
      id: '1',
      name: 'Logistics Pro Backend',
      clientID: 'cli_8x92m...k29s',
      scopes: ['read:fleet', 'write:trips'],
      expiry: 'Dec 31, 2024',
      status: 'active'
    },
    {
      id: '2',
      name: 'Data Analytics Feed',
      clientID: 'cli_3v88q...z99l',
      scopes: ['read:all'],
      expiry: 'May 01, 2024',
      status: 'expiring'
    }
  ];

  return (
    <div className="space-y-6">
      {/* Stats Row */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[
          { label: 'Active Clients', value: '24', icon: FingerPrintIcon, color: 'text-primary' },
          { label: 'Total Requests (24h)', value: '1.2M', icon: ArrowPathIcon, color: 'text-emerald-500' },
          { label: 'Keys Expiring', value: '3', icon: ClockIcon, color: 'text-amber-500' },
        ].map((stat, i) => (
          <div key={i} className="bg-slate-900/40 border border-slate-800 p-6 rounded-2xl flex items-center justify-between backdrop-blur-sm">
            <div>
              <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">{stat.label}</p>
              <p className="text-2xl font-bold text-white">{stat.value}</p>
            </div>
            <stat.icon className={`w-8 h-8 ${stat.color} opacity-20`} />
          </div>
        ))}
      </div>

      <div className="bg-slate-900/40 border border-slate-800 rounded-2xl overflow-hidden backdrop-blur-sm">
        <div className="p-6 border-b border-white/5 flex justify-between items-center">
          <h3 className="text-lg font-bold text-white flex items-center gap-2">
            <KeyIcon className="w-5 h-5 text-primary" />
            Managed API Clients
          </h3>
          <button 
            onClick={() => setShowKeyModal(true)}
            className="flex items-center gap-2 bg-primary/10 hover:bg-primary/20 text-primary border border-primary/20 px-4 py-2 rounded-xl text-xs font-bold transition-all"
          >
            <PlusIcon className="w-4 h-4" />
            Generate New Key
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="bg-slate-950/50 text-[10px] font-bold text-slate-500 uppercase tracking-widest border-b border-white/5">
                <th className="px-6 py-4">Client Name / Owner</th>
                <th className="px-6 py-4">Client ID</th>
                <th className="px-6 py-4">Scopes</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {clients.map((client) => (
                <tr key={client.id} className="hover:bg-white/[0.02] transition-colors group">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded bg-gradient-to-br from-primary to-indigo-600 flex items-center justify-center text-[10px] font-bold">
                        {client.name.split(' ').map(n => n[0]).join('')}
                      </div>
                      <div>
                        <div className="text-sm font-medium text-white">{client.name}</div>
                        <div className="text-[10px] text-slate-500 font-mono">ID: {client.id}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <code className="text-xs bg-slate-950 px-2 py-1 rounded border border-white/5 text-primary font-mono cursor-pointer hover:border-primary/50 transition-all">
                      {client.clientID}
                    </code>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex gap-1">
                      {client.scopes.map(s => (
                        <span key={s} className="px-2 py-0.5 rounded-lg bg-primary/10 text-primary border border-primary/20 text-[10px] font-bold">
                          {s}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-lg text-[10px] font-bold uppercase tracking-widest ${
                      client.status === 'active' 
                        ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20' 
                        : 'bg-amber-500/10 text-amber-500 border border-amber-500/20'
                    }`}>
                      <span className={`w-1 h-1 rounded-full ${client.status === 'active' ? 'bg-emerald-500' : 'bg-amber-500 animate-pulse'}`} />
                      {client.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button className="p-2 text-slate-500 hover:text-white hover:bg-slate-800 rounded-xl transition-all" title="Rotate Key">
                        <ArrowPathIcon className="w-4 h-4" />
                      </button>
                      <button className="p-2 text-slate-500 hover:text-rose-500 hover:bg-rose-500/10 rounded-xl transition-all" title="Revoke Key">
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
    </div>
  );
};

export default ApiClientHub;
