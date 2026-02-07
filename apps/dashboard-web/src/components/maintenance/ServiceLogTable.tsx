import React from 'react';

export const ServiceLogTable: React.FC = () => {
    return (
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col h-full">
            <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
                <h3 className="text-white font-semibold">Service Log Entries</h3>
                <div className="flex gap-3">
                    <button className="text-xs font-medium text-text-muted hover:text-white flex items-center gap-1 bg-border-color/30 px-3 py-1.5 rounded transition-colors border border-border-color/50">
                        <span className="material-symbols-outlined text-[16px]">download</span>
                        Export CSV
                    </button>
                </div>
            </div>
            <div className="overflow-x-auto custom-scrollbar flex-1">
                <table className="w-full text-left text-sm text-text-muted">
                    <thead className="bg-[#1d2936] text-xs uppercase font-medium sticky top-0 z-10 shadow-sm">
                        <tr>
                            <th className="px-6 py-4 cursor-pointer hover:text-white group">
                                <div className="flex items-center gap-1">Vehicle <span className="material-symbols-outlined text-[14px] text-text-muted/50">sort</span></div>
                            </th>
                            <th className="px-6 py-4">Health Status</th>
                            <th className="px-6 py-4 cursor-pointer hover:text-white group">
                                <div className="flex items-center gap-1">Upcoming Service <span className="material-symbols-outlined text-[14px] text-text-muted/50">arrow_upward</span></div>
                            </th>
                            <th className="px-6 py-4">Odometer</th>
                            <th className="px-6 py-4">Open Tickets</th>
                            <th className="px-6 py-4 text-right">Action</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-border-color">
                        {/* Row 1 */}
                        <tr className="hover:bg-[#1d2936] transition-colors group">
                            <td className="px-6 py-4">
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-8 rounded bg-border-color flex items-center justify-center text-text-muted">
                                        <span className="material-symbols-outlined text-[18px]">local_shipping</span>
                                    </div>
                                    <div>
                                        <div className="text-white font-medium">Ford F-150 #301</div>
                                        <div className="text-xs text-text-muted">VIN: ...8892</div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-4">
                                <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                                    <span className="w-1.5 h-1.5 rounded-full bg-red-500 mr-1.5 animate-pulse"></span>
                                    Critical
                                </span>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white font-medium">Oct 25, 2023</div>
                                <div className="text-xs text-red-400">Overdue (2 days)</div>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white">45,202 mi</div>
                            </td>
                            <td className="px-6 py-4">
                                <a className="text-primary hover:underline hover:text-white flex items-center gap-1" href="#">
                                    <span className="material-symbols-outlined text-[16px]">confirmation_number</span>
                                    2 Tickets
                                </a>
                                <div className="text-[10px] text-text-muted mt-0.5">Engine Overheat, Brake Pad</div>
                            </td>
                            <td className="px-6 py-4 text-right">
                                <button className="bg-primary/10 hover:bg-primary text-primary hover:text-white px-3 py-1.5 rounded text-xs font-medium transition-colors border border-primary/20">
                                    Log Service
                                </button>
                            </td>
                        </tr>
                        {/* Row 2 */}
                        <tr className="hover:bg-[#1d2936] transition-colors group">
                            <td className="px-6 py-4">
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-8 rounded bg-border-color flex items-center justify-center text-text-muted">
                                        <span className="material-symbols-outlined text-[18px]">directions_bus</span>
                                    </div>
                                    <div>
                                        <div className="text-white font-medium">Sprinter Van #442</div>
                                        <div className="text-xs text-text-muted">VIN: ...1023</div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-4">
                                <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-yellow-500/10 text-yellow-500 border border-yellow-500/20">
                                    <span className="w-1.5 h-1.5 rounded-full bg-yellow-500 mr-1.5"></span>
                                    Maintenance Due
                                </span>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white font-medium">Oct 28, 2023</div>
                                <div className="text-xs text-yellow-500">In 3 days</div>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white">28,400 mi</div>
                            </td>
                            <td className="px-6 py-4">
                                <a className="text-primary hover:underline hover:text-white flex items-center gap-1" href="#">
                                    <span className="material-symbols-outlined text-[16px]">confirmation_number</span>
                                    1 Ticket
                                </a>
                                <div className="text-[10px] text-text-muted mt-0.5">Scheduled Oil Change</div>
                            </td>
                            <td className="px-6 py-4 text-right">
                                <button className="bg-border-color hover:bg-white text-text-muted hover:text-surface-darker px-3 py-1.5 rounded text-xs font-medium transition-colors">
                                    View Details
                                </button>
                            </td>
                        </tr>
                        {/* Row 3 */}
                        <tr className="hover:bg-[#1d2936] transition-colors group">
                            <td className="px-6 py-4">
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-8 rounded bg-border-color flex items-center justify-center text-text-muted">
                                        <span className="material-symbols-outlined text-[18px]">local_shipping</span>
                                    </div>
                                    <div>
                                        <div className="text-white font-medium">Mack Anthem #105</div>
                                        <div className="text-xs text-text-muted">VIN: ...5512</div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-4">
                                <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1.5"></span>
                                    Healthy
                                </span>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white font-medium">Nov 15, 2023</div>
                                <div className="text-xs text-text-muted">In 2 weeks</div>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white">12,105 mi</div>
                            </td>
                            <td className="px-6 py-4">
                                <span className="text-text-muted/50 text-xs italic">No open tickets</span>
                            </td>
                            <td className="px-6 py-4 text-right">
                                <button className="text-text-muted hover:text-white p-1 rounded hover:bg-border-color transition-colors">
                                    <span className="material-symbols-outlined">more_horiz</span>
                                </button>
                            </td>
                        </tr>
                         {/* Row 4 */}
                         <tr className="hover:bg-[#1d2936] transition-colors group">
                            <td className="px-6 py-4">
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-8 rounded bg-border-color flex items-center justify-center text-text-muted">
                                        <span className="material-symbols-outlined text-[18px]">directions_car</span>
                                    </div>
                                    <div>
                                        <div className="text-white font-medium">Sedan Fleet #004</div>
                                        <div className="text-xs text-text-muted">VIN: ...9911</div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-4">
                                <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1.5"></span>
                                    Healthy
                                </span>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white font-medium">Dec 01, 2023</div>
                                <div className="text-xs text-text-muted">In 1 month</div>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white">5,320 mi</div>
                            </td>
                            <td className="px-6 py-4">
                                <span className="text-text-muted/50 text-xs italic">No open tickets</span>
                            </td>
                            <td className="px-6 py-4 text-right">
                                <button className="text-text-muted hover:text-white p-1 rounded hover:bg-border-color transition-colors">
                                    <span className="material-symbols-outlined">more_horiz</span>
                                </button>
                            </td>
                        </tr>
                        {/* Row 5 */}
                        <tr className="hover:bg-[#1d2936] transition-colors group">
                            <td className="px-6 py-4">
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-8 rounded bg-border-color flex items-center justify-center text-text-muted">
                                        <span className="material-symbols-outlined text-[18px]">local_shipping</span>
                                    </div>
                                    <div>
                                        <div className="text-white font-medium">Freightliner #220</div>
                                        <div className="text-xs text-text-muted">VIN: ...3341</div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-4">
                                <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-yellow-500/10 text-yellow-500 border border-yellow-500/20">
                                    <span className="w-1.5 h-1.5 rounded-full bg-yellow-500 mr-1.5"></span>
                                    Maintenance Due
                                </span>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white font-medium">Oct 30, 2023</div>
                                <div className="text-xs text-yellow-500">In 5 days</div>
                            </td>
                            <td className="px-6 py-4">
                                <div className="text-white">88,100 mi</div>
                            </td>
                            <td className="px-6 py-4">
                                <a className="text-primary hover:underline hover:text-white flex items-center gap-1" href="#">
                                    <span className="material-symbols-outlined text-[16px]">confirmation_number</span>
                                    1 Ticket
                                </a>
                                <div className="text-[10px] text-text-muted mt-0.5">Tire Rotation</div>
                            </td>
                            <td className="px-6 py-4 text-right">
                                <button className="bg-border-color hover:bg-white text-text-muted hover:text-surface-darker px-3 py-1.5 rounded text-xs font-medium transition-colors">
                                    View Details
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div className="bg-surface-darker px-4 py-3 border-t border-border-color flex items-center justify-between">
                <span className="text-xs text-text-muted">Showing <span className="text-white font-medium">1-5</span> of <span className="text-white font-medium">124</span> results</span>
                <div className="flex gap-2">
                    <button className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors">Previous</button>
                    <button className="px-3 py-1 text-xs font-medium text-white bg-primary rounded shadow-sm">1</button>
                    <button className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors">2</button>
                    <button className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors">3</button>
                    <button className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors">Next</button>
                </div>
            </div>
        </div>
    );
};
