import React from 'react';
import { useNavigate } from 'react-router-dom';
import { VehicleMapCard } from '../components/vehicles/VehicleMapCard';
import { VehicleSpecsPanel } from '../components/vehicles/VehicleSpecsPanel';
import { VehicleStatsGrid } from '../components/vehicles/VehicleStatsGrid';

const VehicleDetails: React.FC = () => {
    const navigate = useNavigate();

    return (
        <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
            <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate(-1)} className="text-text-muted hover:text-white">
                        <span className="material-symbols-outlined">arrow_back</span>
                    </button>
                    <div>
                        <div className="flex items-center gap-3">
                            <h2 className="text-white text-lg font-bold">Vehicle #301</h2>
                            <span className="bg-emerald-500/10 text-emerald-500 text-xs px-2 py-0.5 rounded border border-emerald-500/20 font-medium">Active</span>
                            <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded border border-primary/20 font-medium flex items-center gap-1">
                                <span className="material-symbols-outlined text-[12px] animate-pulse">leak_add</span>
                                Live Telemetry
                            </span>
                        </div>
                        <p className="text-[11px] text-text-muted mt-0.5">2022 Freightliner Cascadia • VIN: 1F9E...4829</p>
                    </div>
                </div>
                <div className="flex items-center gap-3">
                    <button className="px-3 py-1.5 bg-surface-dark border border-border-color hover:bg-border-color text-text-muted hover:text-white text-xs font-medium rounded flex items-center gap-2 transition-colors">
                        <span className="material-symbols-outlined text-[16px]">edit</span>
                        Edit Vehicle
                    </button>
                    <button className="px-3 py-1.5 bg-red-500/10 border border-red-500/20 hover:bg-red-500/20 text-red-500 text-xs font-medium rounded flex items-center gap-2 transition-colors">
                        <span className="material-symbols-outlined text-[16px]">delete</span>
                        Decommission
                    </button>
                </div>
            </header>
            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                <div className="grid grid-cols-12 gap-6 h-full min-h-[600px]">
                    <div className="col-span-12 lg:col-span-5 flex flex-col gap-6 h-full">
                        <VehicleMapCard />
                    </div>
                    <div className="col-span-12 lg:col-span-7 flex flex-col h-full gap-6">
                        <VehicleStatsGrid />
                        <VehicleSpecsPanel />
                    </div>
                </div>
            </div>
            <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 flex-shrink-0">
                <span>© 2026 Enterprise Fleet Manager. All rights reserved.</span>
                <div className="flex items-center gap-4">
                    <span>Data Source: Telematics Gateway V2</span>
                    <span>Server Time: <span className="font-mono text-white">14:02 UTC</span></span>
                </div>
            </footer>
        </div>
    );
};

export default VehicleDetails;
