import React from 'react';

export const DriverFilters: React.FC = () => {
    return (
        <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 flex-shrink-0 z-10">
            <div className="flex flex-wrap items-center gap-3 w-full md:w-auto">
                <div className="relative">
                    <select className="appearance-none bg-surface-dark border border-border-color text-white text-sm rounded pl-3 pr-8 py-2 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors min-w-[160px]">
                        <option>All License Status</option>
                        <option>Active</option>
                        <option>Expiring Soon (30 days)</option>
                        <option>Expired</option>
                        <option>Suspended</option>
                    </select>
                    <span className="material-symbols-outlined absolute right-2 top-2.5 pointer-events-none text-text-muted text-sm">filter_list</span>
                </div>
                <div className="relative">
                    <select className="appearance-none bg-surface-dark border border-border-color text-white text-sm rounded pl-3 pr-8 py-2 focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer hover:bg-border-color transition-colors min-w-[160px]">
                        <option>All Safety Scores</option>
                        <option>Excellent (90-100)</option>
                        <option>Good (80-89)</option>
                        <option>Needs Review (&lt; 80)</option>
                        <option>Critical (&lt; 60)</option>
                    </select>
                    <span className="material-symbols-outlined absolute right-2 top-2.5 pointer-events-none text-text-muted text-sm">health_and_safety</span>
                </div>
            </div>
            <div className="flex items-center gap-3 ml-auto w-full md:w-auto justify-end">
                <button className="hidden md:flex items-center gap-2 px-4 py-2 bg-surface-dark border border-border-color text-text-muted text-sm font-medium rounded hover:text-white hover:bg-border-color transition-colors">
                    <span className="material-symbols-outlined text-[18px]">download</span>
                    Download Report
                </button>
                <button className="flex items-center gap-2 px-4 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-blue-600 transition-colors shadow-lg shadow-blue-500/20">
                    <span className="material-symbols-outlined text-[18px]">person_add</span>
                    Add Driver
                </button>
            </div>
        </div>
    );
};
