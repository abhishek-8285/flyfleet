import React from 'react';
import { MaintenanceFilters } from '../components/maintenance/MaintenanceFilters';
import { MaintenanceStats } from '../components/maintenance/MaintenanceStats';
import { ServiceLogTable } from '../components/maintenance/ServiceLogTable';

const Maintenance: React.FC = () => {
    return (
        <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
            
            <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 flex-shrink-0 z-10">
                <div className="flex flex-wrap items-center gap-3">
                    <div className="relative">
                        <select className="appearance-none bg-surface-dark border border-border-color text-white text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors">
                            <option>This Week</option>
                            <option>Next 7 Days</option>
                            <option>Next 30 Days</option>
                            <option>Overdue</option>
                        </select>
                        <span className="material-symbols-outlined absolute right-2 top-1.5 pointer-events-none text-text-muted text-sm">calendar_month</span>
                    </div>
                    <div className="relative">
                        <select className="appearance-none bg-surface-dark border border-border-color text-white text-xs rounded pl-3 pr-8 py-1.5 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors">
                            <option>All Divisions</option>
                            <option>North America</option>
                            <option>Europe</option>
                        </select>
                        <span className="material-symbols-outlined absolute right-2 top-1.5 pointer-events-none text-text-muted text-sm">domain</span>
                    </div>
                </div>
                <div className="flex items-center gap-3 ml-auto">
                    <div className="text-xs text-text-muted">
                        Last sync: <span className="text-white">Just now</span>
                    </div>
                    <div className="h-4 w-px bg-border-color"></div>
                    <button className="bg-primary hover:bg-primary/90 text-white text-xs font-medium py-1.5 px-3 rounded flex items-center gap-2 shadow-[0_0_10px_rgba(19,127,236,0.3)] transition-all">
                        <span className="material-symbols-outlined text-[16px]">add</span>
                        Log New Service
                    </button>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                <div className="max-w-[1600px] mx-auto pb-6 space-y-6">
                    <MaintenanceStats />
                    <div className="grid grid-cols-12 gap-6">
                        <div className="col-span-12 lg:col-span-3">
                            <MaintenanceFilters />
                        </div>
                        <div className="col-span-12 lg:col-span-9">
                            <ServiceLogTable />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Maintenance;
