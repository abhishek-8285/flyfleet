import React from 'react';
import { useNavigate } from 'react-router-dom';
import EnterpriseAuthLayout from '../components/auth/EnterpriseAuthLayout';

const SessionExpired: React.FC = () => {
  const navigate = useNavigate();

  return (
    <EnterpriseAuthLayout 
      title="Session Expired" 
      subtitle="For your security, your session has timed out"
    >
      <div className="text-center space-y-8">
        <div className="relative mx-auto w-24 h-24">
          <div className="absolute inset-0 bg-red-500/20 rounded-full blur-2xl animate-pulse"></div>
          <div className="relative h-24 w-24 bg-slate-900/50 border border-white/10 rounded-full flex items-center justify-center shadow-2xl">
            <span className="material-symbols-outlined text-red-500 text-5xl">history</span>
          </div>
          <div className="absolute bottom-0 right-0 h-8 w-8 bg-red-500 rounded-full border-4 border-slate-900 flex items-center justify-center">
            <span className="material-symbols-outlined text-white text-sm font-bold">lock</span>
          </div>
        </div>

        <div className="space-y-4">
          <p className="text-slate-400 text-sm leading-relaxed px-4">
            Identity tokens have reached their maximum TTL. Please re-authenticate to restore your secure fleet access.
          </p>
          
          <button
            onClick={() => navigate('/login')}
            className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-4 rounded-2xl transition-all shadow-xl shadow-primary/20 flex items-center justify-center gap-2 text-sm uppercase tracking-widest mt-4"
          >
            Re-authenticate
            <span className="material-symbols-outlined text-sm">login</span>
          </button>
        </div>

        <div className="pt-6">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-slate-800/50 border border-white/5">
            <div className="w-1.5 h-1.5 rounded-full bg-red-500 animate-ping"></div>
            <span className="text-[9px] font-bold text-slate-500 uppercase tracking-[0.15em]">Token Revoked</span>
          </div>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default SessionExpired;
