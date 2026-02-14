import React from 'react';

interface StatsGridProps {
  stats: {
    total_vehicles: number;
    active_trips: number;
    active_alerts: number;
    total_drivers?: number;
    online_drivers?: number;
    trends?: {
      vehicles: number;
      trips: number;
      alerts: number;
      drivers?: number;
    };
  };
}

export const StatsGrid: React.FC<StatsGridProps> = ({ stats }) => {
  const getTrendColor = (val: number) => val >= 0 ? 'text-emerald-500' : 'text-error';
  const getTrendIcon = (val: number) => val >= 0 ? 'trending_up' : 'trending_down';

  return (
    <div className="col-span-12 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      {/* Live Vehicles Card */}
      <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group flex flex-col justify-between">
        <div className="flex justify-between items-start z-10 relative">
          <div className="tooltip-trigger relative">
            <div className="flex items-center gap-1 mb-1">
              <p className="text-text-muted text-sm font-medium">Live Vehicles</p>
              <span className="material-symbols-outlined text-[14px] text-text-muted/50 hover:text-primary cursor-help">info</span>
            </div>
            <h3 className="text-3xl font-bold text-white tracking-tight">{stats.total_vehicles}</h3>
          </div>
          <div className="p-2 bg-primary/10 rounded-lg text-primary">
            <span className="material-symbols-outlined">directions_bus</span>
          </div>
        </div>
        <div className="mt-4 flex items-center text-sm">
          <span className={`${getTrendColor(stats.trends?.vehicles ?? 0)} font-medium flex items-center gap-1`}>
            <span className="material-symbols-outlined text-[16px]">{getTrendIcon(stats.trends?.vehicles ?? 0)}</span>
            {Math.abs(stats.trends?.vehicles ?? 0)}%
          </span>
          <span className="text-[#64748b] ml-2">vs last week</span>
        </div>
        <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
          <span>Source: Live IoT Gateway</span>
          <span className="text-emerald-500/80">Active</span>
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
            <h3 className="text-3xl font-bold text-white tracking-tight">{stats.active_trips}</h3>
          </div>
          <div className="p-2 bg-purple-500/10 rounded-lg text-purple-500">
            <span className="material-symbols-outlined">route</span>
          </div>
        </div>
        <div className="mt-4 flex items-center text-sm">
          <span className={`${getTrendColor(stats.trends?.trips ?? 0)} font-medium flex items-center gap-1`}>
            <span className="material-symbols-outlined text-[16px]">{getTrendIcon(stats.trends?.trips ?? 0)}</span>
            {Math.abs(stats.trends?.trips ?? 0)}%
          </span>
          <span className="text-[#64748b] ml-2">vs last week</span>
        </div>
        <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
          <span>Source: Routing Engine v4</span>
          <span className="text-text-muted">Live</span>
        </div>
      </div>

      {/* Drivers Card */}
      <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group flex flex-col justify-between">
        <div className="flex justify-between items-start z-10 relative">
          <div className="tooltip-trigger relative">
            <div className="flex items-center gap-1 mb-1">
              <p className="text-text-muted text-sm font-medium">Active Drivers</p>
              <span className="material-symbols-outlined text-[14px] text-text-muted/50 hover:text-primary cursor-help">info</span>
            </div>
            <h3 className="text-3xl font-bold text-white tracking-tight">{stats.online_drivers ?? 0}</h3>
          </div>
          <div className="p-2 bg-blue-500/10 rounded-lg text-blue-500">
            <span className="material-symbols-outlined">person</span>
          </div>
        </div>
        <div className="mt-4 flex items-center text-sm">
          <span className="text-white font-medium">{stats.total_drivers ?? 0}</span>
          <span className="text-[#64748b] ml-2">Total Registered</span>
        </div>
        <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
          <span>Source: Driver Service</span>
          <span className="text-text-muted">Synced</span>
        </div>
      </div>

      {/* Active Alerts Card */}
      <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group flex flex-col justify-between">
        <div className="flex justify-between items-start z-10 relative">
          <div className="w-full">
            <div className="flex justify-between items-start w-full">
              <div className="tooltip-trigger relative">
                <div className="flex items-center gap-1 mb-1 cursor-help">
                  <p className="text-text-muted text-sm font-medium">Critical Alerts</p>
                  <span className="material-symbols-outlined text-[14px] text-text-muted/50 hover:text-primary">info</span>
                </div>
              </div>
              <div className="p-2 bg-red-500/10 rounded-lg text-red-500">
                <span className="material-symbols-outlined">warning</span>
              </div>
            </div>
            <div className="flex items-baseline gap-2 mt-1">
              <h3 className="text-3xl font-bold text-white tracking-tight">{stats.active_alerts}</h3>
              <span className={`${getTrendColor(-(stats.trends?.alerts ?? 0))} text-xs font-medium`}>
                {(stats.trends?.alerts ?? 0) < 0 ? '↓' : '↑'} {Math.abs(stats.trends?.alerts ?? 0)}%
              </span>
            </div>
          </div>
        </div>
        <div className="mt-4 pt-3 border-t border-border-color text-[10px] text-[#64748b] flex justify-between items-center">
          <span>Source: Radar-VNet</span>
          <a className="text-primary hover:underline" href="/dashboard/alerts">Manage →</a>
        </div>
      </div>
    </div>
  );
};
