import React from 'react';

interface StatsGridProps {
  stats: {
    total_vehicles: number;
    active_trips: number;
    active_alerts: number;
  };
}

export const StatsGrid: React.FC<StatsGridProps> = ({ stats }) => {
  return (
    <div className="col-span-12 grid grid-cols-1 md:grid-cols-3 gap-6">
      {/* Live Vehicles Card */}
      <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group flex flex-col justify-between">
        <div className="flex justify-between items-start z-10 relative">
          <div className="tooltip-trigger relative">
            <div className="flex items-center gap-1 mb-1">
              <p className="text-text-muted text-sm font-medium">Live Vehicles</p>
              <span className="material-symbols-outlined text-[14px] text-text-muted/50 hover:text-primary cursor-help">info</span>
            </div>
            <div className="hidden absolute top-full left-0 mt-2 p-2 bg-surface-darker border border-border-color rounded shadow-xl text-[10px] text-text-muted w-48 z-50 group-hover:block">
              Count of vehicles with 'Ignition On' signal received within last 5 minutes.
            </div>
            <h3 className="text-3xl font-bold text-white tracking-tight">{stats.total_vehicles || 842}</h3>
          </div>
          <div className="p-2 bg-primary/10 rounded-lg text-primary">
            <span className="material-symbols-outlined">directions_bus</span>
          </div>
        </div>
        <div className="mt-4 flex items-center text-sm">
          <span className="text-emerald-500 font-medium flex items-center gap-1">
            <span className="material-symbols-outlined text-[16px]">trending_up</span>
            5%
          </span>
          <span className="text-[#64748b] ml-2">vs last week</span>
        </div>
        <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
          <span>Source: Telematics V2</span>
          <span className="text-emerald-500/80">99.9% Uptime</span>
        </div>
      </div>

      {/* Active Trips Card */}
      <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group flex flex-col justify-between">
        <div className="flex justify-between items-start z-10 relative">
          <div className="tooltip-trigger relative">
            <div className="flex items-center gap-1 mb-1">
              <p className="text-text-muted text-sm font-medium">Active Trips</p>
              <span className="material-symbols-outlined text-[14px] text-text-muted/50 hover:text-primary cursor-help">info</span>
            </div>
            <h3 className="text-3xl font-bold text-white tracking-tight">{stats.active_trips || 125}</h3>
          </div>
          <div className="p-2 bg-purple-500/10 rounded-lg text-purple-500">
            <span className="material-symbols-outlined">route</span>
          </div>
        </div>
        <div className="mt-4 flex items-center text-sm">
          <span className="text-emerald-500 font-medium flex items-center gap-1">
            <span className="material-symbols-outlined text-[16px]">trending_up</span>
            2%
          </span>
          <span className="text-[#64748b] ml-2">vs last week</span>
        </div>
        <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
          <span>Source: Dispatch API</span>
          <span className="text-text-muted">Updated 10s ago</span>
        </div>
      </div>

      {/* Active Alerts Card */}
      <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group flex flex-col justify-between">
        <div className="flex justify-between items-start z-10 relative">
          <div className="w-full">
            <div className="flex justify-between items-start w-full">
              <div className="tooltip-trigger relative">
                <div className="flex items-center gap-1 mb-1 cursor-help">
                  <p className="text-text-muted text-sm font-medium">Active Alerts</p>
                  <span className="material-symbols-outlined text-[14px] text-text-muted/50 hover:text-primary">info</span>
                </div>
                <div className="hidden absolute top-full left-0 mt-2 p-2 bg-surface-darker border border-border-color rounded shadow-xl text-[10px] text-text-muted w-48 z-20 group-hover:block">
                  KPI Logic: Unresolved vehicle health or driving behavior alerts requiring immediate action.
                </div>
              </div>
              <div className="p-2 bg-red-500/10 rounded-lg text-red-500">
                <span className="material-symbols-outlined">warning</span>
              </div>
            </div>
            <div className="flex items-baseline gap-2 mt-1">
              <h3 className="text-3xl font-bold text-white tracking-tight">{stats.active_alerts || 3}</h3>
              <span className="text-xs text-text-muted">Total</span>
            </div>
            <div className="flex gap-3 mt-3">
              <div className="relative group/alert">
                <a href="/dashboard/maintenance" className="flex items-center gap-1.5 px-2 py-1 rounded bg-red-500/10 border border-red-500/20 hover:bg-red-500/20 transition-colors cursor-pointer">
                  <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse"></div>
                  <span className="text-xs font-semibold text-red-400">2 Critical</span>
                </a>
                {/* Quick Actions Popup */}
                <div className="absolute bottom-full left-0 mb-2 w-40 hidden group-hover/alert:block bg-surface-darker border border-border-color rounded shadow-xl z-20 overflow-hidden">
                    <div className="text-[10px] font-semibold text-text-muted px-2 py-1 bg-surface-dark border-b border-border-color">Quick Actions</div>
                    <button className="w-full text-left px-2 py-1.5 text-xs text-white hover:bg-primary hover:text-white transition-colors">Acknowledge All</button>
                    <button className="w-full text-left px-2 py-1.5 text-xs text-white hover:bg-primary hover:text-white transition-colors">Assign to Team</button>
                </div>
              </div>
              <a href="/dashboard/maintenance" className="flex items-center gap-1.5 px-2 py-1 rounded bg-orange-500/10 border border-orange-500/20 hover:bg-orange-500/20 transition-colors cursor-pointer">
                <div className="w-2 h-2 rounded-full bg-orange-500"></div>
                <span className="text-xs font-semibold text-orange-400">1 Warning</span>
              </a>
            </div>
          </div>
        </div>
        <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
          <span>Source: Event Stream</span>
          <a className="text-primary hover:underline" href="/dashboard/history">View Log →</a>
        </div>
      </div>
    </div>
  );
};
