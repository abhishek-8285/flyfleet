import React from 'react';

export const MaintenanceStats: React.FC = () => {
    return (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group flex flex-col justify-between">
                <div className="flex justify-between items-start z-10 relative">
                    <div>
                        <p className="text-text-muted text-sm font-medium mb-1">Maintenance Due</p>
                        <h3 className="text-3xl font-bold text-white tracking-tight">14</h3>
                    </div>
                    <div className="p-2 bg-yellow-500/10 rounded-lg text-yellow-500">
                        <span className="material-symbols-outlined">assignment_late</span>
                    </div>
                </div>
                <div className="mt-4 flex items-center text-sm">
                    <span className="text-yellow-500 font-medium flex items-center gap-1">
                        <span className="material-symbols-outlined text-[16px]">priority_high</span>
                        4 High Priority
                    </span>
                    <span className="text-[#64748b] ml-2">Next 7 days</span>
                </div>
            </div>
            <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group flex flex-col justify-between">
                <div className="flex justify-between items-start z-10 relative">
                    <div>
                        <p className="text-text-muted text-sm font-medium mb-1">Critical Issues</p>
                        <h3 className="text-3xl font-bold text-white tracking-tight">3</h3>
                    </div>
                    <div className="p-2 bg-red-500/10 rounded-lg text-red-500">
                        <span className="material-symbols-outlined">car_crash</span>
                    </div>
                </div>
                <div className="mt-4 flex items-center text-sm">
                    <span className="text-red-500 font-medium flex items-center gap-1">
                        <span className="material-symbols-outlined text-[16px]">trending_up</span>
                        +1
                    </span>
                    <span className="text-[#64748b] ml-2">since yesterday</span>
                </div>
            </div>
            <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group flex flex-col justify-between">
                <div className="flex justify-between items-start z-10 relative">
                    <div>
                        <p className="text-text-muted text-sm font-medium mb-1">Healthy Fleet</p>
                        <h3 className="text-3xl font-bold text-white tracking-tight">96%</h3>
                    </div>
                    <div className="p-2 bg-emerald-500/10 rounded-lg text-emerald-500">
                        <span className="material-symbols-outlined">check_circle</span>
                    </div>
                </div>
                <div className="mt-4 flex items-center text-sm">
                    <span className="text-emerald-500 font-medium flex items-center gap-1">
                        <span className="material-symbols-outlined text-[16px]">trending_up</span>
                        2%
                    </span>
                    <span className="text-[#64748b] ml-2">compliance rate</span>
                </div>
            </div>
        </div>
    );
};
