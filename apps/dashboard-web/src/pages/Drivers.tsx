import React from 'react';
import { DriverFilters } from '../components/drivers/DriverFilters';
import { DriverTable } from '../components/drivers/DriverTable';

const Drivers: React.FC = () => {
    return (
        <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
            <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 flex-shrink-0 z-10">
                <div className="flex flex-wrap items-center gap-3">
                    <div className="relative">
                        <select className="appearance-none bg-surface-dark border border-border-color text-white text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors">
                            <option>Active Drivers</option>
                            <option>Off-Duty</option>
                            <option>Maintenance</option>
                        </select>
                        <span className="material-symbols-outlined absolute right-2 top-1.5 pointer-events-none text-text-muted text-sm">badge</span>
                    </div>
                </div>
                <div className="flex items-center gap-3 ml-auto">
                    <button className="bg-primary hover:bg-primary/90 text-white text-xs font-medium py-1.5 px-3 rounded flex items-center gap-2 shadow-[0_0_10px_rgba(19,127,236,0.3)] transition-all">
                        <span className="material-symbols-outlined text-[16px]">person_add</span>
                        Add Driver
                    </button>
                </div>
            </div>

            <DriverFilters />
            
            <div className="flex-1 overflow-hidden p-6 bg-background-dark">
                <DriverTable />
            </div>
            
            <footer className="mt-auto text-center text-xs text-[#64748b] pb-2 flex flex-col items-center gap-1 bg-background-dark">
                <span>© 2026 FleetCommand. All rights reserved.</span>
                <span className="text-[10px]">v2.5.0 | Server Time: <span className="font-mono">14:02 UTC</span></span>
            </footer>
        </div>
    );
};

export default Drivers;
