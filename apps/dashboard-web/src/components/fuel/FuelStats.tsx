import React from 'react';
import { FuelLogStats } from '../../types/fuel';

interface FuelStatsProps {
    stats: FuelLogStats;
}

export const FuelStats: React.FC<FuelStatsProps> = ({ stats }) => {
    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="bg-surface-dark/40 backdrop-blur-md border border-border-color p-5 rounded-2xl flex items-center gap-4 hover:border-primary/50 transition-all duration-300 group">
                <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center text-primary group-hover:scale-110 transition-transform">
                    <span className="material-symbols-outlined text-[28px]">stat_3</span>
                </div>
                <div>
                    <div className="text-text-muted text-xs font-medium uppercase tracking-wider">Total Consumption</div>
                    <div className="text-2xl font-bold text-white mt-0.5">{stats.total_consumption.toLocaleString()} L</div>
                    <div className="text-[10px] text-green-400 mt-1 flex items-center gap-1">
                        <span className="material-symbols-outlined text-[12px]">trending_up</span>
                        +2.4% from last month
                    </div>
                </div>
            </div>

            <div className="bg-surface-dark/40 backdrop-blur-md border border-border-color p-5 rounded-2xl flex items-center gap-4 hover:border-accent-yellow/50 transition-all duration-300 group">
                <div className="w-12 h-12 rounded-xl bg-accent-yellow/10 flex items-center justify-center text-accent-yellow group-hover:scale-110 transition-transform">
                    <span className="material-symbols-outlined text-[28px]">payments</span>
                </div>
                <div>
                    <div className="text-text-muted text-xs font-medium uppercase tracking-wider">Total Cost</div>
                    <div className="text-2xl font-bold text-white mt-0.5">${stats.total_cost.toLocaleString()}</div>
                    <div className="text-[10px] text-text-muted mt-1 italic">
                        Avg: ${(stats.total_cost / (stats.total_consumption || 1)).toFixed(2)}/L
                    </div>
                </div>
            </div>

            <div className="bg-surface-dark/40 backdrop-blur-md border border-border-color p-5 rounded-2xl flex items-center gap-4 hover:border-accent-blue/50 transition-all duration-300 group">
                <div className="w-12 h-12 rounded-xl bg-accent-blue/10 flex items-center justify-center text-accent-blue group-hover:scale-110 transition-transform">
                    <span className="material-symbols-outlined text-[28px]">speed</span>
                </div>
                <div>
                    <div className="text-text-muted text-xs font-medium uppercase tracking-wider">Efficiency</div>
                    <div className="text-2xl font-bold text-white mt-0.5">{stats.avg_consumption.toFixed(2)} km/L</div>
                    <div className="text-[10px] text-text-muted mt-1">
                        Fleet target: 12.5 km/L
                    </div>
                </div>
            </div>

            <div className="bg-surface-dark/40 backdrop-blur-md border border-border-color p-5 rounded-2xl flex items-center gap-4 hover:border-orange-500/50 transition-all duration-300 group">
                <div className="w-12 h-12 rounded-xl bg-orange-500/10 flex items-center justify-center text-orange-500 group-hover:scale-110 transition-transform">
                    <span className="material-symbols-outlined text-[28px]">pending_actions</span>
                </div>
                <div>
                    <div className="text-text-muted text-xs font-medium uppercase tracking-wider">Pending Approvals</div>
                    <div className="text-2xl font-bold text-white mt-0.5">{stats.pending_approvals}</div>
                    <div className="text-[10px] text-orange-400 mt-1">
                        Verification required
                    </div>
                </div>
            </div>
        </div>
    );
};
