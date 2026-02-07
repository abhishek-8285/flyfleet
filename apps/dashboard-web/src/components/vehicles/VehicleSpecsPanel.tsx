import React from 'react';

export const VehicleSpecsPanel: React.FC = () => {
    return (
        <div className="bg-surface-dark rounded-xl border border-border-color flex-1 flex flex-col overflow-hidden">
            <div className="flex border-b border-border-color bg-surface-darker">
                <button className="px-6 py-3 text-sm font-medium text-primary border-b-2 border-primary bg-primary/5">Specs & Overview</button>
                <button className="px-6 py-3 text-sm font-medium text-text-muted hover:text-white hover:bg-border-color/30 transition-colors">Telemetry Feed</button>
                <button className="px-6 py-3 text-sm font-medium text-text-muted hover:text-white hover:bg-border-color/30 transition-colors">Maintenance Logs</button>
                <button className="px-6 py-3 text-sm font-medium text-text-muted hover:text-white hover:bg-border-color/30 transition-colors">Assigned Driver</button>
            </div>
            <div className="p-6 overflow-y-auto custom-scrollbar flex-1">
                <h4 className="text-white font-semibold mb-4 flex items-center gap-2">
                    <span className="material-symbols-outlined text-primary text-lg">info</span>
                    Vehicle Specifications
                </h4>
                <div className="grid grid-cols-2 gap-x-8 gap-y-4 mb-8">
                    <div className="flex justify-between py-2 border-b border-border-color/50">
                        <span className="text-sm text-text-muted">Make / Model</span>
                        <span className="text-sm text-white font-medium">Freightliner Cascadia</span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border-color/50">
                        <span className="text-sm text-text-muted">Year</span>
                        <span className="text-sm text-white font-medium">2022</span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border-color/50">
                        <span className="text-sm text-text-muted">License Plate</span>
                        <span className="text-sm text-white font-medium">CT-8492-LZ</span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border-color/50">
                        <span className="text-sm text-text-muted">Fuel Type</span>
                        <span className="text-sm text-white font-medium">Diesel</span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border-color/50">
                        <span className="text-sm text-text-muted">Gross Weight</span>
                        <span className="text-sm text-white font-medium">80,000 lbs</span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border-color/50">
                        <span className="text-sm text-text-muted">Axle Config</span>
                        <span className="text-sm text-white font-medium">6x4</span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border-color/50">
                        <span className="text-sm text-text-muted">Engine Type</span>
                        <span className="text-sm text-white font-medium">DD15 Gen 5</span>
                    </div>
                    <div className="flex justify-between py-2 border-b border-border-color/50">
                        <span className="text-sm text-text-muted">Transmission</span>
                        <span className="text-sm text-white font-medium">DT12 Automated</span>
                    </div>
                </div>
                <h4 className="text-white font-semibold mb-4 flex items-center gap-2">
                    <span className="material-symbols-outlined text-primary text-lg">assignment_ind</span>
                    Current Assignment
                </h4>
                <div className="bg-surface-darker rounded-lg p-4 border border-border-color flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <div className="h-12 w-12 rounded-full bg-border-color flex items-center justify-center">
                            <span className="material-symbols-outlined text-text-muted text-2xl">person</span>
                        </div>
                        <div>
                            <div className="text-white font-medium">Michael Johnson</div>
                            <div className="text-xs text-text-muted">ID: #DRV-882 • Since Oct 12, 2023</div>
                        </div>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                        <span className="px-2 py-1 bg-emerald-500/10 text-emerald-500 text-xs rounded border border-emerald-500/20">On Duty</span>
                        <button className="text-xs text-primary hover:text-white hover:underline">View Driver Profile</button>
                    </div>
                </div>
                <h4 className="text-white font-semibold mt-8 mb-4 flex items-center gap-2">
                    <span className="material-symbols-outlined text-primary text-lg">warning</span>
                    Recent Alerts
                </h4>
                <div className="space-y-3">
                    <div className="flex items-start gap-3 p-3 bg-red-500/5 border border-red-500/10 rounded-lg">
                        <span className="material-symbols-outlined text-red-500 text-sm mt-0.5">error</span>
                        <div className="flex-1">
                            <div className="flex justify-between">
                                <span className="text-sm font-medium text-red-400">Engine Overheat Warning</span>
                                <span className="text-xs text-text-muted">2 mins ago</span>
                            </div>
                            <p className="text-xs text-text-muted mt-1">Coolant temperature exceeded 220°F for &gt; 30s.</p>
                        </div>
                    </div>
                    <div className="flex items-start gap-3 p-3 bg-surface-darker border border-border-color rounded-lg">
                        <span className="material-symbols-outlined text-orange-400 text-sm mt-0.5">tire_repair</span>
                        <div className="flex-1">
                            <div className="flex justify-between">
                                <span className="text-sm font-medium text-white">Tire Pressure Low</span>
                                <span className="text-xs text-text-muted">4 hours ago</span>
                            </div>
                            <p className="text-xs text-text-muted mt-1">Right Rear Inner tire pressure at 85psi (Target: 100psi).</p>
                        </div>
                    </div>
                </div>
            </div>
            <div className="p-4 border-t border-border-color bg-surface-darker flex justify-end gap-3">
                <button className="px-4 py-2 text-xs font-medium text-white bg-primary hover:bg-blue-600 rounded transition-colors shadow-lg shadow-primary/20">
                    Download Full Report
                </button>
            </div>
        </div>
    );
};
