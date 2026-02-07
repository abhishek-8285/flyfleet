import React from 'react';
import { useNavigate } from 'react-router-dom';
import EnterpriseAuthLayout from '../components/auth/EnterpriseAuthLayout';

const ErrorScreen: React.FC<{ message?: string }> = ({ message }) => {
  const navigate = useNavigate();

  return (
    <EnterpriseAuthLayout 
      title="System Breach" 
      subtitle="A security anomaly has been detected"
    >
      <div className="text-center space-y-8">
        <div className="relative mx-auto w-24 h-24">
          <div className="absolute inset-0 bg-red-500/20 rounded-full blur-2xl animate-pulse"></div>
          <div className="relative h-24 w-24 bg-slate-900/50 border border-white/10 rounded-full flex items-center justify-center shadow-2xl">
            <span className="material-symbols-outlined text-red-500 text-5xl">warning</span>
          </div>
        </div>

        <div className="space-y-4">
          <p className="text-slate-300 text-sm font-bold uppercase tracking-widest">
            {message || 'Internal Protocol Failure'}
          </p>
          <p className="text-slate-500 text-xs leading-relaxed px-4">
            Security policies prevented the request from completing. Our systems have logged this event for administrative review.
          </p>
          
          <button
            onClick={() => navigate('/login')}
            className="w-full bg-slate-800/50 hover:bg-slate-700 text-white font-bold py-4 rounded-2xl transition-all border border-white/5 flex items-center justify-center gap-2 text-sm uppercase tracking-widest mt-4"
          >
            Return to Safety
          </button>
        </div>

        <div className="pt-6 grid grid-cols-2 gap-4">
          <div className="p-3 rounded-xl bg-slate-800/20 border border-white/5 text-center">
            <p className="text-[8px] font-bold text-slate-600 uppercase tracking-widest mb-1">Error Code</p>
            <p className="text-[10px] font-mono font-bold text-slate-400">SEC_FLT_403</p>
          </div>
          <div className="p-3 rounded-xl bg-slate-800/20 border border-white/5 text-center">
            <p className="text-[8px] font-bold text-slate-600 uppercase tracking-widest mb-1">Status</p>
            <p className="text-[10px] font-mono font-bold text-red-500">DENIED</p>
          </div>
        </div>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default ErrorScreen;
