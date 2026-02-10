import branding from '@flyfleet/config';
import { motion } from 'framer-motion';
import React from 'react';

interface EnterpriseAuthLayoutProps {
  children: React.ReactNode;
  title: string;
  subtitle?: string;
}

const EnterpriseAuthLayout: React.FC<EnterpriseAuthLayoutProps> = ({ children, title, subtitle }) => {
  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4 selection:bg-primary/30">
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/10 rounded-full blur-[120px] animate-pulse"></div>
        <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-600/10 rounded-full blur-[120px] animate-pulse" style={{ animationDelay: '2s' }}></div>
      </div>

      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
        className="max-w-md w-full relative z-10"
      >
        <div className="bg-slate-900/40 backdrop-blur-2xl border border-white/5 p-6 rounded-3xl shadow-[0_20px_50px_rgba(0,0,0,0.3)] border-t-white/10">
          <div className="text-center mb-6">
            <div className="flex justify-center mb-4">
              <div className="h-12 w-12 bg-gradient-to-br from-primary to-blue-600 rounded-2xl flex items-center justify-center shadow-lg shadow-primary/20 rotate-3 hover:rotate-0 transition-transform duration-300">
                <span className="material-symbols-outlined text-white text-2xl">local_shipping</span>
              </div>
            </div>
            <h1 className="text-2xl font-bold text-white tracking-tight mb-1">{branding.companyName}</h1>
            <div className="h-1 w-10 bg-primary mx-auto rounded-full mb-3"></div>
            <h2 className="text-lg font-semibold text-slate-200 mb-1">{title}</h2>
            {subtitle && <p className="text-slate-400 text-sm font-medium">{subtitle}</p>}
          </div>

          {children}

          <div className="mt-6 pt-4 border-t border-slate-800/50 text-center">
            <p className="text-slate-500 text-xs tracking-wide uppercase font-semibold mb-3">
              Secure Enterprise Gateway
            </p>
            <div className="flex justify-center gap-4 text-xs text-slate-500">
              <a href="#" className="hover:text-primary transition-colors">Privacy</a>
              <span className="text-slate-800">•</span>
              <a href="#" className="hover:text-primary transition-colors">Terms</a>
              <span className="text-slate-800">•</span>
              <a href="#" className="hover:text-primary transition-colors">Support</a>
            </div>
          </div>
        </div>
        
        <p className="text-center mt-6 text-slate-600 text-[10px] uppercase tracking-[0.2em] font-bold">
          Powered by FleetCommand Core v4.0
        </p>
      </motion.div>
    </div>
  );
};

export default EnterpriseAuthLayout;
