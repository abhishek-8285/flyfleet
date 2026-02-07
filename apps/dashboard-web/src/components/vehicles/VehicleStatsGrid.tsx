import React from 'react';

export const VehicleStatsGrid: React.FC = () => {
    return (
        <div className="grid grid-cols-4 gap-4">
            <div className="bg-surface-dark rounded-lg p-3 border border-border-color">
                <div className="text-text-muted text-[10px] uppercase font-semibold mb-1">Odometer</div>
                <div className="text-xl font-bold text-white">48,291 <span className="text-xs text-text-muted font-normal">mi</span></div>
            </div>
            <div className="bg-surface-dark rounded-lg p-3 border border-border-color">
                <div className="text-text-muted text-[10px] uppercase font-semibold mb-1">Fuel Level</div>
                <div className="text-xl font-bold text-emerald-400">72%</div>
                <div className="w-full bg-border-color h-1 rounded-full mt-2">
                    <div className="bg-emerald-500 h-1 rounded-full" style={{ width: '72%' }}></div>
                </div>
            </div>
            <div className="bg-surface-dark rounded-lg p-3 border border-border-color">
                <div className="text-text-muted text-[10px] uppercase font-semibold mb-1">Engine Hours</div>
                <div className="text-xl font-bold text-white">2,410 <span className="text-xs text-text-muted font-normal">hrs</span></div>
            </div>
            <div className="bg-surface-dark rounded-lg p-3 border border-border-color">
                <div className="text-text-muted text-[10px] uppercase font-semibold mb-1">Next Service</div>
                <div className="text-xl font-bold text-yellow-400">800 <span className="text-xs text-text-muted font-normal">mi</span></div>
            </div>
        </div>
    );
};
