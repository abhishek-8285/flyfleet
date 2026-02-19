import {
    ArrowRightOnRectangleIcon,
    ClockIcon,
    ComputerDesktopIcon,
    DevicePhoneMobileIcon,
    FingerPrintIcon,
    GlobeAltIcon,
    ShieldExclamationIcon,
    TrashIcon
} from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import SecurityFooter from '../../../components/SecurityFooter';
import SecuritySidebar from '../../../components/SecuritySidebar';
import api from '../../../services/api';

const ActiveSessions: React.FC = () => {
  const [sessions, setSessions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [revoking, setRevoking] = useState<string | null>(null);

  useEffect(() => {
    fetchSessions();
  }, []);

  const fetchSessions = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.getSecuritySessions('default');
      const data = (response as any)?.data ?? response;
      const items = data?.items || data?.sessions || data?.data || data || [];
      setSessions(Array.isArray(items) ? items : []);
    } catch (error) {
       setError('Failed to load active sessions.');
       setSessions([]);
    } finally {
      setLoading(false);
    }
  };

  const handleRevoke = async (id: string) => {
    if (window.confirm('Terminate this session immediately?')) {
       try {
         setRevoking(id);
         await api.revokeSecuritySession(id);
         setSessions((prev) => prev.filter((s) => s.id !== id));
       } catch (err) {
         alert('Failed to revoke session');
       } finally {
         setRevoking(null);
       }
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
              <span className="material-symbols-outlined text-primary text-3xl">devices</span>
              Active Auth Sessions
            </h2>
            <p className="text-slate-400 text-sm mt-1">Authorized access tokens and hardware signatures currently active.</p>
          </div>
          
          <button
             onClick={() => {
               if (!window.confirm('Revoke all other sessions?')) return;
               if ((api as any).revokeAllSecuritySessions) {
                 (api as any).revokeAllSecuritySessions().then(fetchSessions).catch(() => {
                   alert('Failed to revoke all sessions');
                 });
               } else {
                 alert('Revoke-all API not available');
               }
             }}
             className="flex items-center gap-2 px-4 py-2 border border-red-500/30 bg-red-500/5 hover:bg-red-500/10 text-red-500 rounded-lg text-sm font-bold transition-all"
          >
             <ArrowRightOnRectangleIcon className="w-5 h-5" />
             REVOKE ALL OTHER SESSIONS
          </button>
        </header>

        {/* Sessions Grid */}
        <div className="flex-1 overflow-auto custom-scrollbar p-8 z-10 flex flex-col gap-6">
           
           <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {loading ? (
                 <div className="col-span-full text-slate-500">Loading sessions...</div>
              ) : error ? (
                 <div className="col-span-full text-red-400">{error}</div>
              ) : sessions.length === 0 ? (
                 <div className="col-span-full text-slate-500">No active sessions found.</div>
              ) : sessions.map((session) => (
                <div key={session.id} className={`glass-card p-6 rounded-2xl relative transition-all group ${session.isCurrent ? 'ring-2 ring-primary/50 bg-primary/5' : 'hover:bg-white/[0.03]'}`}>
                   
                   {session.isCurrent && (
                     <span className="absolute top-4 right-4 px-2 py-0.5 rounded bg-primary text-white text-[10px] font-bold tracking-widest animate-pulse">
                        CURRENT DEVICE
                     </span>
                   )}

                   <div className="flex items-center gap-4 mb-6">
                      <div className="w-14 h-14 rounded-2xl bg-slate-800 flex items-center justify-center border border-white/5 shadow-inner group-hover:bg-slate-700 transition-colors">
                         {(session.device || '').toLowerCase().includes('iphone') || (session.device || '').toLowerCase().includes('android')
                           ? <DevicePhoneMobileIcon className="w-8 h-8 text-primary" />
                           : <ComputerDesktopIcon className="w-8 h-8 text-primary" />}
                      </div>
                      <div>
                         <h4 className="text-lg font-bold text-white">{session.device || session.device_name || 'Unknown Device'}</h4>
                         <p className="text-sm text-slate-400">{session.browser || session.user_agent || 'Unknown Client'}</p>
                      </div>
                   </div>

                   <div className="space-y-4">
                      <div className="flex items-center gap-3 text-sm text-slate-300">
                         <GlobeAltIcon className="w-5 h-5 text-slate-500" />
                         <span>{session.ip || session.ip_address || '—'} • <span className="text-slate-500">{session.location || session.geo || 'Unknown'}</span></span>
                      </div>
                      <div className="flex items-center gap-3 text-sm text-slate-300">
                         <ClockIcon className="w-5 h-5 text-slate-500" />
                         <span>Last active: <span className="font-semibold text-white">{session.lastActive || session.last_active_at || session.last_seen || '—'}</span></span>
                      </div>
                      <div className="flex items-center gap-3 text-sm text-slate-300">
                         <FingerPrintIcon className="w-5 h-5 text-slate-500" />
                         <span className="font-mono text-[10px] text-slate-500">ID: {session.id}</span>
                      </div>
                   </div>

                   <div className="mt-8 pt-6 border-t border-white/5">
                      {!session.isCurrent && (
                        <button 
                          onClick={() => handleRevoke(session.id)}
                          disabled={revoking === session.id}
                          className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl text-red-400 hover:text-white hover:bg-red-500 font-bold text-sm transition-all border border-red-500/20 disabled:opacity-60"
                        >
                          <TrashIcon className="w-4 h-4" />
                          {revoking === session.id ? 'TERMINATING...' : 'TERMINATE SESSION'}
                        </button>
                      )}
                      {session.isCurrent && (
                        <div className="text-center py-2.5 text-slate-500 text-xs font-medium uppercase tracking-widest italic">
                           Secure Session Primary
                        </div>
                      )}
                   </div>
                </div>
              ))}
           </div>

           {/* Security Note */}
           <div className="glass-card p-6 rounded-2xl bg-amber-500/5 border-amber-500/20">
              <div className="flex gap-4">
                 <div className="p-2 rounded-lg bg-amber-500/10 text-amber-500 h-fit">
                    <ShieldExclamationIcon className="w-6 h-6" />
                 </div>
                 <div>
                    <h4 className="text-white font-bold">Session Hardening Active</h4>
                    <p className="text-sm text-slate-400 mt-1 max-w-4xl">
                      Each session above is bound to a specific hardware signature (TLS Fingerprinting). 
                      Changing devices or network contexts may trigger a re-authentication challenge. 
                      Unauthorized sessions will be automatically terminated by the Threat Engine.
                    </p>
                 </div>
              </div>
           </div>

        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default ActiveSessions;
