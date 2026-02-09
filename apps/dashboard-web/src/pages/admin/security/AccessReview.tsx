import {
    ArrowPathIcon,
    CheckBadgeIcon,
    ClipboardDocumentCheckIcon,
    ClockIcon,
    ExclamationCircleIcon,
    FunnelIcon,
    MagnifyingGlassIcon
} from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import SecurityFooter from '../../../components/SecurityFooter';
import SecuritySidebar from '../../../components/SecuritySidebar';

const AccessReview: React.FC = () => {
  const [reviews, setReviews] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchReviews();
  }, []);

  const fetchReviews = async () => {
    try {
      // Mocking for high-fidelity demonstration
      setReviews([
        { id: 'rev_1', user: 'Abhishek', email: 'admin@flyfleet.com', roles: ['Super Admin', 'Org Admin'], lastReview: '3 months ago', status: 'pending', risk: 'high' },
        { id: 'rev_2', user: 'John Doe', email: 'driver_alpha@flyfleet.com', roles: ['Driver'], lastReview: '1 month ago', status: 'certified', risk: 'low' },
        { id: 'rev_3', user: 'Sarah Connor', email: 'manager@flyfleet.com', roles: ['Org Admin'], lastReview: '6 months ago', status: 'overdue', risk: 'medium' },
      ]);
    } catch (error) {
       setReviews([]);
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'certified': return <span className="flex items-center gap-1 text-emerald-400 font-bold text-[10px] uppercase bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20"><CheckBadgeIcon className="w-3 h-3"/> Certified</span>;
      case 'overdue': return <span className="flex items-center gap-1 text-red-400 font-bold text-[10px] uppercase bg-red-500/10 px-2 py-0.5 rounded border border-red-500/20"><ExclamationCircleIcon className="w-3 h-3"/> Overdue</span>;
      default: return <span className="flex items-center gap-1 text-amber-400 font-bold text-[10px] uppercase bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/20"><ClockIcon className="w-3 h-3"/> Pending</span>;
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
              <span className="material-symbols-outlined text-primary text-3xl">fact_check</span>
              Access Certification Review
            </h2>
            <p className="text-slate-400 text-sm mt-1">Periodic validation of user permissions and resource access levels.</p>
          </div>
          
          <div className="flex items-center gap-3">
             <button className="flex items-center gap-2 px-4 py-2 border border-white/10 bg-slate-800 hover:bg-slate-700 text-white rounded-lg font-bold transition-all">
                <ArrowPathIcon className="w-5 h-5" />
                INITIATE BULK REVIEW
             </button>
          </div>
        </header>

        {/* Content */}
        <div className="flex-1 overflow-auto custom-scrollbar p-8 z-10 flex flex-col gap-6">
           
           {/* Certification Progress */}
           <div className="grid grid-cols-4 gap-6">
              {[
                { label: 'Total items', value: '482', color: 'text-white' },
                { label: 'Certified', value: '312', color: 'text-emerald-400' },
                { label: 'Revoked', value: '14', color: 'text-red-400' },
                { label: 'Completion', value: '64%', color: 'text-primary' },
              ].map((stat, i) => (
                <div key={i} className="glass-card p-4 rounded-xl">
                   <p className="text-[10px] font-bold uppercase tracking-widest text-slate-500">{stat.label}</p>
                   <p className={`text-2xl font-black mt-1 ${stat.color}`}>{stat.value}</p>
                   <div className="mt-3 h-1 w-full bg-slate-800 rounded-full overflow-hidden">
                      <div className="h-full bg-current opacity-20" style={{width: i === 3 ? '64%' : '100%'}} />
                   </div>
                </div>
              ))}
           </div>

           {/* Review Table */}
           <div className="glass-card rounded-2xl flex flex-col overflow-hidden">
             
             {/* Toolbar */}
             <div className="p-4 border-b border-white/5 bg-white/5 flex items-center justify-between">
                <div className="flex items-center gap-4">
                   <div className="relative w-80">
                      <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                      <input 
                         type="text" 
                         placeholder="Filter reviewer or subject..."
                         className="w-full bg-slate-900/50 border border-white/10 rounded-lg py-1.5 pl-10 pr-4 text-sm focus:outline-none focus:ring-1 focus:ring-primary/50"
                      />
                   </div>
                   <button className="flex items-center gap-2 text-xs font-bold text-slate-400 hover:text-white transition-colors">
                      <FunnelIcon className="w-4 h-4" /> FILTERS
                   </button>
                </div>
                <div className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">
                   Review Period: OCT 2026 - DEC 2026
                </div>
             </div>

             <div className="flex-1 overflow-auto custom-scrollbar">
                <table className="w-full text-left">
                   <thead>
                      <tr className="border-b border-white/5 bg-white/[0.02]">
                         <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Subject Identity</th>
                         <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Access Scope</th>
                         <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Last Reviewed</th>
                         <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500">Certification</th>
                         <th className="px-6 py-4 text-xs font-bold uppercase text-slate-500 text-right">Actions</th>
                      </tr>
                   </thead>
                   <tbody className="divide-y divide-white/5">
                      {reviews.map((r) => (
                         <tr key={r.id} className="hover:bg-white/[0.02] transition-colors group">
                            <td className="px-6 py-4">
                               <div className="flex items-center gap-3">
                                  <div className="w-10 h-10 rounded-full bg-slate-800 border border-white/10 flex items-center justify-center text-xs font-bold text-slate-400">
                                     {r.user.charAt(0)}
                                  </div>
                                  <div className="flex flex-col">
                                     <span className="text-sm font-bold text-white">{r.user}</span>
                                     <span className="text-xs text-slate-500 font-mono tracking-tighter">{r.email}</span>
                                  </div>
                               </div>
                            </td>
                            <td className="px-6 py-4">
                               <div className="flex flex-wrap gap-1.5">
                                  {r.roles.map((role: string) => (
                                     <span key={role} className="px-2 py-0.5 rounded bg-slate-800 text-slate-300 text-[10px] font-medium border border-white/5">
                                        {role}
                                     </span>
                                  ))}
                               </div>
                            </td>
                            <td className="px-6 py-4">
                               <span className="text-xs text-slate-400">{r.lastReview}</span>
                            </td>
                            <td className="px-6 py-4">
                               {getStatusBadge(r.status)}
                            </td>
                            <td className="px-6 py-4 text-right">
                               <div className="flex items-center justify-end gap-2">
                                  <button className="px-3 py-1.5 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-[10px] font-bold uppercase tracking-wider transition-all shadow-md shadow-emerald-900/10">
                                     CERTIFY
                                  </button>
                                  <button className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-[10px] font-bold uppercase tracking-wider border border-white/5 transition-all">
                                     MODIFY
                                  </button>
                               </div>
                            </td>
                         </tr>
                      ))}
                   </tbody>
                </table>
             </div>
           </div>

           {/* Security Insight */}
           <div className="glass-card p-6 rounded-2xl bg-slate-800/20 border-white/5 flex gap-6 items-center">
              <div className="p-4 rounded-xl bg-primary/10 text-primary">
                 <ClipboardDocumentCheckIcon className="w-10 h-10" />
              </div>
              <div className="flex-1">
                 <h4 className="text-white font-bold">Policy Enforcement</h4>
                 <p className="text-sm text-slate-400 mt-1 max-w-4xl">
                   Access Review is a SOC2 Type II requirement. Failure to certify active accounts within the 90-day window will trigger automatic 
                   mitigation: higher MFA sensitivity and mandatory manager approval for sensitive operations.
                 </p>
              </div>
              <button className="px-6 py-2.5 rounded-xl bg-slate-800 text-white font-bold text-sm border border-white/10 hover:bg-slate-700 transition-all">
                 VIEW POLICY
              </button>
           </div>

        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default AccessReview;
