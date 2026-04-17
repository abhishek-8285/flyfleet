import React from 'react';

const StatTile = ({ title, value, icon, color, trend }: any) => (
  <div className="bg-surface-container border border-outline rounded-[2rem] p-8 relative overflow-hidden group hover:bg-surface-container-highest transition-all duration-500 shadow-xl">
    <div className="absolute top-0 right-0 p-8 opacity-[0.03] group-hover:opacity-[0.08] transition-all transform group-hover:scale-110 duration-700">
      <span className="material-symbols-outlined text-8xl">{icon}</span>
    </div>
    
    <div className="relative space-y-6">
      <div className="flex justify-between items-center">
        <h4 className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.25em]">{title}</h4>
        <div className={`h-1.5 w-1.5 rounded-full ${color === 'rose' ? 'bg-rose-500 animate-ping shadow-[0_0_10px_rgba(244,63,94,0.5)]' : color === 'amber' ? 'bg-amber-500' : 'bg-emerald-500'}`} />
      </div>
      
      <div>
        <h3 className={`text-4xl font-bold tracking-tighter uppercase font-mono ${color === 'rose' ? 'text-rose-400' : color === 'amber' ? 'text-amber-400' : 'text-white'}`}>
          {value}
        </h3>
      </div>
      
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-1 font-black text-[9px] text-on-surface-variant bg-surface-dim border border-outline px-2 py-1 rounded-lg opacity-60">
          <span className="material-symbols-outlined text-xs">analytics</span>
          {trend || 'Stabilized'}
        </div>
      </div>
    </div>
  </div>
);

export const MaintenanceStats = () => {
    return (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <StatTile 
                title="Operational Readiness" 
                value="98.2%" 
                icon="verified" 
                color="emerald" 
                trend="+1.4% Efficiency"
            />
            <StatTile 
                title="Pending Triage" 
                value="14" 
                icon="assignment_late" 
                color="amber" 
                trend="Queued Assets"
            />
            <StatTile 
                title="Critical Failure" 
                value="03" 
                icon="warning" 
                color="rose" 
                trend="High Priority"
            />
        </div>
    );
};
