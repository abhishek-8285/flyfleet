import React from 'react';

export const DriverTable: React.FC = () => {
    return (
        <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col h-full shadow-sm overflow-hidden">
            <div className="overflow-auto custom-scrollbar flex-1">
                <table className="w-full text-left text-sm text-text-muted">
                    <thead className="bg-[#1d2936] text-xs uppercase font-medium sticky top-0 z-10 shadow-sm">
                        <tr>
                            <th className="px-6 py-4 cursor-pointer hover:text-white group">
                                <div className="flex items-center gap-1">Name / ID <span className="material-symbols-outlined text-[14px] opacity-0 group-hover:opacity-100">sort</span></div>
                            </th>
                            <th className="px-6 py-4 cursor-pointer hover:text-white group">
                                <div className="flex items-center gap-1">License Status <span className="material-symbols-outlined text-[14px] opacity-0 group-hover:opacity-100">filter_list</span></div>
                            </th>
                            <th className="px-6 py-4">Current Vehicle</th>
                            <th className="px-6 py-4 cursor-pointer hover:text-white group text-center">
                                <div className="flex items-center justify-center gap-1">Safety Score <span className="material-symbols-outlined text-[14px] opacity-0 group-hover:opacity-100">arrow_downward</span></div>
                            </th>
                            <th className="px-6 py-4 cursor-pointer hover:text-white group">
                                <div className="flex items-center gap-1">Last Active <span className="material-symbols-outlined text-[14px] opacity-0 group-hover:opacity-100">sort</span></div>
                            </th>
                            <th className="px-6 py-4 text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-border-color">
                        {/* Row 1 */}
                        <tr className="hover:bg-[#1d2936] transition-colors group">
                            <td className="px-6 py-3">
                                <div className="flex items-center gap-3">
                                    <div className="size-8 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-xs">JD</div>
                                    <div>
                                        <div className="font-medium text-white">John Doe</div>
                                        <div className="text-[10px] text-text-muted">ID: DRV-001</div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-3">
                                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                                    Active
                                </span>
                                <div className="text-[10px] mt-0.5">Exp: 2025-12-01</div>
                            </td>
                            <td className="px-6 py-3">
                                <div className="flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[16px] text-text-muted">local_shipping</span>
                                    <span className="text-white">Ford Transit #402</span>
                                </div>
                            </td>
                            <td className="px-6 py-3 text-center">
                                <div className="inline-flex flex-col items-center">
                                    <span className="text-emerald-500 font-bold">98</span>
                                    <div className="w-16 h-1 bg-border-color rounded-full mt-1 overflow-hidden">
                                        <div className="h-full bg-emerald-500" style={{ width: '98%' }}></div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-3 text-xs">
                                <div className="text-white">Just now</div>
                                <div className="text-[10px] text-text-muted">Active trip</div>
                            </td>
                            <td className="px-6 py-3 text-right">
                                <button className="p-1 text-text-muted hover:text-white hover:bg-border-color rounded transition-colors">
                                    <span className="material-symbols-outlined text-[18px]">more_vert</span>
                                </button>
                            </td>
                        </tr>
                        {/* Row 2 */}
                        <tr className="hover:bg-[#1d2936] transition-colors group">
                            <td className="px-6 py-3">
                                <div className="flex items-center gap-3">
                                    <div className="size-8 rounded-full bg-purple-500/20 flex items-center justify-center text-purple-500 font-bold text-xs">SM</div>
                                    <div>
                                        <div className="font-medium text-white">Sarah Miller</div>
                                        <div className="text-[10px] text-text-muted">ID: DRV-045</div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-3">
                                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-orange-500/10 text-orange-500 border border-orange-500/20">
                                    Expiring Soon
                                </span>
                                <div className="text-[10px] mt-0.5">Exp: 15 days</div>
                            </td>
                            <td className="px-6 py-3">
                                <div className="flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[16px] text-text-muted">directions_car</span>
                                    <span className="text-white">Toyota Prius #104</span>
                                </div>
                            </td>
                            <td className="px-6 py-3 text-center">
                                <div className="inline-flex flex-col items-center">
                                    <span className="text-emerald-500 font-bold">92</span>
                                    <div className="w-16 h-1 bg-border-color rounded-full mt-1 overflow-hidden">
                                        <div className="h-full bg-emerald-500" style={{ width: '92%' }}></div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-3 text-xs">
                                <div className="text-white">2 hours ago</div>
                                <div className="text-[10px] text-text-muted">Shift ended</div>
                            </td>
                            <td className="px-6 py-3 text-right">
                                <button className="p-1 text-text-muted hover:text-white hover:bg-border-color rounded transition-colors">
                                    <span className="material-symbols-outlined text-[18px]">more_vert</span>
                                </button>
                            </td>
                        </tr>
                        {/* Row 3 */}
                        <tr className="hover:bg-[#1d2936] transition-colors group">
                            <td className="px-6 py-3">
                                <div className="flex items-center gap-3">
                                    <div className="size-8 rounded-full bg-blue-500/20 flex items-center justify-center text-blue-500 font-bold text-xs">RJ</div>
                                    <div>
                                        <div className="font-medium text-white">Robert Johnson</div>
                                        <div className="text-[10px] text-text-muted">ID: DRV-089</div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-3">
                                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                                    Active
                                </span>
                                <div className="text-[10px] mt-0.5">Exp: 2024-08-15</div>
                            </td>
                            <td className="px-6 py-3">
                                <div className="flex items-center gap-2">
                                    <span className="material-symbols-outlined text-[16px] text-text-muted">local_shipping</span>
                                    <span className="text-white">Mack Anthem #892</span>
                                </div>
                            </td>
                            <td className="px-6 py-3 text-center">
                                <div className="inline-flex flex-col items-center">
                                    <span className="text-yellow-500 font-bold">78</span>
                                    <div className="w-16 h-1 bg-border-color rounded-full mt-1 overflow-hidden">
                                        <div className="h-full bg-yellow-500" style={{ width: '78%' }}></div>
                                    </div>
                                </div>
                            </td>
                            <td className="px-6 py-3 text-xs">
                                <div className="text-white">10 mins ago</div>
                                <div className="text-[10px] text-text-muted">En route</div>
                            </td>
                            <td className="px-6 py-3 text-right">
                                <button className="p-1 text-text-muted hover:text-white hover:bg-border-color rounded transition-colors">
                                    <span className="material-symbols-outlined text-[18px]">more_vert</span>
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div className="bg-surface-darker px-6 py-3 border-t border-border-color flex items-center justify-between flex-shrink-0">
                <div className="text-xs text-text-muted">
                    Showing <span className="text-white font-medium">1-8</span> of <span className="text-white font-medium">124</span> drivers
                </div>
                <div className="flex gap-2">
                    <button className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors disabled:opacity-50" disabled>Previous</button>
                    <button className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors">Next</button>
                </div>
            </div>
        </div>
    );
};
